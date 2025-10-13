# EggyByte Core Makefile
# Enterprise-Grade Go Microservice Foundation

.PHONY: help build test lint clean install release deps security examples

# Default target
.DEFAULT_GOAL := help

# Variables
BINARY_NAME := ebcctl
VERSION := $(shell git describe --tags --always --dirty)
COMMIT := $(shell git rev-parse HEAD)
DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -ldflags "-s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"

# Go parameters
GOCMD := go
GOBUILD := $(GOCMD) build
GOCLEAN := $(GOCMD) clean
GOTEST := $(GOCMD) test
GOGET := $(GOCMD) get
GOMOD := $(GOCMD) mod
GOFMT := $(GOCMD) fmt
GOVET := $(GOCMD) vet

# Directories
CMD_DIR := ./cmd/ebcctl
BIN_DIR := ./bin
DIST_DIR := ./dist
EXAMPLES_DIR := ./examples

## help: Display this help message
help:
	@echo "🥚 EggyByte Core - Enterprise-Grade Go Microservice Foundation"
	@echo ""
	@echo "📋 Available targets:"
	@echo ""
	@echo "🔨 Build & Development:"
	@echo "  build              Build the ebcctl binary"
	@echo "  build-all          Build binaries for all platforms"
	@echo "  dev                Development mode with live reload"
	@echo "  install            Install the binary to GOPATH/bin"
	@echo "  clean              Clean build artifacts"
	@echo ""
	@echo "🧪 Testing & Quality:"
	@echo "  test               Run tests"
	@echo "  test-coverage      Run tests with coverage report"
	@echo "  lint               Run linters"
	@echo "  fmt                Format code"
	@echo "  vet                Run go vet"
	@echo "  check              Run all checks (test, lint, vet, fmt)"
	@echo "  ci                 Run CI pipeline locally"
	@echo ""
	@echo "📦 Dependencies & Modules:"
	@echo "  deps               Download dependencies"
	@echo "  deps-update        Update dependencies"
	@echo "  mod-tidy           Tidy go modules"
	@echo "  mod-verify         Verify go modules"
	@echo ""
	@echo "🔒 Security & Scanning:"
	@echo "  security           Run security checks"
	@echo ""
	@echo "🚀 Release & Deployment:"
	@echo "  release            Create a release (requires goreleaser)"
	@echo "  release-snapshot   Create a snapshot release"
	@echo "  github-update      Update GitHub repository with current changes"
	@echo "  github-release     Create and push a new tag to GitHub"
	@echo "  github-release-local Create a local release without pushing"
	@echo "  prepare-release    Prepare project for release (usage: make prepare-release VERSION=v1.0.0)"
	@echo "  create-release     Create complete GitHub release with binaries (usage: make create-release VERSION=v1.0.0)"
	@echo "  create-release-force Force create release even if tag exists (usage: make create-release-force VERSION=v1.0.0)"
	@echo ""
	@echo "📚 Examples & Documentation:"
	@echo "  examples           Run all example applications"
	@echo "  examples-build     Build all example applications"
	@echo "  examples-clean     Clean example build artifacts"
	@echo "  docs-serve         Serve documentation locally"
	@echo "  docs-build         Build documentation"
	@echo ""
	@echo "ℹ️  Info & Utilities:"
	@echo "  version            Show version information"
	@echo "  info               Show project information"

## build: Build the ebcctl binary
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BIN_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BIN_DIR)/$(BINARY_NAME) $(CMD_DIR)

## build-all: Build binaries for all platforms
build-all:
	@echo "Building $(BINARY_NAME) for all platforms..."
	@mkdir -p $(DIST_DIR)
	@GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-linux-amd64 $(CMD_DIR)
	@GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-linux-arm64 $(CMD_DIR)
	@GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-darwin-amd64 $(CMD_DIR)
	@GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-darwin-arm64 $(CMD_DIR)
	@GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-windows-amd64.exe $(CMD_DIR)
	@echo "Binaries built in $(DIST_DIR)/"

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
clean: examples-clean
	@echo "Cleaning build artifacts..."
	$(GOCLEAN)
	@rm -rf $(BIN_DIR)
	@rm -rf $(DIST_DIR)
	@rm -f coverage.out coverage.html
	@rm -rf site/

## install: Install the binary to GOPATH/bin
install: build
	@echo "Installing $(BINARY_NAME)..."
	@cp $(BIN_DIR)/$(BINARY_NAME) $(GOPATH)/bin/

## dev: Development mode with live reload
dev:
	@echo "Starting development mode..."
	@if command -v air >/dev/null 2>&1; then \
		air; \
	else \
		echo "air not installed. Install with: go install github.com/cosmtrek/air@latest"; \
		echo "Running make build instead..."; \
		make build; \
	fi

## release: Create a release (requires goreleaser)
release:
	@echo "Creating release..."
	@if command -v goreleaser >/dev/null 2>&1; then \
		goreleaser release --clean; \
	else \
		echo "goreleaser not installed. Install with: go install github.com/goreleaser/goreleaser@latest"; \
		exit 1; \
	fi

## release-snapshot: Create a snapshot release
release-snapshot:
	@echo "Creating snapshot release..."
	@if command -v goreleaser >/dev/null 2>&1; then \
		goreleaser release --snapshot --clean; \
	else \
		echo "goreleaser not installed. Install with: go install github.com/goreleaser/goreleaser@latest"; \
		exit 1; \
	fi

## check: Run all checks (test, lint, vet, fmt)
check: test lint vet fmt
	@echo "All checks passed!"

## ci: Run CI pipeline locally
ci: mod-tidy mod-verify check build
	@echo "CI pipeline completed successfully!"

## version: Show version information
version:
	@echo "Version: $(VERSION)"
	@echo "Commit: $(COMMIT)"
	@echo "Date: $(DATE)"

## deps: Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download

## deps-update: Update dependencies
deps-update:
	@echo "Updating dependencies..."
	$(GOGET) -u ./...
	$(GOMOD) tidy

## security: Run security checks
security:
	@echo "Running security checks..."
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "gosec not installed. Install with: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest"; \
		exit 1; \
	fi

## github-update: Update GitHub repository with current changes
github-update:
	@echo "Updating GitHub repository..."
	@scripts/sh/github-update.sh

## github-release: Create and push a new tag to GitHub
github-release:
	@echo "Creating GitHub release..."
	@if [ -z "$(VERSION)" ]; then \
		echo "Error: VERSION variable is required"; \
		echo "Usage: make github-release VERSION=v1.0.0"; \
		exit 1; \
	fi
	@scripts/sh/github-release.sh -v $(VERSION)

## github-release-local: Create a local release without pushing
github-release-local:
	@echo "Creating local release..."
	@if [ -z "$(VERSION)" ]; then \
		echo "Error: VERSION variable is required"; \
		echo "Usage: make github-release-local VERSION=v1.0.0"; \
		exit 1; \
	fi
	@if command -v goreleaser >/dev/null 2>&1; then \
		goreleaser release --snapshot --clean; \
	else \
		echo "goreleaser not installed. Install with: go install github.com/goreleaser/goreleaser@latest"; \
		exit 1; \
	fi

## prepare-release: Prepare project for release (usage: make prepare-release VERSION=v1.0.0)
prepare-release:
	@if [ -z "$(VERSION)" ]; then \
		echo "Error: VERSION variable is required"; \
		echo "Usage: make prepare-release VERSION=v1.0.0"; \
		exit 1; \
	fi
	@scripts/sh/prepare-release.sh $(VERSION)

## create-release: Create complete GitHub release with binaries (usage: make create-release VERSION=v1.0.0)
create-release:
	@if [ -z "$(VERSION)" ]; then \
		echo "Error: VERSION variable is required"; \
		echo "Usage: make create-release VERSION=v1.0.0"; \
		exit 1; \
	fi
	@scripts/sh/create-release.sh -v $(VERSION)

## create-release-force: Force create release even if tag exists (usage: make create-release-force VERSION=v1.0.0)
create-release-force:
	@if [ -z "$(VERSION)" ]; then \
		echo "Error: VERSION variable is required"; \
		echo "Usage: make create-release-force VERSION=v1.0.0"; \
		exit 1; \
	fi
	@scripts/sh/create-release.sh -v $(VERSION) -f

## examples: Run all example applications
examples:
	@echo "Running example applications..."
	@for example in $(EXAMPLES_DIR)/*; do \
		if [ -d "$$example" ] && [ -f "$$example/go.mod" ]; then \
			echo "Running example: $$(basename $$example)"; \
			cd $$example && go run . && cd -; \
		fi; \
	done

## examples-build: Build all example applications
examples-build:
	@echo "Building example applications..."
	@for example in $(EXAMPLES_DIR)/*; do \
		if [ -d "$$example" ] && [ -f "$$example/go.mod" ]; then \
			echo "Building example: $$(basename $$example)"; \
			cd $$example && go build -o bin/example . && cd -; \
		fi; \
	done

## examples-clean: Clean example build artifacts
examples-clean:
	@echo "Cleaning example build artifacts..."
	@for example in $(EXAMPLES_DIR)/*; do \
		if [ -d "$$example" ]; then \
			rm -rf $$example/bin; \
		fi; \
	done

## docs-serve: Serve documentation locally
docs-serve:
	@echo "Serving documentation..."
	@if command -v mkdocs >/dev/null 2>&1; then \
		mkdocs serve; \
	else \
		echo "mkdocs not installed. Install with: pip install mkdocs"; \
		echo "Serving README.md with python http server instead..."; \
		python3 -m http.server 8000; \
	fi

## docs-build: Build documentation
docs-build:
	@echo "Building documentation..."
	@if command -v mkdocs >/dev/null 2>&1; then \
		mkdocs build; \
	else \
		echo "mkdocs not installed. Install with: pip install mkdocs"; \
		echo "Documentation available in docs/ directory"; \
	fi

## info: Show project information
info:
	@echo "🥚 EggyByte Core Project Information"
	@echo "====================================="
	@echo "Version: $(VERSION)"
	@echo "Commit: $(COMMIT)"
	@echo "Date: $(DATE)"
	@echo "Binary: $(BINARY_NAME)"
	@echo "Go Version: $(shell go version)"
	@echo "Module: $(shell go list -m)"
	@echo ""
	@echo "📁 Project Structure:"
	@echo "  cmd/        - Command-line applications"
	@echo "  pkg/        - Library packages"
	@echo "  examples/   - Example applications"
	@echo "  docs/       - Documentation"
	@echo "  testdata/   - Test data and fixtures"
	@echo "  .github/    - GitHub workflows and templates"
	@echo ""
	@echo "🔧 Available Tools:"
	@echo "  ebcctl      - CLI tool for code generation"
	@echo ""
	@echo "📊 Test Coverage:"
	@if [ -f "coverage.out" ]; then \
		echo "  Coverage: $$(go tool cover -func=coverage.out | grep total | awk '{print $$3}')"; \
	else \
		echo "  Run 'make test-coverage' to generate coverage report"; \
	fi
