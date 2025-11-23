.PHONY: help build run stop clean test migrate-up migrate-down migrate-create logs dev

# Default target
help:
	@echo "Available targets:"
	@echo "  make build            - Build Docker images"
	@echo "  make run              - Start all services with Docker Compose"
	@echo "  make dev              - Start services in development mode"
	@echo "  make stop             - Stop all services"
	@echo "  make clean            - Stop and remove all containers, volumes"
	@echo "  make logs             - View logs from all services"
	@echo "  make logs-api         - View API logs"
	@echo "  make logs-worker      - View worker logs"
	@echo "  make migrate-up       - Run database migrations"
	@echo "  make migrate-down     - Rollback 1 migration"
	@echo "  make migrate-down-all - Rollback all migrations"
	@echo "  make migrate-create   - Create a new migration (name=migration_name)"
	@echo "  make migrate-status   - Check migration status"
	@echo "  make test             - Run tests"
	@echo "  make lint             - Run linter"
	@echo "  make deps             - Download Go dependencies"
	@echo "  make shell-api        - Open shell in API container"
	@echo "  make shell-db         - Open PostgreSQL shell"

# Build Docker images
build:
	docker-compose build

# Start all services
run:
	docker-compose up -d
	@echo "Services started!"
	@echo "API: http://localhost:8080"
	@echo "MailHog UI: http://localhost:8025"
	@echo "SuperTokens: http://localhost:3567"

# Development mode (with logs)
dev:
	docker-compose up

# Stop all services
stop:
	docker-compose stop

# Clean up everything
clean:
	docker-compose down -v
	@echo "All containers and volumes removed"

# View logs
logs:
	docker-compose logs -f

logs-api:
	docker-compose logs -f api

logs-worker:
	docker-compose logs -f worker

# Database migrations
migrate-up:
	@echo "Running migrations..."
	@docker-compose exec -T postgres psql -U utmuser -d utm_backend -c "CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";" 2>/dev/null || true
	docker-compose run --rm api go run cmd/migrate/main.go -up

migrate-down:
	@echo "Rolling back migrations..."
	docker-compose run --rm api go run cmd/migrate/main.go -down -steps=1

migrate-down-all:
	@echo "Rolling back all migrations..."
	docker-compose run --rm api go run cmd/migrate/main.go -down -steps=100

migrate-create:
	@if [ -z "$(name)" ]; then \
		echo "Usage: make migrate-create name=migration_name"; \
		exit 1; \
	fi
	@echo "Creating migration: $(name)"
	@TIMESTAMP=$$(date +%s); \
	UP_FILE="migrations/$${TIMESTAMP}_$(name).up.sql"; \
	DOWN_FILE="migrations/$${TIMESTAMP}_$(name).down.sql"; \
	touch "$$UP_FILE" "$$DOWN_FILE"; \
	echo "-- Migration: $(name)" > "$$UP_FILE"; \
	echo "-- Write your UP migration here" >> "$$UP_FILE"; \
	echo "" >> "$$UP_FILE"; \
	echo "-- Migration: $(name)" > "$$DOWN_FILE"; \
	echo "-- Write your DOWN migration here" >> "$$DOWN_FILE"; \
	echo "" >> "$$DOWN_FILE"; \
	echo "Created migration files:"; \
	echo "  $$UP_FILE"; \
	echo "  $$DOWN_FILE"

migrate-status:
	@echo "Checking migration status..."
	docker-compose exec postgres psql -U utmuser -d utm_backend -c "SELECT version, dirty FROM schema_migrations;" 2>/dev/null || echo "No migrations applied yet"

# Development helpers
deps:
	go mod download
	go mod tidy

test:
	go test -v -race -coverprofile=coverage.out ./...

lint:
	golangci-lint run

# Shell access
shell-api:
	docker-compose exec api sh

shell-db:
	docker-compose exec postgres psql -U utmuser -d utm_backend

# Docker commands
docker-rebuild:
	docker-compose up -d --build

docker-restart:
	docker-compose restart

docker-ps:
	docker-compose ps

# Generate SuperTokens API key
generate-api-key:
	@echo "Generated API Key: $$(openssl rand -hex 32)"

