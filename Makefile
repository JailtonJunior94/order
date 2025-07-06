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