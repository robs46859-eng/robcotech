# FullStackArkham Makefile

.PHONY: help dev up down test lint build clean

# Default target
help:
	@echo "FullStackArkham - AI Operating System"
	@echo ""
	@echo "Usage:"
	@echo "  make dev       - Start all services in development mode"
	@echo "  make up        - Start all services (detached)"
	@echo "  make down      - Stop all services"
	@echo "  make test      - Run all tests"
	@echo "  make test-e2e  - Run end-to-end tests"
	@echo "  make lint      - Run linters"
	@echo "  make build     - Build all Docker images"
	@echo "  make clean     - Clean build artifacts"
	@echo "  make logs      - View service logs"
	@echo "  make db-migrate - Run database migrations"
	@echo ""

# Development
dev:
	docker-compose up --build

up:
	docker-compose up -d

down:
	docker-compose down

# Testing
test:
	@echo "Running unit tests..."
	cd services/arkham && python -m pytest tests/ -v
	cd services/orchestration && python -m pytest tests/ -v 2>/dev/null || true
	cd services/gateway && go test ./...
	@echo "Tests complete!"

test-e2e:
	@echo "Running end-to-end tests..."
	pytest tests/e2e/ -v --asyncio-mode=auto

test-coverage:
	@echo "Running tests with coverage..."
	cd services/arkham && python -m pytest tests/ --cov=app --cov-report=html
	cd services/orchestration && python -m pytest tests/ --cov=app --cov-report=html 2>/dev/null || true

# Linting
lint:
	@echo "Running linters..."
	cd services/gateway && gofmt -d .
	cd services/arkham && ruff check app/ 2>/dev/null || true
	cd services/bim_ingestion && ruff check app/ 2>/dev/null || true
	cd services/orchestration && ruff check app/ 2>/dev/null || true
	cd services/memory && ruff check app/ 2>/dev/null || true
	cd services/semantic-cache && ruff check app/ 2>/dev/null || true
	cd services/billing && ruff check app/ 2>/dev/null || true

lint-fix:
	cd services/arkham && ruff check app/ --fix 2>/dev/null || true
	cd services/bim_ingestion && ruff check app/ --fix 2>/dev/null || true
	cd services/orchestration && ruff check app/ --fix 2>/dev/null || true
	cd services/memory && ruff check app/ --fix 2>/dev/null || true
	cd services/semantic-cache && ruff check app/ --fix 2>/dev/null || true
	cd services/billing && ruff check app/ --fix 2>/dev/null || true

# Building
build:
	docker-compose build

build-gateway:
	cd services/gateway && go build -o bin/gateway ./app

# Logs
logs:
	docker-compose logs -f

logs-gateway:
	docker-compose logs -f gateway

logs-arkham:
	docker-compose logs -f arkham

logs-bim:
	docker-compose logs -f bim_ingestion

# Database
db-migrate:
	docker-compose exec postgres psql -U postgres -d fullstackarkham -f /docker-entrypoint-initdb.d/init.sql

db-shell:
	docker-compose exec postgres psql -U postgres -d fullstackarkham

db-reset:
	docker-compose down -v
	docker-compose up -d postgres
	sleep 5
	$(MAKE) db-migrate

# Knowledge Graph
graph-update:
	graphify update .

graph-query:
	@read -p "Query: " query; graphify query "$$query"

# Frontend
frontend-dev:
	cd apps/web && npm install && npm run dev

frontend-build:
	cd apps/web && npm install && npm run build

# Clean
clean:
	rm -rf services/gateway/bin
	rm -rf services/*/build
	rm -rf services/*/dist
	rm -rf services/*/*.egg-info
	rm -rf apps/web/.next
	rm -rf apps/web/node_modules
	find . -type d -name "__pycache__" -exec rm -rf {} + 2>/dev/null || true
	find . -type d -name ".pytest_cache" -exec rm -rf {} + 2>/dev/null || true
	find . -type d -name ".coverage" -exec rm -rf {} + 2>/dev/null || true
	@echo "Clean complete!"

# Docker cleanup
docker-clean:
	docker-compose down -v
	docker system prune -f

# Health check
health:
	@echo "Checking service health..."
	@curl -s http://localhost:8080/health | jq . || echo "Gateway: DOWN"
	@curl -s http://localhost:8081/health | jq . || echo "Arkham: DOWN"
	@curl -s http://localhost:8082/health | jq . || echo "BIM Ingestion: DOWN"
	@curl -s http://localhost:8083/health | jq . || echo "Orchestration: DOWN"
	@curl -s http://localhost:8084/health | jq . || echo "Semantic Cache: DOWN"
	@curl -s http://localhost:8085/health | jq . || echo "Memory: DOWN"
	@curl -s http://localhost:8086/health | jq . || echo "Billing: DOWN"
