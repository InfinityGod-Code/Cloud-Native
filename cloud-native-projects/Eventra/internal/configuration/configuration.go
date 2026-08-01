package configuration

import (
	"encoding/json"
	"eventra/internal/repository"
	"log"
	"os"
)

var (
	DBTypeDefault       = repository.DBTYPE("mongodb")
	DBConnectionDefault = "mongodb://127.0.0.1"
	RestfulEPDefault    = "localhost:8181"
	RestfulTLSEPDefault = "localhost:9191"
)

type ServiceConfig struct {
	Databasetype      repository.DBTYPE `json:"databasetype"`
	DBConnection      string            `json:"dbconnection"`
	RestfulEndpoint   string            `json:"restfulapi_endpoint"`
	RestfulTLSEndPint string            `json:"restfulapi-tlsendpoint"`
}

func ExtractConfiguration(filename string) (ServiceConfig, error) {
	conf := ServiceConfig{
		DBTypeDefault,
		DBConnectionDefault,
		RestfulEPDefault,
		RestfulTLSEPDefault,
	}

	file, err := os.Open(filename)
	if err != nil {
		log.Printf("Configuration file not found. Continuing with default values.")
		return conf, err
	}
	defer file.Close()

	err = json.NewDecoder(file).Decode(&conf)
	return conf, err
}
