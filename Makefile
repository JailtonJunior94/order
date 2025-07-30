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

start_docker_minimal:
	@echo "Starting Docker containers..."
	docker compose -f deployment/docker-compose.yml up --build -d cockroachdb zookeeper broker kafka_ui order_migration

stop_docker:
	@echo "Stopping Docker containers..."
	docker compose -f deployment/docker-compose.yml down

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