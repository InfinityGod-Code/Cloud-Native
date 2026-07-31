package transport

import (
	"eventra/repository"
	"net/http"
)

type eventServiceHandler struct {
	dbHandler repository.DatabaseHandler
}

func EventHandler(databaseHandler repository.DatabaseHandler) *eventServiceHandler {
	return &eventServiceHandler{
		dbHandler: databaseHandler,
	}
}

// Adding methods to the struct type eventServiceHandler

func (et *eventServiceHandler) addEventHandler(w http.ResponseWriter, r *http.Request) {

}

func (et *eventServiceHandler) allEventHandler(w http.ResponseWriter, r *http.Request) {

}

func (et *eventServiceHandler) singleEventHandler(w http.ResponseWriter, r *http.Request) {

}
