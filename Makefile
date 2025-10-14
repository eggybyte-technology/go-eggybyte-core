# EggyByte Core Library Makefile
# Pure Go Library for Microservice Development

.PHONY: help test lint clean deps mod-tidy mod-verify check version info

# Default target
.DEFAULT_GOAL := help

# Variables
VERSION := $(shell git describe --tags --always --dirty)
COMMIT := $(shell git rev-parse HEAD)
DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Go parameters
GOCMD := go
GOCLEAN := $(GOCMD) clean
GOTEST := $(GOCMD) test
GOGET := $(GOCMD) get
GOMOD := $(GOCMD) mod
GOFMT := $(GOCMD) fmt
GOVET := $(GOCMD) vet

## help: Display this help message
help:
	@echo "🥚 EggyByte Core Library - Enterprise-Grade Go Microservice Foundation"
	@echo ""
	@echo "📋 Available targets:"
	@echo ""
	@echo "🧪 Testing & Quality:"
	@echo "  test               Run tests"
	@echo "  test-coverage      Run tests with coverage report"
	@echo "  lint               Run linters"
	@echo "  fmt                Format code"
	@echo "  vet                Run go vet"
	@echo "  check              Run all checks (test, lint, vet, fmt)"
	@echo ""
	@echo "📦 Dependencies & Modules:"
	@echo "  deps               Download dependencies"
	@echo "  deps-update        Update dependencies"
	@echo "  mod-tidy           Tidy go modules"
	@echo "  mod-verify         Verify go modules"
	@echo ""
	@echo "🔧 Utilities:"
	@echo "  clean              Clean build artifacts"
	@echo "  version            Show version information"
	@echo "  info               Show project information"

## test: Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v -race -coverprofile=coverage.out ./...

## test-coverage: Run tests with coverage report
test-coverage: test
	@echo "Generating coverage report..."
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

## lint: Run linters
lint:
	@echo "Running linters..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
		exit 1; \
	fi

## fmt: Format code
fmt:
	@echo "Formatting code..."
	$(GOFMT) ./...

## vet: Run go vet
vet:
	@echo "Running go vet..."
	$(GOVET) ./...

## mod-tidy: Tidy go modules
mod-tidy:
	@echo "Tidying go modules..."
	$(GOMOD) tidy

## mod-verify: Verify go modules
mod-verify:
	@echo "Verifying go modules..."
	$(GOMOD) verify

## clean: Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	$(GOCLEAN)
	@rm -f coverage.out coverage.html

## deps: Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download

## deps-update: Update dependencies
deps-update:
	@echo "Updating dependencies..."
	$(GOGET) -u ./...
	$(GOMOD) tidy

## check: Run all checks (test, lint, vet, fmt)
check: test lint vet fmt
	@echo "All checks passed!"

## version: Show version information
version:
	@echo "Version: $(VERSION)"
	@echo "Commit: $(COMMIT)"
	@echo "Date: $(DATE)"

## info: Show project information
info:
	@echo "🥚 EggyByte Core Library Information"
	@echo "==================================="
	@echo "Version: $(VERSION)"
	@echo "Commit: $(COMMIT)"
	@echo "Date: $(DATE)"
	@echo "Go Version: $(shell go version)"
	@echo "Module: $(shell go list -m)"
	@echo ""
	@echo "📁 Project Structure:"
	@echo "  pkg/        - Library packages"
	@echo "  docs/       - Documentation"
	@echo "  internal/   - Internal packages"
	@echo ""
	@echo "🔧 Available Packages:"
	@echo "  pkg/core/     - Bootstrap orchestrator & service lifecycle"
	@echo "  pkg/config/   - Environment-based configuration management"
	@echo "  pkg/log/      - Structured logging with context propagation"
	@echo "  pkg/db/       - Database with auto-registration & pooling"
	@echo "  pkg/service/  - Service launcher & graceful shutdown"
	@echo "  pkg/monitoring/ - Unified metrics & health endpoints"
	@echo "  pkg/metrics/  - Prometheus metrics collection"
	@echo "  pkg/health/   - Health check service implementation"
	@echo ""
	@echo "📊 Test Coverage:"
	@if [ -f "coverage.out" ]; then \
		echo "  Coverage: $$(go tool cover -func=coverage.out | grep total | awk '{print $$3}')"; \
	else \
		echo "  Run 'make test-coverage' to generate coverage report"; \
	fi