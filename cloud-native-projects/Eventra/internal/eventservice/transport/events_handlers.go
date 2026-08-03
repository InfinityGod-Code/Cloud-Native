package transport

import (
	"encoding/json"
	"eventservice/domain"
	"eventservice/repository"
	"infra"
	"infra/events"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"gopkg.in/mgo.v2/bson"
)

type eventServiceHandler struct {
	dbHandler repository.DatabaseHandler
	rabbit    *infra.RabbitMQ
}

func NewEventHandler(databaseHandler repository.DatabaseHandler, rabbit *infra.RabbitMQ) *eventServiceHandler {
	return &eventServiceHandler{
		dbHandler: databaseHandler,
		rabbit:    rabbit,
	}
}

// Adding methods to the struct type eventServiceHandler

func (et *eventServiceHandler) addEventHandler(w http.ResponseWriter, r *http.Request) {
	event := domain.Event{}
	err := json.NewDecoder(r.Body).Decode(&event)
	if err != nil {
		WriteErrorResponse(w, http.StatusBadRequest, "bad request: unable to decode event")
		return
	}

	if event.Name == "" {
		WriteErrorResponse(w, http.StatusBadRequest, "bad request: event name is required")
		return
	}

	id, err := et.dbHandler.AddEvent(event)
	if err != nil {
		WriteErrorResponse(w, http.StatusInternalServerError, "internal server error: unable to add event")
		return
	}
	event.ID = bson.ObjectId(id)

	// Publish an event.created message so other services (e.g. bookingservice)
	// can react to the new event.
	//
	// We translate the internal domain.Event into the shared events.EventCreatedEvent
	// DTO (defined in infra) so the wire format stays stable and decoupled from the
	// storage model. This is the "producer publishes to a named exchange with a
	// routing key" pattern.
	if et.rabbit != nil {
		createdEvent := events.EventCreatedEvent{
			ID:         event.ID.Hex(),
			Name:       event.Name,
			LocationID: event.Location.ID.Hex(),
			Start:      time.Unix(event.StartDate, 0),
			End:        time.Unix(event.EndDate, 0),
		}
		payload, _ := json.Marshal(createdEvent)
		if err := et.rabbit.PublishOnExchange(infra.Event, infra.EventCreatedRoutingKey, string(payload)); err != nil {
			log.Printf("failed to publish event.created to rabbitmq: %v", err)
		}
	}

	WriteJSONResponse(w, http.StatusCreated, event)
}

func (et *eventServiceHandler) allEventHandler(w http.ResponseWriter, r *http.Request) {
	events, err := et.dbHandler.FindAllAvailableEvents()
	if err != nil {
		WriteErrorResponse(w, http.StatusInternalServerError, "internal server error")
		return
	}

	WriteJSONResponse(w, http.StatusOK, events)
}

func (et *eventServiceHandler) singleEventHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	searchCriteria, ok := vars["SearchCriteria"]
	if !ok {
		WriteErrorResponse(w, http.StatusBadRequest, "bad request: missing search criteria")
		return
	}

	searchQuery, ok := vars["search"]
	if !ok {
		WriteErrorResponse(w, http.StatusBadRequest, "bad request: missing search query")
		return
	}

	var event domain.Event
	var err error
	switch strings.ToLower(searchCriteria) {
	case "name":
		event, err = et.dbHandler.FindEventByName(searchQuery)
	case "id":
		event, err = et.dbHandler.FindEvent([]byte(searchQuery))
	default:
		WriteErrorResponse(w, http.StatusBadRequest, "bad request: invalid search criteria")
		return
	}
	if err != nil {
		WriteErrorResponse(w, http.StatusInternalServerError, "internal server error")
		return
	}

	WriteJSONResponse(w, http.StatusOK, event)
}
