SHELL := /bin/bash

SQLC ?= $(HOME)/.go/bin/sqlc
GOCTL ?= $(HOME)/.go/bin/goctl
COMPOSE ?= docker compose
SWAGGER_PATCH ?= ./deploy/swagger/patch.sh

.PHONY: help up up-build down down-v build logs sqlc swagger goctl goctl-api goctl-rpc k8s-build-images k8s-apply k8s-delete k8s-status k8s-kind-up k8s-kind-down k8s-k9s

KUBECTL ?= kubectl
KIND ?= kind
K9S ?= k9s
K8S_NAMESPACE ?= pallink
KIND_CLUSTER_NAME ?= pallink-local
KUBE_CONTEXT ?= kind-$(KIND_CLUSTER_NAME)
K8S_OVERLAY ?= ./deploy/k8s/overlays/local

help:
	@echo "Targets:"
	@echo "  make up         Start all services with docker compose"
	@echo "  make up-build   Build and start all services with docker compose"
	@echo "  make down       Stop all services with docker compose"
	@echo "  make build      Build all docker images"
	@echo "  make logs       Tail docker compose logs"
	@echo "  make sqlc       Generate DAO code with sqlc"
	@echo "  make swagger    Regenerate deploy swagger json"
	@echo "  make k8s-build-images Build local Kubernetes images"
	@echo "  make k8s-apply  Apply local Kubernetes manifests to current cluster"
	@echo "  make k8s-delete Delete local Kubernetes manifests from current cluster"
	@echo "  make k8s-status Show Kubernetes pods and services"
	@echo "  make k8s-kind-up Create/reuse kind cluster, build/load images, apply manifests"
	@echo "  make k8s-kind-down Delete the local kind cluster"
	@echo "  make k8s-k9s    Open k9s on the local kind context and namespace"
	@echo "  make goctl-api  Regenerate gateway API scaffolding"
	@echo "  make goctl-rpc  Regenerate rpc/proto scaffolding"
	@echo "  make goctl      Run goctl-api and goctl-rpc"

up:
	$(COMPOSE) up -d

up-build:
	$(COMPOSE) up --build -d

down:
	$(COMPOSE) down

down-v:
	$(COMPOSE) down -v

build:
	$(COMPOSE) build

logs:
	$(COMPOSE) logs -f

sqlc:
	$(SQLC) generate

swagger:
	$(GOCTL) api swagger -api ./gateway/gateway.api -dir ./deploy/swagger --filename swagger
	$(SWAGGER_PATCH) ./deploy/swagger/swagger.json

goctl-api:
	$(GOCTL) api go -api ./gateway/gateway.api -dir ./gateway
	$(GOCTL) api swagger -api ./gateway/gateway.api -dir ./deploy/swagger --filename swagger
	$(SWAGGER_PATCH) ./deploy/swagger/swagger.json

goctl-rpc:
	$(GOCTL) rpc protoc ./user/user.proto --go_out=./user --go-grpc_out=./user --zrpc_out=./user
	$(GOCTL) rpc protoc ./activity/activity.proto --go_out=./activity --go-grpc_out=./activity --zrpc_out=./activity
	$(GOCTL) rpc protoc ./notification/notification.proto --go_out=./notification --go-grpc_out=./notification --zrpc_out=./notification
	$(GOCTL) rpc protoc ./im/im.proto --go_out=./im --go-grpc_out=./im --zrpc_out=./im

goctl: goctl-api goctl-rpc

k8s-build-images:
	./deploy/k8s/build-images.sh

k8s-apply:
	$(KUBECTL) apply -k $(K8S_OVERLAY)

k8s-delete:
	$(KUBECTL) delete -k $(K8S_OVERLAY) --ignore-not-found

k8s-status:
	$(KUBECTL) -n $(K8S_NAMESPACE) get pods,svc,ingress

k8s-kind-up:
	./deploy/k8s/kind-up.sh

k8s-kind-down:
	$(KIND) delete cluster --name $(KIND_CLUSTER_NAME)

k8s-k9s:
	@if ! command -v $(K9S) >/dev/null 2>&1; then \
		echo "missing required command: $(K9S)"; \
		exit 1; \
	fi
	$(K9S) --context $(KUBE_CONTEXT) --namespace $(K8S_NAMESPACE)
