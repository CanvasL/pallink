SHELL := /bin/bash

SQLC ?= $(HOME)/.go/bin/sqlc
GOCTL ?= $(HOME)/.go/bin/goctl
COMPOSE ?= docker compose
ALIYUN_SYNC ?= ./deploy/aliyun/sync-to-acr.sh
ALIYUN_ENV_FILE ?= ./deploy/aliyun/compose.env
ALIYUN_COMPOSE ?= ./docker-compose.aliyun.yml
SWAGGER_PATCH ?= ./deploy/swagger/patch.sh

.PHONY: help up up-build down down-v build logs sqlc swagger goctl goctl-api goctl-rpc aliyun-compose aliyun-sync aliyun-up aliyun-down

help:
	@echo "Targets:"
	@echo "  make up         Start all services with docker compose"
	@echo "  make up-build   Build and start all services with docker compose"
	@echo "  make down       Stop all services with docker compose"
	@echo "  make build      Build all docker images"
	@echo "  make logs       Tail docker compose logs"
	@echo "  make sqlc       Generate DAO code with sqlc"
	@echo "  make swagger    Regenerate deploy swagger json"
	@echo "  make goctl-api  Regenerate gateway API scaffolding"
	@echo "  make goctl-rpc  Regenerate rpc/proto scaffolding"
	@echo "  make goctl      Run goctl-api and goctl-rpc"
	@echo "  make aliyun-compose  Regenerate the aliyun compose file"
	@echo "  make aliyun-sync     Build/push app images and mirror third-party images to ACR"
	@echo "  make aliyun-up       Start services with docker-compose.aliyun.yml"
	@echo "  make aliyun-down     Stop services with docker-compose.aliyun.yml"

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

aliyun-compose:
	$(ALIYUN_SYNC) --generate-only --compose-file ./docker-compose.yml --output-file $(ALIYUN_COMPOSE) --env-file $(ALIYUN_ENV_FILE)

aliyun-sync:
	$(ALIYUN_SYNC) --compose-file ./docker-compose.yml --output-file $(ALIYUN_COMPOSE) --env-file $(ALIYUN_ENV_FILE)

aliyun-up:
	$(COMPOSE) --env-file $(ALIYUN_ENV_FILE) -f $(ALIYUN_COMPOSE) up -d

aliyun-down:
	$(COMPOSE) --env-file $(ALIYUN_ENV_FILE) -f $(ALIYUN_COMPOSE) down
