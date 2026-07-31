package main

import (
	"log"

	"eventra/internal/transport"
	"eventra/repository"
	"eventra/repository/mongo"
)

func main() {
	repository.RegisterMongoDBLayer(mongo.NewMongoDBLayer)

	dbHandler, err := repository.NewPersistenceLayer(repository.MONGODB, "mongodb://localhost:27017")
	if err != nil {
		log.Fatal(err)
	}

	log.Fatal(transport.ServeAPI(":8080", dbHandler))
}
