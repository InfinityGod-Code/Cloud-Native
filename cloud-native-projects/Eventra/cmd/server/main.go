package main

import (
	"flag"
	"log"

	"eventra/internal/configuration"
	"eventra/internal/repository"
	"eventra/internal/repository/mongo"
	"eventra/internal/transport"
)

func main() {
	configPath := flag.String("config", "internal/configuration/config.json", "Configuration file path")
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

	log.Fatal(transport.ServeAPI(conf.RestfulEndpoint, dbHandler))
}
