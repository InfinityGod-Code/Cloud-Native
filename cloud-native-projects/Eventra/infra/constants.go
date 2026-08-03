package infra

// creating enums for the exchange name, Queue name and the routing key name
type ExchangeName string
type QueueName string
type RoutingKeyName string

const (
	Event ExchangeName = "event"
)

const (
	EventQueue QueueName = "event-queue"
)

/*
Routing Key is basically binds the exchange to the queue. It is used to route the message from the exchange to the queue.
*/
const (
	EventCreatedRoutingKey RoutingKeyName = "event.created"
)
