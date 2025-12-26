dotenv:
	@echo "Creating .env file..."
	cp cmd/.env.example cmd/.env

.PHONY: migrate
migrate:
	@migrate create -ext sql -dir database/migrations -format unix $(NAME)

build_order:
	@echo "Compiling Order..."
	@CGO_ENABLED=0 go build -ldflags="-w -s" -o ./bin/order ./cmd/main.go

start_docker:
	@echo "Starting Docker containers..."
	docker compose -f deployment/docker-compose.yml up --build -d
	
stop_docker:
	@echo "Stopping Docker containers..."
	docker compose -f deployment/docker-compose.yml down

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