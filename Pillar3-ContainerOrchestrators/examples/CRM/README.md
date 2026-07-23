# CRM - Containerized Microservices Platform

A multi-service FastAPI application demonstrating the **sidecar pattern** on Kubernetes.

## Architecture

```
Pod: crm
├── container: gateway  (port 8001) — mounts Students + Library in-process
├── container: students (port 8002) — standalone Students microservice
└── container: library  (port 8003) — standalone Library microservice
```

The **gateway** serves as the entry point with an intelligent `/portal/{id}` router that redirects to the appropriate sub-service via exception-based redirects. All 3 containers share the same network namespace (`localhost`).

## Quick Start

### Prerequisites

- [Docker](https://docs.docker.com/get-docker/)
- [Kubernetes cluster](https://kubernetes.io/docs/setup/) (recommended: [kind](https://kind.sigs.k8s.io))
- [kubectl](https://kubernetes.io/docs/tasks/tools/)
- [GNU Make](https://www.gnu.org/software/make/)

### Step 1: Start Local Registry

```bash
make registry-up
```

Starts a Docker registry on port 5001 for storing your built images (port 5000 is used by macOS AirPlay).

### Step 2: Build & Push Images

```bash
make build   # Builds all 3 Docker images
make push    # Pushes them to local registry
```

### Step 3: Create Kubernetes Cluster

**kind (recommended):**

```bash
cat <<EOF | kind create cluster --name crm --config=-
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
containerdConfigPatches:
- |-
  [plugins."io.containerd.grpc.v1.cri".registry.mirrors."localhost:5000"]
    endpoint = ["http://kind-registry:5000"]
EOF

docker network connect kind kind-registry 2>/dev/null || true
```

**minikube:**
```bash
minikube start
minikube addons enable registry
```

### Step 4: Deploy

```bash
make deploy
```

Applies all Kubernetes manifests (`namespace`, `configmap`, `deployment`, `service`, `ingress`).

### Step 5: Verify

```bash
kubectl get pods -n crm -w
# Wait for READY 3/3
```

### Step 6: Access

```bash
make port-forward
```

In another terminal:

```bash
curl http://localhost:8001/           # → {"content":"Hello"}
curl -L http://localhost:8001/portal/1 # → {"content":"Students"}
curl -L http://localhost:8001/portal/3 # → {"content":"Library"}
curl http://localhost:8001/portal/999  # → {"message":"University ERP Systems"}
```

## Makefile Reference

| Command | Action | Description |
|---|---|---|
| `make registry-up` | Start local Docker registry on :5001 | Runs `docker run -d -p 5001:5000 --name local-registry registry:2`. Creates a private Docker image registry accessible at port 5001. Required before building/pushing images. |
| `make registry-down` | Stop & remove registry container | Stops and deletes the local registry container. Use when done to free up resources. |
| `make build` | Build all 3 Docker images | Runs `docker build` for gateway (CRM root), students (Students/), and library (Library/). Tags each as `localhost:5001/crm-*:latest`. |
| `make push` | Push images to local registry | Uploads all 3 built images to the running local registry on port 5001 so Kubernetes can pull them. |
| `make all` | Build → push → deploy | One-command workflow: builds images, pushes to registry, then deploys to Kubernetes. |
| `make deploy` | Apply k8s manifests | Runs `kubectl apply -f k8s/` to create/update all Kubernetes resources (namespace, configmap, deployment, service, ingress). |
| `make restart` | Rollout restart deployment | Runs `kubectl rollout restart` to force a rolling restart of the CRM pod without changing the YAML. Useful after updating images. |
| `make port-forward` | Tunnel localhost:8001 → cluster | Runs `kubectl port-forward` to forward your local port 8001 to the gateway service inside Kubernetes. Open http://localhost:8001 to access the app. |
| `make logs-gateway` | Tail gateway container logs | Streams live logs from the gateway container. Useful for debugging the main entry point. |
| `make logs-students` | Tail students container logs | Streams live logs from the students sidecar container. |
| `make logs-library` | Tail library container logs | Streams live logs from the library sidecar container. |
| `make delete` | Delete the entire crm namespace | Runs `kubectl delete namespace crm` which removes ALL CRM resources (pod, service, configmap, ingress) at once. |

## API Endpoints

| Method | Path | Description |
|---|---|---|
| GET | `/` | Root health check |
| GET | `/portal/1` | Redirects to Students microservice |
| GET | `/portal/3` | Redirects to Library microservice |
| GET | `/portal/{id}` | Invalid ID returns fallback message |
| GET | `/students` | (via mount) Students sub-app |
| GET | `/library` | (via mount) Library sub-app |

## Project Structure

```
CRM/
├── main.py                  # Gateway entry point
├── Dockerfile               # Gateway container image
├── Makefile                 # Automation commands
├── requirements.txt         # Python dependencies
├── docker-compose.yml       # Local compose orchestration
├── pyproject.toml           # Project config (uv workspace)
├── k8s/                     # Kubernetes manifests
│   ├── namespace.yaml
│   ├── configmap.yaml
│   ├── deployment.yaml      # 3-container pod spec
│   ├── service.yaml
│   ├── ingress.yaml
│   └── README.md            # Detailed K8s deployment guide
├── Students/
│   ├── main.py              # Students microservice
│   ├── Dockerfile
│   └── requirements.txt
└── Library/
    ├── main.py              # Library microservice
    ├── Dockerfile
    └── requirements.txt
```

## Cleanup

```bash
make delete                 # Remove all K8s resources
kind delete cluster --name crm  # Delete kind cluster
make registry-down          # Stop local registry
```

## Kubernetes Concepts Demonstrated

- **Multi-container Pods** — 3 containers sharing network namespace
- **Sidecar Pattern** — Gateway with students/library as sidecars
- **ConfigMap Injection** — Environment variables from ConfigMap
- **Health Probes** — Readiness + liveness probes on gateway
- **Resource Management** — CPU/memory requests and limits per container
- **Local Registry** — Build, push, and deploy workflow
- **Service Discovery** — ClusterIP service for internal DNS
- **Ingress** — Path-based routing with NGINX ingress controller
