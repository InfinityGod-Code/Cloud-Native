# Eventra

Eventra is an online event booking platform built with Go. It aggregates events from different sources and lets users discover and book them. The service stores and serves all information about events, their locations, halls, and available seats.

## Project Structure

### Current Structure

```
Eventra/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point (wires DI, starts HTTP server)
├── internal/
│   ├── configuration/
│   │   ├── configuration.go     # Config loader with sensible defaults
│   │   └── config.json          # Runtime configuration
│   ├── domain/
│   │   └── models.go            # Core business models (Event, Location, Hall, User, Booking)
│   └── transport/
│       ├── events_handlers.go   # HTTP handlers for events
│       ├── events_routes.go     # Route registration
│       └── response.go          # JSON response helpers
├── repository/
│   ├── database_handler.go      # DatabaseHandler interface
│   ├── dblayer.go               # Persistence layer factory (DI registration)
│   └── mongo/
│       └── mongolayer.go        # MongoDB implementation of DatabaseHandler
├── go.mod                       # Dependency management
└── go.sum
```

### Future / Target Structure

The layout below is the planned reference structure this service will grow into as it evolves:

```
my-microservice/
├── .github/               # CI/CD workflows and automation templates
├── cmd/                   # Main entry points for the application
│   └── server/
│       └── main.go        # Minimal bootstrap (wires up DI, starts app)
├── internal/              # Private application and business code (cannot be imported by other services)
│   ├── config/            # Environment configurations and secret loading
│   ├── domain/            # Core business models, entities, and repository interfaces (No external dependencies)
│   ├── repository/        # Data access implementations (PostgreSQL, MongoDB, Redis)
│   │   ├── postgres/
│   │   └── redis/
│   ├── service/           # Use cases / Orchestration layer (orchestrates domain entities)
│   └── transport/         # Network/API layer (HTTP handlers, gRPC servers)
│       ├── grpc/
│       └── http/
├── pkg/                   # Public libraries that can be safely shared with other projects
│   └── sdk/               # Client SDK generated for consumers of this microservice
├── api/                   # API contracts and definitions
│   ├── openapi.yaml       # OpenAPI/Swagger documentation
│   └── protobuf/          # gRPC .proto definitions
├── scripts/               # Scripts for builds, migrations, or local database setups
├── build/                 # Packaging configurations (e.g., Dockerfiles)
├── deployments/           # Infrastructure manifests (Helm charts, Kubernetes manifests, Terraform)
├── test/                  # Additional external integration and end-to-end (E2E) tests
├── go.mod                 # Dependency management
├── go.sum
└── Makefile               # Task automation (build, test, lint, migrate)
```

## Prerequisites

- [Go](https://go.dev/dl/) 1.26+
- A running [MongoDB](https://www.mongodb.com/docs/manual/installation/) instance (default: `mongodb://localhost:27017`)

## Getting Started

Start a local MongoDB, then run the server:

```bash
go run ./cmd/server
```

By default the server connects to `mongodb://localhost:27017` and listens on `:8080`.

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/events` | Create a new event |
| `GET` | `/events` | List all available events |
| `GET` | `/events/{SearchCriteria}/{search}` | Find a single event by `name` or `id` |

### Examples

Create an event:

```bash
curl -X POST http://localhost:8080/events \
  -H "Content-Type: application/json" \
  -d '{"name":"Concert","duration":120,"startdate":1700000000,"enddate":1700003600,"location":{"name":"Main Hall","address":"1 Example St","country":"US"}}'
```

List all events:

```bash
curl http://localhost:8080/events
```

Find an event by name:

```bash
curl http://localhost:8080/events/name/Concert
```

## Tech Stack

- **Language:** Go
- **HTTP Router:** [gorilla/mux](https://github.com/gorilla/mux)
- **Database:** MongoDB via [mgo.v2](https://gopkg.in/mgo.v2)
- **Architecture:** Layered (transport → repository → domain)
