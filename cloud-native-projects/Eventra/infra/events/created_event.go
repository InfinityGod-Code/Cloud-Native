package events

import "time"

type EventCreatedEvent struct {
	ID         string    `json:"id"`
	Name       string    `json:"id"`
	LocationID string    `json:"id"`
	Start      time.Time `json:"start_time"`
	End        time.Time `json:"end_time"`
}
