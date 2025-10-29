.PHONY: help build up down logs test scout clean restart status

help:
	@echo "Comandos disponíveis:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

build:
	@echo "Building images..."
	@docker compose build

up:
	@echo "Starting services..."
	@docker compose up -d
	@make status

down:
	@echo "Stopping services..."
	@docker compose down

logs:
	@docker compose logs -f

logs-producer:
	@docker compose logs -f producer-1 producer-2 producer-3

logs-consumer:
	@docker compose logs -f consumer

logs-nginx:
	@docker compose logs -f nginx

test:
	@echo "Running tests..."
	@./test.sh

scout:
	@echo "Running security scan..."
	@./scout.sh

clean:
	@echo "Cleaning up..."
	@docker compose down -v
	@docker system prune -f

restart:
	@make down
	@make up

status:
	@docker compose ps

stats:
	@docker stats --no-stream

network:
	@docker network inspect docker_backend

mongo-shell:
	@docker compose exec mongodb mongosh -u $$MONGO_INITDB_ROOT_USERNAME -p $$MONGO_INITDB_ROOT_PASSWORD

rabbitmq-queues:
	@docker compose exec rabbitmq rabbitmqctl list_queues

all: build up test

dev:
	@make build
	@make up
	@make logs
