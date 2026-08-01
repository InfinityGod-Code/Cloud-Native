package repository

import "errors"

type DBTYPE string

const (
	MONGODB  DBTYPE = "mongodb"
	DYNAMODB DBTYPE = "dynamodb"
)

var mongoDBLayerFactory func(string) (DatabaseHandler, error)

func RegisterMongoDBLayer(factory func(string) (DatabaseHandler, error)) {
	mongoDBLayerFactory = factory
}

func NewPersistenceLayer(options DBTYPE, connection string) (DatabaseHandler, error) {
	switch options {
	case MONGODB:
		if mongoDBLayerFactory == nil {
			return nil, errors.New("mongo layer not registered")
		}
		return mongoDBLayerFactory(connection)
	}
	return nil, errors.New("unsupported database type")
}
