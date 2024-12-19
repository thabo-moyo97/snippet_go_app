.PHONY: migrate-create migrate-up migrate-down

migrate-create:
	@read -p "Enter migration name: " name; \
	version=$$(date +%Y%m%d%H%M%S); \
	echo "Creating migration $$version_$$name"; \
	touch cmd/web/migrations/$${version}_$$name.up.sql; \
	touch cmd/web/migrations/$${version}_$$name.down.sql

migrate-status:
	@echo "Current migration status:"
	@mysql -u$(DB_USER) -p$(DB_PASS) $(DB_NAME) -e "SELECT * FROM schema_migrations ORDER BY version DESC;"

migrate-up:
	docker compose --env-file .env -f .docker/compose/docker-compose.dev.yml exec app go run cmd/web/migrations/up.go

migrate-down:
	docker compose --env-file .env -f .docker/compose/docker-compose.dev.yml exec app go run cmd/web/migrations/down.go

.PHONY: help
help:
	@echo "Available commands:"
	@echo "  make migrate-create  - Create a new migration"
	@echo "  make migrate-status  - Show migration status" 


.PHONY: dev prod down logs

dev:
	docker compose --env-file .env -f .docker/compose/docker-compose.dev.yml up --remove-orphans --force-recreate

prod:
	docker compose --env-file .env -f .docker/compose/docker-compose.prod.yml up

down:
	docker compose --env-file .env -f .docker/compose/docker-compose.dev.yml down --remove-orphans

logs:
	docker compose --env-file .env -f .docker/compose/docker-compose.dev.yml logs -f

build-dev:
	docker build -t snippet-dev -f Dockerfile.dev .

build-dev-config:
	docker compose --env-file .env -f .docker/compose/docker-compose.dev.yml config

build-prod:
	docker build -t snippet-prod -f Dockerfile.prod .

start-debug: 
	DEBUG=true docker compose --env-file .env -f .docker/compose/docker-compose.dev.yml up --remove-orphans --force-recreate

tailwind:
	cd ./ui && npm run build

.DEFAULT_GOAL := dev


