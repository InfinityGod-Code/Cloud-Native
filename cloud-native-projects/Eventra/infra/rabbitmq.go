package infra

import (
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// RabbitMQ is a thin wrapper around the amqp091-go client.
//
// It bundles the two things you interact with most in AMQP:
//   - Connection: a long-lived TCP connection to the RabbitMQ server.
//   - Channel: a lightweight, multiplexed "virtual connection" over the
//     Connection. All operations (declare/publish/consume) go through a
//     Channel, not the Connection directly.
type RabbitMQ struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	url     string
}

// NewRabbitMQ establishes a connection and opens a channel.
//
// The URL has the form: amqp://user:password@host:port/vhost
// Example (native):  amqp://guest:guest@localhost:5672/
// Example (Docker):  amqp://guest:guest@rabbitmq:5672/
//
// A retry loop with backoff is used so the application can start up even if
// RabbitMQ is not ready yet (e.g. when both are brought up together with
// docker-compose). After 5 failed attempts we give up and return an error.
func NewRabbitMQ(url string) (*RabbitMQ, error) {
	var (
		conn *amqp.Connection
		err  error
	)

	// amqp.Dial opens the underlying TCP connection and performs the AMQP
	// handshake. If the broker is unreachable it returns an error.
	for attempt := 1; attempt <= 5; attempt++ {
		conn, err = amqp.Dial(url)
		if err == nil {
			break
		}
		log.Printf("rabbitmq not ready (attempt %d/5), retrying in %ds: %v", attempt, attempt, err)
		time.Sleep(time.Duration(attempt) * time.Second)
	}
	if err != nil {
		return nil, fmt.Errorf("could not connect to rabbitmq: %w", err)
	}

	// Open a channel on top of the connection. A channel is where all
	// AMQP operations (queue declarations, publish, consume) actually run.
	channel, err := conn.Channel()
	if err != nil {
		// We opened the connection successfully but failed to get a channel,
		// so close the connection to avoid leaking it.
		_ = conn.Close()
		return nil, fmt.Errorf("could not open rabbitmq channel: %w", err)
	}

	return &RabbitMQ{
		conn:    conn,
		channel: channel,
		url:     url,
	}, nil
}

// Close releases the channel first and then the connection.
// Closing order matters: the channel must be closed before the connection
// it belongs to, otherwise RabbitMQ logs an unexpected-channel-close error.
func (r *RabbitMQ) Close() error {
	if r.channel != nil {
		_ = r.channel.Close()
	}
	if r.conn != nil {
		return r.conn.Close()
	}
	return nil
}

// URL returns the AMQP connection string used by this client (useful for logs).
func (r *RabbitMQ) URL() string {
	return r.url
}

// DeclareQueue ensures a queue exists on the broker.
//
// Queue declaration is idempotent: calling it repeatedly is safe and simply
// confirms the queue with the given settings.
//
// Parameters (QueueDeclare):
//   - name:  the queue name
//   - durable=true:  the queue survives a broker restart
//   - autoDelete=false: keep the queue even when no consumers are attached
//   - exclusive=false:  allow other connections to use the queue
//   - noWait=false:     wait for the broker to confirm the declaration
//   - args=nil:         no extra arguments (e.g. TTL, dead-letter config)
func (r *RabbitMQ) DeclareQueue(name string) (amqp.Queue, error) {
	return r.channel.QueueDeclare(name, true, false, false, false, nil)
}

// DeclareExchange ensures an exchange exists on the broker.
//
// An exchange is the "post office" that receives messages from publishers and
// routes them to queues via bindings. The publisher never sends to a queue
// directly; it always sends to an exchange using a routing key.
//
// kind is the exchange type:
//   - "direct": routes to queues whose binding key exactly matches the routing key
//   - "topic":  routes to queues whose binding key matches a pattern (e.g. "event.*")
//   - "fanout": broadcasts to every bound queue, ignoring the routing key
//   - "headers": routes based on message headers instead of routing keys
//
// Like QueueDeclare, ExchangeDeclare is idempotent — declaring the same
// exchange repeatedly is safe. It is durable (survives broker restarts).
func (r *RabbitMQ) DeclareExchange(name ExchangeName, kind string) error {
	// ExchangeDeclare(name, kind, durable, autoDelete, internal, noWait, args)
	return r.channel.ExchangeDeclare(string(name), kind, true, false, false, false, nil)
}

// BindQueue creates a binding that connects a queue to an exchange using a
// routing key. This is what tells the exchange "deliver messages whose routing
// key matches <routingKey> to <queue>".
//
// A binding is a separate, declarative step — it is NOT done implicitly when
// you publish or consume. Both the producer and consumer can (re)declare it
// safely at startup because QueueBind is idempotent.
func (r *RabbitMQ) BindQueue(queue QueueName, routingKey RoutingKeyName, exchange ExchangeName) error {
	// QueueBind(queue, routingKey, exchange, noWait, args)
	return r.channel.QueueBind(string(queue), string(routingKey), string(exchange), false, nil)
}

// SetupEventTopology declares the full event plumbing used by this project:
//
//	[ event exchange (topic) ]  --event.created-->  [ event-queue ]
//
// It is idempotent, so it can be called from BOTH sides:
//   - the publisher (eventservice) calls it before publishing, so the
//     exchange/queue/binding exist and messages are never dropped;
//   - the consumer (bookingservice) calls it before consuming, so it is
//     guaranteed to receive messages even if it starts before the publisher.
//
// This is the "consumer declares its own topology" best practice from notes.md.
func (r *RabbitMQ) SetupEventTopology() error {
	// 1. Create the "event" topic exchange (publishers send here).
	if err := r.DeclareExchange(Event, "topic"); err != nil {
		return fmt.Errorf("could not declare exchange %q: %w", Event, err)
	}

	// 2. Create the "event-queue" (messages accumulate here until consumed).
	if _, err := r.DeclareQueue(string(EventQueue)); err != nil {
		return fmt.Errorf("could not declare queue %q: %w", EventQueue, err)
	}

	// 3. Bind the queue to the exchange via the "event.created" routing key.
	if err := r.BindQueue(EventQueue, EventCreatedRoutingKey, Event); err != nil {
		return fmt.Errorf("could not bind queue %q to exchange %q: %w", EventQueue, Event, err)
	}
	return nil
}

// Publish sends a message to a queue.
//
// Publishing to the default exchange ("") with the queue name as the routing
// key is the simplest way to deliver a message directly to a named queue.
// We declare the queue first so it definitely exists before publishing.
func (r *RabbitMQ) Publish(queue, body string) error {
	if _, err := r.DeclareQueue(queue); err != nil {
		return fmt.Errorf("could not declare queue %q: %w", queue, err)
	}

	// Publish args:
	//   exchange  ""     -> the default (nameless) exchange
	//   key       queue  -> routing key, matches the queue name
	//   mandatory false  -> don't return the message if no queue matches
	//   immediate false  -> deprecated, always false
	err := r.channel.Publish("", queue, false, false, amqp.Publishing{
		ContentType: "text/plain",
		Body:        []byte(body),
	})
	if err != nil {
		return fmt.Errorf("could not publish to queue %q: %w", queue, err)
	}
	return nil
}

// PublishOnExchange sends a message to a NAMED exchange using a routing key.
//
// This is the recommended way to publish when you have a real exchange with
// bindings. The broker uses the routing key to decide which bound queues
// receive the message. For the "event" topic exchange, publishing with the
// "event.created" routing key delivers to any queue bound on "event.created"
// (or a matching wildcard pattern).
func (r *RabbitMQ) PublishOnExchange(exchange ExchangeName, routingKey RoutingKeyName, body string) error {
	// Publish(exchange, routingKey, mandatory, immediate, msg).
	//   mandatory=false -> if no queue matches the routing key, the broker
	//   silently drops the message instead of returning it to us.
	err := r.channel.Publish(string(exchange), string(routingKey), false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        []byte(body),
	})
	if err != nil {
		return fmt.Errorf("could not publish to exchange %q with routing key %q: %w", exchange, routingKey, err)
	}
	return nil
}

// Consume starts a consumer on a queue and returns a channel of deliveries.
//
// The returned channel yields one amqp.Delivery per message. Messages are
// delivered as they arrive; the caller iterates over the channel and calls
// d.Ack()/d.Nack() as needed.
//
// autoAck=true means the broker considers the message delivered as soon as it
// is sent to us, so the caller does not need to call Ack() manually.
func (r *RabbitMQ) Consume(queue string) (<-chan amqp.Delivery, error) {
	if _, err := r.DeclareQueue(queue); err != nil {
		return nil, fmt.Errorf("could not declare queue %q: %w", queue, err)
	}

	return r.channel.Consume(queue, "", true, false, false, false, nil)
}
