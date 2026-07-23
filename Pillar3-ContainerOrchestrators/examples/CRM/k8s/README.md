# CRM Kubernetes Deployment Guide

## Architecture

This deployment uses the **sidecar pattern** — a single pod running 3 containers:

```
Pod: crm
├── container: gateway  (port 8001) — mounts Students + Library in-process
├── container: students (port 8002) — standalone Students microservice
└── container: library  (port 8003) — standalone Library microservice
```

All 3 containers share the same network namespace (`localhost`).

## Prerequisites

- [Docker](https://docs.docker.com/get-docker/)
- [Kubernetes cluster](https://kubernetes.io/docs/setup/) (local: [kind](https://kind.sigs.k8s.io) / [minikube](https://minikube.sigs.k8s.io), cloud: any K8s cluster)
- [kubectl](https://kubernetes.io/docs/tasks/tools/)
- (Optional) [Ingress NGINX Controller](https://kubernetes.github.io/ingress-nginx/deploy/) for ingress support

---

## Step 0: Start Local Registry

Run a local Docker registry on port 5001 (one-time, port 5000 is used by macOS AirPlay):

```bash
make registry-up
```

This starts a `registry:2` container that stores images locally.

## Step 1: Build & Push Images

The `Makefile` at the CRM root automates everything:

```bash
# Build all 3 images tagged for local registry
make build

# Push to local registry
make push
```

Or do it all in one command:

```bash
make all
```

### Manual equivalent

```bash
docker build -t localhost:5001/crm-gateway:latest .
docker build -t localhost:5001/crm-students:latest ./Students
docker build -t localhost:5001/crm-library:latest ./Library

docker push localhost:5001/crm-gateway:latest
docker push localhost:5001/crm-students:latest
docker push localhost:5001/crm-library:latest
```

### For remote clusters

Replace `localhost:5001` with your registry address (e.g., `yourdockerhub/crm-gateway:latest`) and update `image:` in `deployment.yaml` accordingly.

---

## Step 2: Deploy to Kubernetes

```bash
# Create namespace + all resources
kubectl apply -f k8s/namespace.yaml
kubectl apply -f k8s/configmap.yaml
kubectl apply -f k8s/deployment.yaml
kubectl apply -f k8s/service.yaml
kubectl apply -f k8s/ingress.yaml

# Or apply everything at once
kubectl apply -f k8s/
```

---

## Step 3: Verify the Deployment

```bash
# Check pod status (3/3 containers ready)
kubectl get pods -n crm

# Check services
kubectl get svc -n crm

# Check ingress
kubectl get ingress -n crm

# View logs from a specific container
kubectl logs -n crm -l app=crm -c gateway
kubectl logs -n crm -l app=crm -c students
kubectl logs -n crm -l app=crm -c library
```

---

## Step 4: Access the Application

### Port-forward (quick test)

```bash
kubectl port-forward -n crm svc/gateway-service 8001:8001
```

Then visit: http://localhost:8001

### Via Ingress (requires Ingress Controller)

Get the ingress IP/host:

```bash
kubectl get ingress -n crm gateway-ingress
```

Add a hosts entry or use the IP to access.

---

## Step 5: Test the Endpoints

```bash
# Root
curl http://localhost:8001/
# → {"content": "Hello"}

# Gateway redirect to Students (via portal)
curl -L http://localhost:8001/portal/1
# → 307 → /students → {"content": "Students"}

# Gateway redirect to Library (via portal)
curl -L http://localhost:8001/portal/3
# → 307 → /library → {"content": "Library"}

# Invalid portal ID
curl http://localhost:8001/portal/999
# → {"message": "University ERP Systems"}
```

---

## Useful Commands

```bash
# Describe pod details
kubectl describe pod -n crm -l app=crm

# Watch pod status
kubectl get pods -n crm -w

# Exec into a container
kubectl exec -n crm -it deploy/crm -c gateway -- sh

# Check resource usage
kubectl top pods -n crm

# Scale the deployment
kubectl scale deployment -n crm crm --replicas=3

# Delete everything
kubectl delete namespace crm
```

---

## Manifests Overview

| File | Kind | Purpose |
|---|---|---|
| `namespace.yaml` | Namespace | Isolates all CRM resources |
| `configmap.yaml` | ConfigMap | Shared environment variables |
| `deployment.yaml` | Deployment | 3-container pod (gateway + students + library) |
| `service.yaml` | Service | ClusterIP exposing gateway port 8001 |
| `ingress.yaml` | Ingress | External HTTP routing to gateway |

## Key Kubernetes Concepts Demonstrated

| Concept | Implementation |
|---|---|
| **Multi-container Pod** | 3 containers sharing network namespace |
| **Sidecar Pattern** | Gateway handles traffic; students/library as sidecars |
| **ConfigMap Injection** | Env vars from ConfigMap into all containers |
| **Health Probes** | Readiness + liveness probes on gateway container |
| **Resource Management** | requests/limits per container |
| **Service Discovery** | ClusterIP service for internal DNS resolution |
| **Ingress** | Path-based routing with NGINX ingress controller |
