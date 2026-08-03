package main

import (
	"encoding/json"
	"log"
	"os"

	"infra"
	"infra/events"
)

func main() {
	// The consumer listens on the RabbitMQ queue bound to the event exchange.
	//
	// Flow:  eventservice publishes to [event exchange] --event.created-->
	//        [event-queue] --delivered--> this listener.
	//
	// The URL comes from the RABBITMQ_URL env var (so docker-compose / k8s can
	// inject the right broker address) and falls back to a local default.
	url := os.Getenv("RABBITMQ_URL")
	if url == "" {
		url = "amqp://guest:guest@localhost:5672/"
	}

	// Establish a long-lived connection. NewRabbitMQ retries internally at
	// startup, so this tolerates RabbitMQ coming up slightly after us.
	rabbit, err := infra.NewRabbitMQ(url)
	if err != nil {
		log.Fatalf("unable to connect to rabbitmq: %v", err)
	}
	defer rabbit.Close()
	log.Printf("connected to rabbitmq at %s", rabbit.URL())

	// Declare the same topology as the publisher (idempotent, safe to repeat).
	// This guarantees the exchange, queue, and binding exist before we consume,
	// even if the bookingservice starts before the eventservice.
	if err := rabbit.SetupEventTopology(); err != nil {
		log.Fatalf("unable to set up event topology: %v", err)
	}

	// Start consuming from the queue. Consume returns a channel of deliveries;
	// the for loop below blocks until a message arrives. The loop only ends
	// when RabbitMQ closes the channel (e.g. on a network error), at which
	// point the program exits and the orchestrator (docker/k8s) restarts it.
	deliveries, err := rabbit.Consume(string(infra.EventQueue))
	if err != nil {
		log.Fatalf("unable to consume from queue %q: %v", infra.EventQueue, err)
	}

	log.Printf("listening for events on queue %q (routing key %q)", infra.EventQueue, infra.EventCreatedRoutingKey)
	for d := range deliveries {
		// Deserialize into the shared DTO so we work with a typed value
		// instead of raw bytes.
		var evt events.EventCreatedEvent
		if err := json.Unmarshal(d.Body, &evt); err != nil {
			log.Printf("could not decode event message %q: %v", d.Body, err)
			continue
		}

		// In a real service this is where you'd process the event — e.g. open
		// the booking window for the new event, send notifications, etc.
		log.Printf("received event.created: id=%s name=%q location_id=%s start=%s end=%s",
			evt.ID, evt.Name, evt.LocationID, evt.Start, evt.End)
	}
}
