.PHONY: run build test migrate-up migrate-down migrate-create seed reset-activities

# Database URL for local development
DB_URL ?= postgres://postgres:postgres@localhost:5432/velometric?sslmode=disable

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
	migrate -path migrations -database "$(DB_URL)" up

# Rollback last migration
migrate-down:
	migrate -path migrations -database "$(DB_URL)" down 1

# Rollback all migrations
migrate-reset:
	migrate -path migrations -database "$(DB_URL)" down -all

# Create a new migration (usage: make migrate-create name=create_users)
migrate-create:
	migrate create -ext sql -dir migrations -seq $(name)

# Run seed script
seed:
	go run cmd/seed/main.go

# Check migration version
migrate-version:
	migrate -path migrations -database "$(DB_URL)" version

# Force migration version (usage: make migrate-force version=1)
migrate-force:
	migrate -path migrations -database "$(DB_URL)" force $(version)

# Reset activity data (keeps users and zones)
reset-activities:
	docker-compose -f ../docker-compose.yml exec -T db psql -U postgres -d velometric -f - < scripts/reset-activities.sql
