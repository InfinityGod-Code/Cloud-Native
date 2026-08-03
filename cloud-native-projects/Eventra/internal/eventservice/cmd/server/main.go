package main

import (
	"flag"
	"log"

	"eventservice/configuration"
	"eventservice/repository"
	"eventservice/repository/mongo"
	"eventservice/transport"
	"infra"
)

func main() {
	configPath := flag.String("config", "internal/eventservice/configuration/config.json", "Configuration file path")
	flag.Parse()

	conf, err := configuration.ExtractConfiguration(*configPath)
	if err != nil {
		log.Printf("Unable to load configuration from %s. Continuing with default values: %v", *configPath, err)
	}

	repository.RegisterMongoDBLayer(mongo.NewMongoDBLayer)

	dbHandler, err := repository.NewPersistenceLayer(conf.Databasetype, conf.DBConnection)
	if err != nil {
		log.Fatal(err)
	}

	// Establish the RabbitMQ connection via the infra layer.
	// NewRabbitMQ retries internally, so this tolerates RabbitMQ starting
	// slightly after the app (e.g. with docker-compose).
	rabbit, err := infra.NewRabbitMQ(conf.RabbitMQURL)
	if err != nil {
		log.Fatalf("Unable to connect to RabbitMQ: %v", err)
	}
	defer rabbit.Close()
	log.Printf("Connected to RabbitMQ at %s", rabbit.URL())

	// Declare the event exchange, queue, and binding so that messages
	// published later on POST /events always have a destination.
	if err := rabbit.SetupEventTopology(); err != nil {
		log.Fatalf("Unable to set up RabbitMQ event topology: %v", err)
	}
	log.Printf("Event topology ready (exchange=%s, queue=%s, routing_key=%s)", infra.Event, infra.EventQueue, infra.EventCreatedRoutingKey)

	log.Fatal(transport.ServeAPI(conf.RestfulEndpoint, dbHandler, rabbit))
}
