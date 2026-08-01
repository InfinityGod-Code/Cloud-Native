package mongo

import (
	"eventra/internal/domain"
	"eventra/internal/repository"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
	DB     = "myevents"
	USERS  = "users"
	EVENTS = "events"
)

// MongoLayer that gives use the session for connecting to MongoDB.
// Also this is is the implementation for the
type MongoDBLayer struct {
	session *mgo.Session
}

// creating constructor function for returning MongoDBLayer

func NewMongoDBLayer(connection string) (repository.DatabaseHandler, error) {
	/*
		Dial establishes a new session to the cluster identified by the given seed server(s).
		The session will enable communication with all of the servers in the cluster,
		so the seed servers are used only to find out about the cluster topology.
	*/
	conn, err := mgo.Dial(connection)
	return &MongoDBLayer{
		session: conn,
	}, err

}

// adding method to the above type struct so that it implements the [repository.DatabaseHandler]

func (mDB *MongoDBLayer) FindEvent(id []byte) (domain.Event, error) {
	s := mDB.getFreshSession()
	defer s.Close()
	e := domain.Event{}

	err := s.DB(DB).C(EVENTS).FindId(bson.ObjectIdHex(string(id))).One(&e)
	return e, err
}

func (mDB *MongoDBLayer) AddEvent(e domain.Event) ([]byte, error) {
	s := mDB.getFreshSession()
	defer s.Close()

	if !e.ID.Valid() {
		e.ID = bson.NewObjectId()
	}

	if !e.Location.ID.Valid() {
		e.Location.ID = bson.NewObjectId()
	}

	return []byte(e.ID), s.DB(DB).C(EVENTS).Insert(e)
}

func (mDB *MongoDBLayer) FindEventByName(name string) (domain.Event, error) {
	s := mDB.getFreshSession()
	defer s.Close()
	e := domain.Event{}
	err := s.DB(DB).C(EVENTS).Find(bson.M{"name": name}).One(&e)
	return e, err
}

func (mDB *MongoDBLayer) FindAllAvailableEvents() ([]domain.Event, error) {
	s := mDB.getFreshSession()
	defer s.Close()
	events := []domain.Event{}
	err := s.DB(DB).C(EVENTS).Find(nil).All(&events)
	return events, err
}

func (mgoLayer *MongoDBLayer) getFreshSession() *mgo.Session {
	return mgoLayer.session.Copy()
}
