package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var (
	// initProjectCmd initializes a complete full-stack project
	initProjectCmd = &cobra.Command{
		Use:   "project <project-name>",
		Short: "Initialize a complete full-stack project",
		Long: `Initialize a complete full-stack project with backend and frontend.

Creates a new project directory containing:
  - backend/ - Backend microservices directory
    - services/auth/ - Authentication service
    - services/user/ - User management service
  - frontend/ - Flutter application
  - api/ - Shared API definitions (protobuf)
  - Makefile - Unified build management
  - docker-compose.yml - Local development setup
  - README.md - Complete project documentation

Example:
  ebcctl init project eggybyte-platform
  ebcctl init project my-app --module github.com/myorg/my-app`,
		Args: cobra.ExactArgs(1),
		RunE: runInitProject,
	}

	// projectModulePath is the base module path for the project
	projectModulePath string
)

// init registers flags for the init project command
func init() {
	initProjectCmd.Flags().StringVarP(&projectModulePath, "module", "m", "",
		"Base Go module path (default: github.com/eggybyte-technology/<project-name>)")

	// Add to parent init command
	initCmd.AddCommand(initProjectCmd)
}

// runInitProject executes the init project command to create a complete full-stack project.
func runInitProject(cmd *cobra.Command, args []string) error {
	projectName := args[0]

	// Validate project name
	if err := validateProjectName(projectName); err != nil {
		return err
	}

	// Set default module path if not provided
	if projectModulePath == "" {
		projectModulePath = fmt.Sprintf("github.com/eggybyte-technology/%s", projectName)
	}

	logInfo("Initializing full-stack project: %s", projectName)
	logDebug("Module path: %s", projectModulePath)

	// Create project structure
	if err := createFullProjectStructure(projectName); err != nil {
		return fmt.Errorf("failed to create project structure: %w", err)
	}

	// Generate backend services
	if err := generateBackendServices(projectName); err != nil {
		return fmt.Errorf("failed to generate backend services: %w", err)
	}

	// Generate frontend
	if err := generateFrontendApp(projectName); err != nil {
		return fmt.Errorf("failed to generate frontend: %w", err)
	}

	// Generate root project files
	if err := generateRootProjectFiles(projectName); err != nil {
		return fmt.Errorf("failed to generate project files: %w", err)
	}

	// Print success message
	printProjectSuccessMessage(projectName)

	return nil
}

// validateProjectName checks if the project name is valid.
func validateProjectName(name string) error {
	if name == "" {
		return fmt.Errorf("project name cannot be empty")
	}

	// Check for valid characters (lowercase, numbers, hyphens)
	for _, char := range name {
		if !((char >= 'a' && char <= 'z') ||
			(char >= '0' && char <= '9') ||
			char == '-') {
			return fmt.Errorf("project name must contain only lowercase letters, numbers, and hyphens")
		}
	}

	// Check if directory already exists
	if _, err := os.Stat(name); !os.IsNotExist(err) {
		return fmt.Errorf("directory '%s' already exists", name)
	}

	return nil
}

// createFullProjectStructure creates the complete project directory structure.
func createFullProjectStructure(projectName string) error {
	logInfo("Creating project structure...")

	dirs := []string{
		projectName,
		filepath.Join(projectName, "backend", "services", "auth", "cmd"),
		filepath.Join(projectName, "backend", "services", "auth", "internal", "handlers"),
		filepath.Join(projectName, "backend", "services", "auth", "internal", "services"),
		filepath.Join(projectName, "backend", "services", "auth", "internal", "repositories"),
		filepath.Join(projectName, "backend", "services", "user", "cmd"),
		filepath.Join(projectName, "backend", "services", "user", "internal", "handlers"),
		filepath.Join(projectName, "backend", "services", "user", "internal", "services"),
		filepath.Join(projectName, "backend", "services", "user", "internal", "repositories"),
		filepath.Join(projectName, "api", "proto", "common"),
		filepath.Join(projectName, "api", "proto", "auth"),
		filepath.Join(projectName, "api", "proto", "user"),
		filepath.Join(projectName, "scripts"),
	}

	for _, dir := range dirs {
		logDebug("Creating directory: %s", dir)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	logInfo("Project structure created")
	return nil
}

// generateBackendServices generates the backend microservices.
func generateBackendServices(projectName string) error {
	logInfo("Generating backend services...")

	services := []string{"auth", "user"}

	for _, svc := range services {
		logDebug("Generating %s service...", svc)

		// Generate go.mod
		if err := generateServiceGoMod(projectName, svc); err != nil {
			return err
		}

		// Generate main.go
		if err := generateServiceMain(projectName, svc); err != nil {
			return err
		}

		// Generate README
		if err := generateServiceREADME(projectName, svc); err != nil {
			return err
		}

		// Generate sample repository
		if err := generateSampleRepository(projectName, svc); err != nil {
			return err
		}
	}

	logInfo("Backend services generated")
	return nil
}

// generateServiceGoMod generates go.mod for a service.
func generateServiceGoMod(projectName, serviceName string) error {
	modPath := filepath.Join(projectName, "backend", "services", serviceName, "go.mod")

	// Calculate relative path from service to go-eggybyte-core
	// From: examples/demo-platform/backend/services/auth
	// To:   go-eggybyte-core
	// Path: ../../../../../

	content := fmt.Sprintf(`module %s/backend/services/%s

go 1.25.1

require (
	github.com/eggybyte-technology/go-eggybyte-core v1.0.0
)

// Local development - adjust path based on your directory structure
// Path is relative from backend/services/%s to go-eggybyte-core root
replace github.com/eggybyte-technology/go-eggybyte-core => ../../../../../
`, projectModulePath, serviceName, serviceName)

	return os.WriteFile(modPath, []byte(content), 0644)
}

// generateServiceMain generates main.go for a service.
func generateServiceMain(projectName, serviceName string) error {
	mainPath := filepath.Join(projectName, "backend", "services", serviceName, "cmd", "main.go")

	content := fmt.Sprintf(`package main

import (
	"github.com/eggybyte-technology/go-eggybyte-core/config"
	"github.com/eggybyte-technology/go-eggybyte-core/core"
	"github.com/eggybyte-technology/go-eggybyte-core/log"

	// Import repositories for auto-registration
	_ "%s/backend/services/%s/internal/repositories"
)

func main() {
	// Load configuration from environment
	cfg := &config.Config{}
	config.MustReadFromEnv(cfg)

	// Bootstrap service with core infrastructure
	if err := core.Bootstrap(cfg); err != nil {
		log.Fatal("Bootstrap failed", log.Field{Key: "error", Value: err})
	}
}
`, projectModulePath, serviceName)

	return os.WriteFile(mainPath, []byte(content), 0644)
}

// generateServiceREADME generates README for a service.
func generateServiceREADME(projectName, serviceName string) error {
	readmePath := filepath.Join(projectName, "backend", "services", serviceName, "README.md")

	name := strings.Title(serviceName)
	content := fmt.Sprintf(`# %s Service

%s microservice for %s platform.

## Development

### Run Locally

`+"```bash"+`
export SERVICE_NAME=%s-service
export PORT=8080
export METRICS_PORT=9090
export LOG_LEVEL=info
export LOG_FORMAT=console

go run cmd/main.go
`+"```"+`

### Build

`+"```bash"+`
go build -o bin/%s cmd/main.go
`+"```"+`

## Configuration

See project root README.md for complete configuration guide.
`, name, name, projectName, serviceName, serviceName)

	return os.WriteFile(readmePath, []byte(content), 0644)
}

// generateSampleRepository generates a sample repository for the service.
func generateSampleRepository(projectName, serviceName string) error {
	repoPath := filepath.Join(projectName, "backend", "services", serviceName,
		"internal", "repositories", serviceName+"_repository.go")

	var modelName, structName, tableName string
	switch serviceName {
	case "auth":
		modelName = "session"
		structName = "Session"
		tableName = "sessions"
	case "user":
		modelName = "user"
		structName = "User"
		tableName = "users"
	default:
		modelName = serviceName
		structName = strings.Title(serviceName)
		tableName = serviceName + "s"
	}

	content := fmt.Sprintf(`package repositories

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/eggybyte-technology/go-eggybyte-core/db"
)

// %s represents the %s data model.
type %s struct {
	ID        uint      `+"`gorm:\"primaryKey\"`"+`
	CreatedAt time.Time
	UpdatedAt time.Time
	// TODO: Add your model fields here
}

// %sRepository handles database operations for %s models.
type %sRepository struct {
	db *gorm.DB
}

// New%sRepository creates a new instance of %sRepository.
func New%sRepository() *%sRepository {
	return &%sRepository{}
}

// TableName returns the database table name for this repository.
func (r *%sRepository) TableName() string {
	return "%s"
}

// InitTable performs table creation and schema migration.
func (r *%sRepository) InitTable(ctx context.Context, database *gorm.DB) error {
	r.db = database
	return r.db.WithContext(ctx).AutoMigrate(&%s{})
}

// Create inserts a new %s record into the database.
func (r *%sRepository) Create(ctx context.Context, model *%s) error {
	return r.db.WithContext(ctx).Create(model).Error
}

// FindByID retrieves a %s by its ID.
func (r *%sRepository) FindByID(ctx context.Context, id uint) (*%s, error) {
	var model %s
	err := r.db.WithContext(ctx).First(&model, id).Error
	return &model, err
}

// Update modifies an existing %s record.
func (r *%sRepository) Update(ctx context.Context, model *%s) error {
	return r.db.WithContext(ctx).Save(model).Error
}

// Delete removes a %s record by ID.
func (r *%sRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&%s{}, id).Error
}

// init registers this repository for automatic table initialization.
func init() {
	db.RegisterRepository(New%sRepository())
}
`,
		// Model definition
		structName, modelName, structName,
		// Repository struct
		structName, modelName, structName,
		// Constructor
		structName, structName, structName, structName, structName,
		// TableName
		structName, tableName,
		// InitTable
		structName, structName,
		// Create
		structName, structName, structName,
		// FindByID
		structName, structName, structName, structName,
		// Update
		structName, structName, structName,
		// Delete
		structName, structName, structName,
		// init
		structName,
	)

	return os.WriteFile(repoPath, []byte(content), 0644)
}

// generateFrontendApp generates the Flutter frontend application.
func generateFrontendApp(projectName string) error {
	logInfo("Generating frontend application...")

	// Check if Flutter is available
	if err := exec.Command("flutter", "--version").Run(); err != nil {
		logError("Flutter not found, skipping frontend generation")
		logInfo("Install Flutter from: https://flutter.dev/docs/get-started/install")
		logInfo("You can generate frontend later with: ebcctl init frontend <name>")
		return nil
	}

	// Generate Flutter app in frontend directory
	frontendPath := filepath.Join(projectName, "frontend")
	appName := strings.ReplaceAll(projectName, "-", "_") + "_app"

	cmd := exec.Command("flutter", "create",
		"--org", "com.eggybyte",
		"--project-name", appName,
		"--platforms", "android,ios,web",
		frontendPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		logError("Flutter create failed, skipping frontend")
		return nil
	}

	logInfo("Frontend application generated")
	return nil
}

// generateRootProjectFiles generates root-level project files.
func generateRootProjectFiles(projectName string) error {
	logInfo("Generating root project files...")

	files := map[string]func(string) string{
		"README.md":          generateProjectREADME,
		"Makefile":           generateProjectMakefile,
		"docker-compose.yml": generateDockerCompose,
		".gitignore":         generateProjectGitignore,
	}

	for filename, generator := range files {
		content := generator(projectName)
		path := filepath.Join(projectName, filename)

		logDebug("Writing file: %s", filename)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write %s: %w", filename, err)
		}
	}

	logInfo("Root project files generated")
	return nil
}

// generateProjectREADME generates the project README.
func generateProjectREADME(projectName string) string {
	name := strings.Title(strings.ReplaceAll(projectName, "-", " "))
	return fmt.Sprintf(`# %s

Full-stack application built with EggyByte technology stack.

## Project Structure

- `+"`backend/`"+` - Backend microservices (Go)
  - `+"`services/auth/`"+` - Authentication service
  - `+"`services/user/`"+` - User management service
- `+"`frontend/`"+` - Flutter application (Android/iOS/Web)
- `+"`api/`"+` - Shared API definitions (Protocol Buffers)
- `+"`scripts/`"+` - Build and deployment scripts

## Prerequisites

### Backend
- Go 1.25.1+
- Docker and Docker Compose
- Make

### Frontend
- Flutter SDK 3.16.0+
- Dart 3.2.0+

## Quick Start

### 1. Start Backend Services

`+"```bash"+`
# Start all backend services with Docker Compose
make dev-up

# Or run individual services locally
cd backend/services/auth
go run cmd/main.go
`+"```"+`

### 2. Run Frontend

`+"```bash"+`
cd frontend
flutter pub get
flutter run
`+"```"+`

## Development

### Backend Development

Each service is an independent Go module:

`+"```bash"+`
cd backend/services/user
go run cmd/main.go
`+"```"+`

### Frontend Development

`+"```bash"+`
cd frontend
flutter run -d chrome  # Run on web
flutter run            # Run on mobile device/emulator
`+"```"+`

### Generate Repository Code

`+"```bash"+`
cd backend/services/user
ebcctl new repo <model-name>
`+"```"+`

## Build & Deploy

### Build All Services

`+"```bash"+`
make build-all
`+"```"+`

### Build Docker Images

`+"```bash"+`
make docker-build-all
`+"```"+`

### Deploy to Kubernetes

`+"```bash"+`
make deploy-dev
`+"```"+`

## Configuration

### Backend Services

Each service is configured via environment variables. See individual service README.md files.

Common variables:
- `+"`SERVICE_NAME`"+` - Service identifier
- `+"`PORT`"+` - HTTP server port
- `+"`DATABASE_DSN`"+` - Database connection string
- `+"`LOG_LEVEL`"+` - Logging level (debug, info, warn, error)

### Frontend

Configure API endpoints in `+"`frontend/lib/config/api_config.dart`"+`.

## Testing

### Backend Tests

`+"```bash"+`
make test-backend
`+"```"+`

### Frontend Tests

`+"```bash"+`
cd frontend
flutter test
`+"```"+`

## License

Copyright © 2025 EggyByte Technology
`, name)
}

// generateProjectMakefile generates the project Makefile.
func generateProjectMakefile(projectName string) string {
	return `# Makefile for ` + projectName + `

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
	docker build -t ` + projectName + `-auth:latest -f backend/services/auth/Dockerfile .
	docker build -t ` + projectName + `-user:latest -f backend/services/user/Dockerfile .

## clean: Clean build artifacts
clean:
	rm -rf backend/services/*/bin
	cd frontend && flutter clean
`
}

// generateDockerCompose generates docker-compose.yml.
func generateDockerCompose(projectName string) string {
	return `version: '3.8'

services:
  # TiDB Database
  tidb:
    image: pingcap/tidb:latest
    ports:
      - "4000:4000"
    command:
      - --store=mocktikv
      - --advertise-address=tidb
      - --log.level=info

  # Redis Cache
  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"

  # Auth Service
  auth:
    build:
      context: .
      dockerfile: backend/services/auth/Dockerfile
    ports:
      - "8081:8080"
      - "9091:9090"
    environment:
      - SERVICE_NAME=auth-service
      - PORT=8080
      - METRICS_PORT=9090
      - DATABASE_DSN=root@tcp(tidb:4000)/eggybyte?charset=utf8mb4&parseTime=True
      - LOG_LEVEL=info
    depends_on:
      - tidb
      - redis

  # User Service
  user:
    build:
      context: .
      dockerfile: backend/services/user/Dockerfile
    ports:
      - "8082:8080"
      - "9092:9090"
    environment:
      - SERVICE_NAME=user-service
      - PORT=8080
      - METRICS_PORT=9090
      - DATABASE_DSN=root@tcp(tidb:4000)/eggybyte?charset=utf8mb4&parseTime=True
      - LOG_LEVEL=info
    depends_on:
      - tidb
      - redis
`
}

// generateProjectGitignore generates .gitignore.
func generateProjectGitignore(projectName string) string {
	return `# Backend
backend/services/*/bin/
*.exe
*.exe~
*.dll
*.so
*.dylib
*.test
*.out

# Frontend
frontend/build/
frontend/.dart_tool/
frontend/.packages
frontend/.flutter-plugins
frontend/.flutter-plugins-dependencies

# IDE
.idea/
.vscode/
*.swp
*.swo
*~

# OS
.DS_Store

# Environment
.env
.env.local

# Logs
*.log
`
}

// printProjectSuccessMessage prints the success message for project creation.
func printProjectSuccessMessage(projectName string) {
	logSuccess("Full-stack project '%s' initialized successfully!", projectName)
	fmt.Println("\nProject structure:")
	fmt.Printf("  %s/\n", projectName)
	fmt.Println("  ├── backend/services/")
	fmt.Println("  │   ├── auth/      - Authentication service")
	fmt.Println("  │   └── user/      - User management service")
	fmt.Println("  ├── frontend/      - Flutter application")
	fmt.Println("  ├── api/           - API definitions")
	fmt.Println("  └── Makefile       - Build management")
	fmt.Println("\nNext steps:")
	fmt.Printf("  1. cd %s\n", projectName)
	fmt.Println("  2. Start services: make dev-up")
	fmt.Println("  3. Run frontend: cd frontend && flutter run")
	fmt.Println("\nFor more information, see README.md")
	fmt.Println()
}
