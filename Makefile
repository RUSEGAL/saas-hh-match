.PHONY: help build up down logs ps dev clean test lint

help:
	@echo "Development:"
	@echo "  make dev          - Start development mode with hot reload"
	@echo "  make dev-down    - Stop development mode"
	@echo "  make dev-logs    - Show development logs"
	@echo ""
	@echo "Production:"
	@echo "  make build        - Build all Docker images"
	@echo "  make up          - Start production services"
	@echo "  make down        - Stop services"
	@echo "  make logs        - Show logs"
	@echo "  make ps          - Show running services"
	@echo ""
	@echo "Testing:"
	@echo "  make test        - Run all tests"
	@echo "  make test-api    - Run go-api tests"
	@echo "  make test-ai    - Run ai-service tests"
	@echo "  make test-bot    - Run tg-bot tests"
	@echo ""
	@echo "Code Quality:"
	@echo "  make lint        - Run linters"
	@echo "  make vet         - Run go vet"
	@echo "  make fmt         - Format code"
	@echo ""
	@echo "CI/CD:"
	@echo "  make ci          - Run CI locally"
	@echo "  make build-local - Build binaries locally"

build:
	docker-compose build

up:
	docker-compose up -d
	@echo "Services started. Check status with: make ps"

down:
	docker-compose down

logs:
	docker-compose logs -f

logs-api:
	docker-compose logs -f api

logs-bot:
	docker-compose logs -f tg-bot

logs-ai:
	docker-compose logs -f ai-service

ps:
	docker-compose ps

dev:
	docker-compose -f docker-compose.dev.yml up -d

dev-logs:
	docker-compose -f docker-compose.dev.yml logs -f

dev-down:
	docker-compose -f docker-compose.dev.yml down

clean:
	docker-compose down -v
	docker system prune -f

restart:
	docker-compose restart

restart-api:
	docker-compose restart api

scale-api:
	@echo "Usage: make scale-api N=5"
	docker-compose up -d --scale api=$(N)

test:
	cd go-api && go test -v ./...
	cd ai-service && go test -v ./...
	cd tg-bot && go test -v ./...

test-api:
	cd go-api && go test -v -race -coverprofile=coverage.out ./...

test-ai:
	cd ai-service && go test -v -race ./...

test-bot:
	cd tg-bot && go test -v -race ./...

lint:
	golangci-lint run ./...

vet:
	cd go-api && go vet ./...
	cd ai-service && go vet ./...
	cd tg-bot && go vet ./...

fmt:
	cd go-api && go fmt ./...
	cd ai-service && go fmt ./...
	cd tg-bot && go fmt ./...

ci: vet test
	@echo "CI passed!"

build-local:
	mkdir -p artifacts
	cd go-api && go build -ldflags="-s -w" -o ../artifacts/go-api ./cmd/api/main.go
	cd ai-service && go build -ldflags="-s -w" -o ../artifacts/ai-service ./cmd/main.go
	cd tg-bot && go build -ldflags="-s -w" -o ../artifacts/tg-bot ./cmd/main.go
	@echo "Binaries built in artifacts/"

# Docker buildx for multi-platform
buildx:
	docker buildx build \
		--platform linux/amd64,linux/arm64 \
		--tag app/go-api:latest \
		--push false \
		-f go-api/Dockerfile ./go-api
