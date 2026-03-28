# Root Makefile for Smart Home Full Stack Application

.PHONY: help setup build run test clean lint docker-up docker-down

# Setup development environment
setup:
	@echo "Setting up Smart Home development environment..."
	cd backend && make dev-setup
	cd frontend && npm install || echo "Frontend setup pending - choose framework first"
	@echo "✅ Setup complete!"

# Build both backend and frontend
build:
	@echo "Building backend..."
	cd backend && make build
	@echo "Building frontend..."
	cd frontend && npm run build || echo "Frontend build pending - implement after framework selection"

# Run development servers
run-backend:
	cd backend && make run

run-frontend:
	cd frontend && npm run dev || echo "Frontend dev server pending"

# Run both in development mode (requires tmux or separate terminals)
run-dev:
	@echo "Start backend: make run-backend"
	@echo "Start frontend: make run-frontend" 
	@echo "Or use: docker-compose up for full stack"

# Testing
test:
	@echo "Running backend tests..."
	cd backend && make test
	@echo "Running frontend tests..."
	cd frontend && npm test || echo "Frontend tests pending"

# Linting
lint:
	@echo "Linting backend..."
	cd backend && make lint
	@echo "Linting frontend..."
	cd frontend && npm run lint || echo "Frontend linting pending"

# Cleanup
clean:
	cd backend && make clean
	cd frontend && npm run clean || rm -rf dist/ build/ || true
	docker-compose down -v --rmi all || true

# Docker operations
docker-up:
	docker-compose up --build

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f

# Database operations
db-up:
	docker-compose up -d postgres redis

db-down:
	docker-compose stop postgres redis

# Full application lifecycle
start: docker-up
stop: docker-down
restart: docker-down docker-up

help:
	@echo "Smart Home Full Stack Commands:"
	@echo ""
	@echo "Development:"
	@echo "  setup         - Setup development environment"
	@echo "  run-backend   - Run backend development server" 
	@echo "  run-frontend  - Run frontend development server"
	@echo "  run-dev       - Instructions for running both"
	@echo ""
	@echo "Build & Test:"
	@echo "  build         - Build both backend and frontend"
	@echo "  test          - Run all tests"
	@echo "  lint          - Run all linters"
	@echo "  clean         - Clean all build artifacts"
	@echo ""
	@echo "Docker Operations:"
	@echo "  docker-up     - Start full stack with Docker"
	@echo "  docker-down   - Stop all services"
	@echo "  docker-logs   - View logs"
	@echo "  start         - Alias for docker-up"
	@echo "  stop          - Alias for docker-down"
	@echo "  restart       - Restart all services"
	@echo ""
	@echo "Database:"
	@echo "  db-up         - Start only database services"
	@echo "  db-down       - Stop database services"