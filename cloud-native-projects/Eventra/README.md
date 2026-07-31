## Eventra 
Eventra is the online booking platform that takes events from different sources and give you booking for the those events. Its contains all the information about the events, locations etc.

``
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
``
