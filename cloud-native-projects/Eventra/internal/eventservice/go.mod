module eventservice

go 1.26.5

require (
	github.com/gorilla/mux v1.8.1
	gopkg.in/mgo.v2 v2.0.0-20190816093944-a6b53ec6cb22
	infra v0.0.0
)

require (
	github.com/rabbitmq/amqp091-go v1.13.0 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

replace infra => ../../infra
