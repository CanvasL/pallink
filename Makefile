SHELL := /bin/bash

SQLC ?= $(HOME)/.go/bin/sqlc
GOCTL ?= $(HOME)/.go/bin/goctl
COMPOSE ?= docker compose

.PHONY: help up up-build down build logs sqlc goctl goctl-api goctl-rpc

help:
	@echo "Targets:"
	@echo "  make up         Start all services with docker compose"
	@echo "  make up-build   Build and start all services with docker compose"
	@echo "  make down       Stop all services with docker compose"
	@echo "  make build      Build all docker images"
	@echo "  make logs       Tail docker compose logs"
	@echo "  make sqlc       Generate DAO code with sqlc"
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

goctl-api:
	cd gateway && $(GOCTL) api go -api gateway.api -dir .

goctl-rpc:
	cd user && $(GOCTL) rpc protoc user.proto --go_out=. --go-grpc_out=. --zrpc_out=.
	cd activity && $(GOCTL) rpc protoc activity.proto --go_out=. --go-grpc_out=. --zrpc_out=.
	cd notification && $(GOCTL) rpc protoc notification.proto --go_out=. --go-grpc_out=. --zrpc_out=.
	cd im && $(GOCTL) rpc protoc im.proto --go_out=. --go-grpc_out=. --zrpc_out=.

goctl: goctl-api goctl-rpc
