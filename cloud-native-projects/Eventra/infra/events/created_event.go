package events

import "time"

// EventCreatedEvent is the message published when a new event is created.
// It is the "contract" exchanged between services over RabbitMQ: the
// eventservice produces it (on POST /events) and the bookingservice consumes
// it. Keeping it in the shared infra module means both services serialize
// the same JSON shape without importing each other's code.
//
// The JSON tags define the wire format. NOTE: these differ from the internal
// domain.Event model (which uses int64 epoch seconds); this DTO uses
// time.Time so consumers get human-friendly timestamps.
type EventCreatedEvent struct {
	ID         string    `json:"id"`         // MongoDB ObjectId as hex string
	Name       string    `json:"name"`       // Event name
	LocationID string    `json:"location_id"` // MongoDB ObjectId of the venue
	Start      time.Time `json:"start_time"` // Event start time (RFC3339)
	End        time.Time `json:"end_time"`   // Event end time (RFC3339)
}
