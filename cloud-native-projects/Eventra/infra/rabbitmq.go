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
