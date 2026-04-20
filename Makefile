# ============================================
# Trello Agent — Makefile
# ============================================

.PHONY: help dev up down logs build test migrate clean

# Default target
help:
	@echo "Trello Agent - Available commands:"
	@echo ""
	@echo "  Development:"
	@echo "    make dev          - Start development environment (docker compose up)"
	@echo "    make up           - Start all services in background"
	@echo "    make down         - Stop all services"
	@echo "    make logs         - View logs (all services)"
	@echo "    make logs-api     - View API logs only"
	@echo "    make restart      - Restart all services"
	@echo ""
	@echo "  Database:"
	@echo "    make migrate      - Run database migrations"
	@echo "    make migrate-new  - Create new migration (NAME=migration_name)"
	@echo "    make migrate-down - Rollback last migration"
	@echo "    make db-shell     - Open PostgreSQL shell"
	@echo "    make redis-cli    - Open Redis CLI"
	@echo ""
	@echo "  Build & Test:"
	@echo "    make build        - Build Go binary"
	@echo "    make test         - Run all tests"
	@echo "    make test-cover   - Run tests with coverage"
	@echo "    make lint         - Run linter"
	@echo ""
	@echo "  Docker:"
	@echo "    make docker-build - Build Docker image"
	@echo "    make docker-push  - Push to registry"
	@echo "    make clean        - Remove containers and volumes"
	@echo ""

# ==========================================
# Development
# ==========================================

dev:
	docker compose up

up:
	docker compose up -d

down:
	docker compose down

logs:
	docker compose logs -f

logs-api:
	docker compose logs -f api

restart:
	docker compose restart

# ==========================================
# Database
# ==========================================

migrate:
	docker compose exec api /app/server migrate up

migrate-new:
	@if [ -z "$(NAME)" ]; then echo "Usage: make migrate-new NAME=migration_name"; exit 1; fi
	goose -dir migrations create $(NAME) sql

migrate-down:
	docker compose exec api /app/server migrate down

db-shell:
	docker compose exec postgres psql -U trello_agent -d trello_agent

redis-cli:
	docker compose exec redis redis-cli -a redis_secret

# ==========================================
# Build & Test
# ==========================================

build:
	go build -o bin/server ./cmd/server

test:
	go test -v ./...

test-cover:
	go test -cover -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

test-race:
	go test -race ./...

lint:
	golangci-lint run ./...

fmt:
	go fmt ./...
	goimports -w .

# ==========================================
# Docker
# ==========================================

VERSION ?= latest
REGISTRY ?= ghcr.io/your-org

docker-build:
	docker build -t trello-agent-api:$(VERSION) .

docker-push:
	docker tag trello-agent-api:$(VERSION) $(REGISTRY)/trello-agent-api:$(VERSION)
	docker push $(REGISTRY)/trello-agent-api:$(VERSION)

docker-prod:
	docker compose -f docker-compose.prod.yml up -d

# ==========================================
# Cleanup
# ==========================================

clean:
	docker compose down -v --remove-orphans
	rm -rf bin/ coverage.out coverage.html

clean-all: clean
	docker system prune -af
	docker volume prune -f

# ==========================================
# Utilities
# ==========================================

# Generate API docs
docs:
	swag init -g cmd/server/main.go -o docs/swagger

# Generate sqlc code
sqlc:
	sqlc generate

# Install development tools
tools:
	go install github.com/swaggo/swag/cmd/swag@latest
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	go install github.com/pressly/goose/v3/cmd/goose@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest

# Watch mode for development (requires air)
watch:
	air -c .air.toml
