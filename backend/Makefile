.PHONY: run build test migrate-up migrate-down migrate-create seed reset-activities

-include .env

# Database URL for local development
DATABASE_URL ?= postgres://postgres:postgres@localhost:5432/velometric?sslmode=disable

# Set to true on the server to run migrate via Docker network
MIGRATE_VIA_DOCKER ?= false

ifeq ($(MIGRATE_VIA_DOCKER),true)
MIGRATE_CMD = docker run --rm \
	--network velometric_default \
	-v $(shell pwd)/migrations:/migrations \
	migrate/migrate \
	-path /migrations \
	-database "postgres://postgres:postgres@db:5432/velometric?sslmode=disable"
else
MIGRATE_CMD = $(MIGRATE_CMD)
endif

# Run the API server
run:
	go run cmd/api/main.go

# Build the binary
build:
	go build -o bin/api cmd/api/main.go

# Run tests
test:
	go test -v ./...

# Run migrations up
migrate-up:
	$(MIGRATE_CMD) up

# Rollback last migration
migrate-down:
	$(MIGRATE_CMD) down 1

# Rollback all migrations
migrate-reset:
	$(MIGRATE_CMD) down -all

# Create a new migration (usage: make migrate-create name=create_users)
migrate-create:
	migrate create -ext sql -dir migrations -seq $(name)

# Run seed script
seed:
	go run cmd/seed/main.go

# Check migration version
migrate-version:
	$(MIGRATE_CMD) version

# Force migration version (usage: make migrate-force version=1)
migrate-force:
	$(MIGRATE_CMD) force $(version)

# Reset activity data (keeps users and zones)
reset-activities:
	docker compose -f ../docker-compose.yml exec -T db psql -U postgres -d velometric -f - < scripts/reset-activities.sql
