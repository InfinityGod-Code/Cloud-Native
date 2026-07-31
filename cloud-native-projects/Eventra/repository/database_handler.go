package repository

import (
	"eventra/internal/domain"
)

type DatabaseHandler interface {
	AddEvent(domain.Event) ([]byte, error)
	FindEvent([]byte) (domain.Event, error)
	FindEventByName(string) (domain.Event, error)
	FindAllAvailableEvents() ([]domain.Event, error)
}
