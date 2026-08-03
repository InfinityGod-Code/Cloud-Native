package transport

import (
	"eventservice/repository"
	"infra"
	"net/http"

	"github.com/gorilla/mux"
)

func ServeAPI(endpoint string, databaseHandler repository.DatabaseHandler, rabbit *infra.RabbitMQ) error {
	handler := NewEventHandler(databaseHandler, rabbit)
	r := mux.NewRouter()
	eventsrouter := r.PathPrefix("/events").Subrouter()
	eventsrouter.Methods("GET").Path("/{SearchCriteria}/{search}").HandlerFunc(handler.singleEventHandler)
	eventsrouter.Methods("GET").Path("").HandlerFunc(handler.allEventHandler)
	eventsrouter.Methods("POST").Path("").HandlerFunc(handler.addEventHandler)
	return http.ListenAndServe(endpoint, r)
}
