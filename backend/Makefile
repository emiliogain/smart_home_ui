# Makefile for Smart Home Backend

.PHONY: help build run test clean lint fmt deps docker-build docker-run \
        migrate-up migrate-down migrate-create migrate-status dev-setup \
        simulate simulate-scenario

# Build the application
build:
	go build -o bin/smart-home-backend ./cmd/smart-home-backend

# Run the application in development mode (applies embedded DB migrations on startup)
run:
	go run ./cmd/smart-home-backend

# Run tests
test:
	go test -v ./...

# Run tests with race detector
test-race:
	go test -race -v ./...

# Run tests with coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	rm -rf bin/ coverage.out coverage.html
	go clean

# Run linter
lint:
	golangci-lint run

# Run go vet
vet:
	go vet ./...

# Format code
fmt:
	go fmt ./...
	goimports -w .

# Download dependencies
deps:
	go mod download
	go mod tidy

# --- Database Migrations (goose) ---

MIGRATE_DIR := migrations
DATABASE_URL ?= postgres://postgres:password@localhost:5432/smarthome?sslmode=disable

migrate-up:
	goose -dir $(MIGRATE_DIR) postgres "$(DATABASE_URL)" up

migrate-down:
	goose -dir $(MIGRATE_DIR) postgres "$(DATABASE_URL)" down

migrate-status:
	goose -dir $(MIGRATE_DIR) postgres "$(DATABASE_URL)" status

migrate-create:
	@read -p "Migration name: " name; \
	goose -dir $(MIGRATE_DIR) create $$name sql

# --- Docker ---

docker-build:
	docker build -t smart-home-backend .

docker-run:
	docker run -p 8080:8080 smart-home-backend

# --- Development Setup ---

dev-setup: deps
	@echo "Setting up development environment..."
	@echo "1. Copy .env.example to .env and configure"
	@echo "2. Copy config/config.example.yaml to config/config.yaml (if you use an example file)"
	@echo "3. Install required tools:"
	@echo "   go install github.com/pressly/goose/v3/cmd/goose@latest"
	@echo "   brew install golangci-lint   (or see https://golangci-lint.run/usage/install/)"
	@echo "   go install golang.org/x/tools/cmd/goimports@latest"

# --- Simulator ---

SCENARIO ?= comfortable

simulate:
	go run ./cmd/simulator --cycle

simulate-scenario:
	go run ./cmd/simulator --scenario $(SCENARIO)

# --- Pre-commit ---

check: vet lint test
	@echo "All checks passed."

# --- Help ---

help:
	@echo "Available commands:"
	@echo ""
	@echo "  Build & Run:"
	@echo "    build           - Build the application binary"
	@echo "    run             - Run in development mode (migrations run in-process)"
	@echo ""
	@echo "  Testing:"
	@echo "    test            - Run tests"
	@echo "    test-race       - Run tests with race detector"
	@echo "    test-coverage   - Run tests with coverage report"
	@echo ""
	@echo "  Code Quality:"
	@echo "    lint            - Run golangci-lint"
	@echo "    vet             - Run go vet"
	@echo "    fmt             - Format code"
	@echo "    check           - Run vet + lint + test (pre-commit)"
	@echo ""
	@echo "  Database:"
	@echo "    migrate-up      - Apply all pending migrations"
	@echo "    migrate-down    - Roll back one migration"
	@echo "    migrate-status  - Show migration status"
	@echo "    migrate-create  - Create a new migration file"
	@echo ""
	@echo "  Docker:"
	@echo "    docker-build    - Build Docker image"
	@echo "    docker-run      - Run in Docker container"
	@echo ""
	@echo "  Other:"
	@echo "    deps            - Download and tidy dependencies"
	@echo "    dev-setup       - Setup development environment"
	@echo "    clean           - Clean build artifacts"
