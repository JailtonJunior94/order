start_docker:
	docker compose -f deployment/docker-compose.yml up --build -d

stop_docker:
	docker compose -f deployment/docker-compose.yml down