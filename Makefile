.DEFAULT_GOAL := help

.PHONY: help
help:
	@echo "Available commands:"
	@echo ""
	@echo "Setup:"
	@echo "  make dotenv              - Create .env file from example"
	@echo "  make setup-coralogix     - Setup Coralogix configuration"
	@echo ""
	@echo "Development:"
	@echo "  make build_order         - Build the application binary"
	@echo ""
	@echo "Docker - Infrastructure:"
	@echo "  make infra-up            - Start only infrastructure (DB, Kafka, Observability)"
	@echo "  make infra-down          - Stop infrastructure services"
	@echo ""
	@echo "Docker - Application:"
	@echo "  make up                  - Start all services (Infrastructure + Apps)"
	@echo "  make down                - Stop all services"
	@echo "  make ps                  - Show services status"
	@echo "  make rebuild             - Rebuild and restart application services"
	@echo "  make restart             - Restart all services"
	@echo "  make restart-api         - Restart only Order API"
	@echo "  make restart-worker      - Restart only Order Worker"
	@echo "  make restart-consumer    - Restart only Order Consumer"
	@echo ""
	@echo "Logs:"
	@echo "  make logs-api            - Show Order API logs"
	@echo "  make logs-consumer       - Show Order Consumer logs"
	@echo "  make logs-worker         - Show Order Worker logs"
	@echo "  make logs-otel           - Show OpenTelemetry Collector logs"
	@echo "  make logs-all            - Show all services logs"
	@echo ""
	@echo "Observability:"
	@echo "  make health-check        - Check services health"
	@echo "  make restart-otel        - Restart OpenTelemetry Collector"
	@echo ""
	@echo "Database:"
	@echo "  make migrate NAME=name   - Create new migration"
	@echo ""
	@echo "Testing & Quality:"
	@echo "  make test                - Run tests"
	@echo "  make cover               - Generate coverage report"
	@echo "  make lint                - Run linter"
	@echo "  make vulncheck           - Run vulnerability check"
	@echo "  make mocks               - Generate mocks"
	@echo ""

dotenv:
	@echo "Creating .env file..."
	cp cmd/.env.example cmd/.env

.PHONY: migrate
migrate:
	@migrate create -ext sql -dir database/migrations -format unix $(NAME)

build_order:
	@echo "Compiling Order..."
	@CGO_ENABLED=0 go build -ldflags="-w -s" -o ./bin/order ./cmd/main.go

# Docker Infrastructure Commands
.PHONY: infra-up
infra-up:
	@echo "Starting infrastructure services (Database, Kafka, Observability)..."
	docker compose -f deployment/docker-compose.yml up -d cockroachdb kafka kafka-init redpandadata otel-collector jaeger prometheus loki grafana
	@echo "Infrastructure services started successfully!"
	@echo "Available services:"
	@echo "  - CockroachDB:  http://localhost:8080"
	@echo "  - Kafka UI:     http://localhost:8085"
	@echo "  - Jaeger:       http://localhost:16686"
	@echo "  - Prometheus:   http://localhost:9090"
	@echo "  - Grafana:      http://localhost:3000"

.PHONY: infra-down
infra-down:
	@echo "Stopping infrastructure services..."
	docker compose -f deployment/docker-compose.yml down

# Docker Application Commands
.PHONY: up
up:
	@echo "Starting all services (Infrastructure + Applications)..."
	docker compose -f deployment/docker-compose.yml up --build -d
	@echo "All services started successfully!"
	@echo "Available services:"
	@echo "  - Order API:    http://localhost:8001"
	@echo "  - CockroachDB:  http://localhost:8080"
	@echo "  - Kafka UI:     http://localhost:8085"
	@echo "  - Jaeger:       http://localhost:16686"
	@echo "  - Prometheus:   http://localhost:9090"
	@echo "  - Grafana:      http://localhost:3000"

.PHONY: down
down:
	@echo "Stopping all services..."
	docker compose -f deployment/docker-compose.yml down

.PHONY: ps
ps:
	@echo "Showing services status..."
	docker compose -f deployment/docker-compose.yml ps

.PHONY: rebuild
rebuild:
	@echo "Rebuilding and restarting application services..."
	docker compose -f deployment/docker-compose.yml up --build -d order-api order-worker order-consumer
	@echo "Application services rebuilt successfully!"

.PHONY: restart
restart:
	@echo "Restarting all services..."
	docker compose -f deployment/docker-compose.yml restart

.PHONY: restart-api
restart-api:
	@echo "Restarting Order API..."
	docker compose -f deployment/docker-compose.yml restart order-api

.PHONY: restart-worker
restart-worker:
	@echo "Restarting Order Worker..."
	docker compose -f deployment/docker-compose.yml restart order-worker

.PHONY: restart-consumer
restart-consumer:
	@echo "Restarting Order Consumer..."
	docker compose -f deployment/docker-compose.yml restart order-consumer

# Legacy commands (deprecated, use 'up' and 'down' instead)
start_docker: up
	@echo "⚠️  'make start_docker' is deprecated, use 'make up' instead"

stop_docker: down
	@echo "⚠️  'make stop_docker' is deprecated, use 'make down' instead"

# Coralogix observability commands
.PHONY: logs-otel
logs-otel:
	@echo "Showing OpenTelemetry Collector logs..."
	docker compose -f deployment/docker-compose.yml logs -f otel-collector

.PHONY: logs-api
logs-api:
	@echo "Showing Order API logs..."
	docker compose -f deployment/docker-compose.yml logs -f order-api

.PHONY: logs-consumer
logs-consumer:
	@echo "Showing Order Consumer logs..."
	docker compose -f deployment/docker-compose.yml logs -f order-consumer

.PHONY: logs-worker
logs-worker:
	@echo "Showing Order Worker logs..."
	docker compose -f deployment/docker-compose.yml logs -f order-worker

.PHONY: logs-all
logs-all:
	@echo "Showing all services logs..."
	docker compose -f deployment/docker-compose.yml logs -f order-api order-consumer order-worker otel-collector

.PHONY: restart-otel
restart-otel:
	@echo "Restarting OpenTelemetry Collector..."
	docker compose -f deployment/docker-compose.yml restart otel-collector

.PHONY: health-check
health-check:
	@echo "Checking services health..."
	@echo "\n=== Order API ==="
	@curl -s http://localhost:8000/health | jq . || echo "Order API is not responding"
	@echo "\n=== Jaeger UI ==="
	@curl -s http://localhost:16686 > /dev/null && echo "✓ Jaeger is running" || echo "✗ Jaeger is not running"
	@echo "\n=== Prometheus ==="
	@curl -s http://localhost:9090/-/healthy > /dev/null && echo "✓ Prometheus is running" || echo "✗ Prometheus is not running"
	@echo "\n=== Grafana ==="
	@curl -s http://localhost:3000/api/health | jq . || echo "✗ Grafana is not running"

.PHONY: setup-coralogix
setup-coralogix:
	@echo "Setting up Coralogix configuration..."
	@if [ ! -f deployment/.env ]; then \
		cp deployment/.env.example deployment/.env; \
		echo "✓ Created deployment/.env file"; \
		echo "⚠️  Please edit deployment/.env and add your CORALOGIX_PRIVATE_KEY"; \
	else \
		echo "✓ deployment/.env already exists"; \
	fi
	@echo "\nNext steps:"
	@echo "1. Edit deployment/.env and add your Coralogix private key"
	@echo "2. Run 'make start_docker' to start all services"
	@echo "3. Run 'make logs-otel' to verify Coralogix connection"

.PHONY: mockery
mocks:
	@echo "Generating mocks..."
	go install github.com/vektra/mockery/v3@v3.5.0
	mockery
	
lint:
	@echo "Running linter..."
	@echo "Installing golangci-lint..."
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.2.1
	GOGC=20 golangci-lint run --config .golangci.yml ./...

.PHONY: test
test:
	@echo "Running tests..."
	go test -count=1 -race -covermode=atomic -coverprofile=coverage.out ./...

cover:
	@echo "Generating coverage report..."
	go tool cover -html=coverage.out

.PHONY: vulncheck
vulncheck:
	@echo "Running vulnerability check..."
	go install golang.org/x/vuln/cmd/govulncheck@latest
	govulncheck ./...