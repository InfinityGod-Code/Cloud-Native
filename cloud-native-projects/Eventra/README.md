# Eventra

Eventra is an online event booking platform built with Go. It aggregates events from different sources and lets users discover and book them. The service stores and serves all information about events, their locations, halls, and available seats.

## Project Structure

### Current Structure

The project is a monorepo of independent Go modules, wired together with a root `go.work` workspace:

```
Eventra/
├── go.work                     # Go workspace tying the modules together
├── infra/                      # Own module (`infra`): shared infrastructure
│   ├── go.mod
│   ├── go.sum
│   └── rabbitmq.go             # RabbitMQ connection + messaging helpers (AMQP)
├── internal/
│   ├── eventservice/           # Own module (`eventservice`): event API service
│   │   ├── cmd/
│   │   │   └── server/
│   │   │       └── main.go     # Application entry point (wires DI, starts HTTP server)
│   │   ├── configuration/
│   │   │   ├── configuration.go # Config loader with sensible defaults
│   │   │   └── config.json      # Runtime configuration
│   │   ├── domain/
│   │   │   └── models.go        # Core business models (Event, Location, Hall, User, Booking)
│   │   ├── repository/
│   │   │   ├── database_handler.go  # DatabaseHandler interface
│   │   │   ├── dblayer.go           # Persistence layer factory (DI registration)
│   │   │   └── mongo/
│   │   │       └── mongolayer.go    # MongoDB implementation of DatabaseHandler
│   │   ├── transport/
│   │   │   ├── events_handlers.go   # HTTP handlers for events
│   │   │   ├── events_routes.go     # Route registration
│   │   │   └── response.go          # JSON response helpers
│   │   ├── go.mod
│   │   └── go.sum
│   ├── bookingservice/         # Own module (`bookingservice`): booking API service
│   │   ├── cmd/
│   │   │   └── server/
│   │   │       └── main.go
│   │   └── go.mod
│   └── lib/
└── build/                      # Docker & deployment assets
    ├── Dockerfile
    ├── docker-compose.yml      # mongo + rabbitmq + eventra
    ├── config.docker.json
    ├── Makefile
    └── RUNNING.md
```

`eventservice` depends on the local `infra` module via a `replace` directive in `internal/eventservice/go.mod` (`replace infra => ../../infra`). The root `go.work` enables running Go commands across all modules from the project root.

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
- A running [RabbitMQ](https://www.rabbitmq.com/docs/download) instance (default: `amqp://guest:guest@localhost:5672/`)

## Getting Started

Start a local MongoDB and RabbitMQ, then run the server from the `Eventra/` root (the root `go.work` resolves the modules):

```bash
go run ./internal/eventservice/cmd/server
```

By default the server connects to `mongodb://localhost:27017`, `amqp://guest:guest@localhost:5672/`, and listens on `localhost:8181`.

The connection strings come from `internal/eventservice/configuration/config.json`:

| Field | Default (native) | Docker value |
|-------|------------------|--------------|
| `dbconnection` | `mongodb://127.0.0.1` | `mongodb://mongo:27017` |
| `rabbitmq_url` | `amqp://guest:guest@localhost:5672/` | `amqp://guest:guest@rabbitmq:5672/` |
| `restfulapi_endpoint` | `localhost:8181` | `:8181` |

## RabbitMQ

The app connects to RabbitMQ at startup via the `infra` package (AMQP client `amqp091-go`). The connection is established before the HTTP server starts; a retry loop tolerates RabbitMQ coming up slightly later (e.g. under docker-compose).

### Accessing the Management Dashboard

Start the stack with Docker Compose (from the `Eventra/` root):

```bash
docker compose -f build/docker-compose.yml up -d --build
```

Then open the dashboard:

- **URL:** http://localhost:15672
- **Username:** `guest`
- **Password:** `guest`

The dashboard lets you inspect connections, channels, queues, exchanges, and publish/consume messages interactively. AMQP traffic for the app runs on port `5672`.

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/events` | Create a new event |
| `GET` | `/events` | List all available events |
| `GET` | `/events/{SearchCriteria}/{search}` | Find a single event by `name` or `id` |

### Examples

Create an event:

```bash
curl -X POST http://localhost:8181/events \
  -H "Content-Type: application/json" \
  -d '{"name":"Concert","duration":120,"startdate":1700000000,"enddate":1700003600,"location":{"name":"Main Hall","address":"1 Example St","country":"US"}}'
```

List all events:

```bash
curl http://localhost:8181/events
```

Find an event by name:

```bash
curl http://localhost:8181/events/name/Concert
```

## Tech Stack

- **Language:** Go
- **HTTP Router:** [gorilla/mux](https://github.com/gorilla/mux)
- **Database:** MongoDB via [mgo.v2](https://gopkg.in/mgo.v2)
- **Message Broker:** [RabbitMQ](https://www.rabbitmq.com/) via [amqp091-go](https://github.com/rabbitmq/amqp091-go)
- **Architecture:** Layered (transport → service → repository → domain + infra)
