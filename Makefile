.PHONY: migrate
migrate:
	@migrate create -ext sql -dir database/migrations -format unix $(NAME)

build_outbox:
	@echo "Compiling Outbox..."
	@CGO_ENABLED=0 go build -ldflags="-w -s" -o ./bin/outbox ./cmd/main.go

start_docker:
	docker compose -f deployment/docker-compose.yml up --build -d

stop_docker:
	docker compose -f deployment/docker-compose.yml down