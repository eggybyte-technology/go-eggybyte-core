# Makefile for demo-platform

.PHONY: help dev-up dev-down build-all test-all

## help: Display this help message
help:
	@echo "Available targets:"
	@grep -E '^## [a-zA-Z_-]+:' $(MAKEFILE_LIST) | sed 's/## /  /'

## dev-up: Start all services with Docker Compose
dev-up:
	docker-compose up -d

## dev-down: Stop all services
dev-down:
	docker-compose down

## build-all: Build all backend services
build-all:
	@echo "Building auth service..."
	cd backend/services/auth && go build -o bin/auth cmd/main.go
	@echo "Building user service..."
	cd backend/services/user && go build -o bin/user cmd/main.go

## test-backend: Run backend tests
test-backend:
	cd backend/services/auth && go test ./...
	cd backend/services/user && go test ./...

## test-frontend: Run frontend tests
test-frontend:
	cd frontend && flutter test

## test-all: Run all tests
test-all: test-backend test-frontend

## docker-build-all: Build all Docker images
docker-build-all:
	docker build -t demo-platform-auth:latest -f backend/services/auth/Dockerfile .
	docker build -t demo-platform-user:latest -f backend/services/user/Dockerfile .

## clean: Clean build artifacts
clean:
	rm -rf backend/services/*/bin
	cd frontend && flutter clean
