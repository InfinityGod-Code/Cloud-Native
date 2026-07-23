REGISTRY = localhost:5001

.PHONY: all build push deploy restart registry-up registry-down

all: build push deploy

registry-up:
	docker run -d -p 5001:5000 --name local-registry registry:2 || echo "Registry already running"

registry-down:
	docker stop local-registry && docker rm local-registry || true

build:
	docker build -t $(REGISTRY)/crm-gateway:latest .
	docker build -t $(REGISTRY)/crm-students:latest ./Students
	docker build -t $(REGISTRY)/crm-library:latest ./Library

push:
	docker push $(REGISTRY)/crm-gateway:latest
	docker push $(REGISTRY)/crm-students:latest
	docker push $(REGISTRY)/crm-library:latest

minikube-load:
	minikube image load $(REGISTRY)/crm-gateway:latest
	minikube image load $(REGISTRY)/crm-students:latest
	minikube image load $(REGISTRY)/crm-library:latest

deploy:
	kubectl apply -f k8s/namespace.yaml
	kubectl apply -f k8s/configmap.yaml
	kubectl apply -f k8s/deployment.yaml
	kubectl apply -f k8s/service.yaml
	kubectl apply -f k8s/ingress.yaml

restart:
	kubectl rollout restart -n crm deployment/crm

delete:
	kubectl delete namespace crm

logs-gateway:
	kubectl logs -n crm -l app=crm -c gateway -f

logs-students:
	kubectl logs -n crm -l app=crm -c students -f

logs-library:
	kubectl logs -n crm -l app=crm -c library -f

port-forward:
	kubectl port-forward -n crm svc/gateway-service 8001:8001
