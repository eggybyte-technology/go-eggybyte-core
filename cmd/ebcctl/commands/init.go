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
	// initCmd is the parent command for initialization
	initCmd = &cobra.Command{
		Use:   "init",
		Short: "Initialize a new project, service, or frontend",
		Long: `Initialize various types of projects:
  - backend: Create a Go microservice
  - frontend: Create a Flutter application
  - project: Create a complete full-stack project

Use 'ebcctl init <type> <name>' to create a new project.`,
	}

	// initBackendCmd initializes a new backend microservice
	initBackendCmd = &cobra.Command{
		Use:   "backend <service-name>",
		Short: "Initialize a new backend microservice",
		Long: `Initialize a new Go microservice with complete structure.

Creates a new directory with the service name containing:
  - Standard project structure (cmd/, internal/, etc.)
  - go.mod with core dependencies (local replace for development)
  - main.go with Bootstrap integration
  - README.md with documentation
  - ENV.md with configuration guide
  - Dockerfile for containerization
  - .gitignore with Go best practices

Example:
  ebcctl init backend user-service
  ebcctl init backend payment-service --module github.com/myorg/payment`,
		Args: cobra.ExactArgs(1),
		RunE: runInitBackend,
	}

	// modulePath is the Go module path for the new service
	modulePath string

	// goVersion is the Go version to use
	goVersion string
)

// init registers flags for the init commands
func init() {
	// Backend command flags
	initBackendCmd.Flags().StringVarP(&modulePath, "module", "m", "",
		"Go module path (default: github.com/eggybyte-technology/<service-name>)")
	initBackendCmd.Flags().StringVar(&goVersion, "go-version", "1.25.1",
		"Go version to use in go.mod")

	// Add subcommands to init
	initCmd.AddCommand(initBackendCmd)
}

// runInitBackend executes the init backend command to create a new service project.
func runInitBackend(cmd *cobra.Command, args []string) error {
	serviceName := args[0]

	// Validate service name
	if err := validateServiceName(serviceName); err != nil {
		return err
	}

	// Set default module path if not provided
	if modulePath == "" {
		modulePath = fmt.Sprintf("github.com/eggybyte-technology/%s", serviceName)
	}

	logInfo("Initializing new service: %s", serviceName)
	logDebug("Module path: %s", modulePath)
	logDebug("Go version: %s", goVersion)

	// Create project structure
	if err := createProjectStructure(serviceName); err != nil {
		return fmt.Errorf("failed to create project structure: %w", err)
	}

	// Generate project files
	if err := generateProjectFiles(serviceName); err != nil {
		return fmt.Errorf("failed to generate project files: %w", err)
	}

	// Run go mod tidy
	if err := runGoModTidy(serviceName); err != nil {
		logError("Failed to run go mod tidy: %v", err)
		logInfo("You can run 'go mod tidy' manually in the project directory")
	}

	// Print success message with next steps
	printSuccessMessage(serviceName)

	return nil
}

// validateServiceName checks if the service name is valid.
func validateServiceName(name string) error {
	if name == "" {
		return fmt.Errorf("service name cannot be empty")
	}

	// Check for valid characters (lowercase, numbers, hyphens)
	for _, char := range name {
		if !((char >= 'a' && char <= 'z') ||
			(char >= '0' && char <= '9') ||
			char == '-') {
			return fmt.Errorf("service name must contain only lowercase letters, numbers, and hyphens")
		}
	}

	// Check if directory already exists
	if _, err := os.Stat(name); !os.IsNotExist(err) {
		return fmt.Errorf("directory '%s' already exists", name)
	}

	return nil
}

// createProjectStructure creates the standard directory structure.
func createProjectStructure(serviceName string) error {
	logInfo("Creating project structure...")

	dirs := []string{
		serviceName,
		filepath.Join(serviceName, "cmd"),
		filepath.Join(serviceName, "internal", "handlers"),
		filepath.Join(serviceName, "internal", "services"),
		filepath.Join(serviceName, "internal", "repositories"),
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

// generateProjectFiles generates all project files from templates.
func generateProjectFiles(serviceName string) error {
	logInfo("Generating project files...")

	files := map[string]func(string) string{
		"go.mod":                     generateGoMod,
		"cmd/main.go":                generateMainGo,
		"README.md":                  generateREADME,
		"Dockerfile":                 generateDockerfile,
		"ENV.md":                     generateENVDoc,
		".gitignore":                 generateGitignore,
		"internal/handlers/.gitkeep": func(string) string { return "" },
	}

	for filename, generator := range files {
		content := generator(serviceName)
		path := filepath.Join(serviceName, filename)

		logDebug("Writing file: %s", filename)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write %s: %w", filename, err)
		}
	}

	logInfo("Project files generated")
	return nil
}

// runGoModTidy runs go mod tidy in the project directory.
func runGoModTidy(serviceName string) error {
	logInfo("Running go mod tidy...")

	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = serviceName
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// printSuccessMessage prints the success message with next steps.
func printSuccessMessage(serviceName string) {
	logSuccess("Service '%s' initialized successfully!", serviceName)
	fmt.Println("\nNext steps:")
	fmt.Printf("  1. cd %s\n", serviceName)
	fmt.Println("  2. Set environment variables (see ENV.md)")
	fmt.Println("  3. Implement your business logic in internal/")
	fmt.Println("  4. Run your service: go run cmd/main.go")
	fmt.Println("\nGenerate repository code:")
	fmt.Printf("  ebcctl new repo <model-name>\n\n")
}

// Template generators (simplified versions)

func generateGoMod(serviceName string) string {
	return fmt.Sprintf(`module %s

go %s

require (
	github.com/eggybyte-technology/go-eggybyte-core v0.1.0
)

// Local development - adjust path to point to go-eggybyte-core
// Example: if go-eggybyte-core is in parent directory, use ../go-eggybyte-core
replace github.com/eggybyte-technology/go-eggybyte-core => ../go-eggybyte-core
`, modulePath, goVersion)
}

func generateMainGo(serviceName string) string {
	return `package main

import (
	"github.com/eggybyte-technology/go-eggybyte-core/config"
	"github.com/eggybyte-technology/go-eggybyte-core/core"
	"github.com/eggybyte-technology/go-eggybyte-core/log"
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
`
}

func generateREADME(serviceName string) string {
	name := strings.Title(strings.ReplaceAll(serviceName, "-", " "))
	return fmt.Sprintf(`# %s

## Overview

%s microservice built with go-eggybyte-core.

## Getting Started

### Prerequisites

- Go 1.24.5+
- Docker (optional)
- Kubernetes (optional)

### Configuration

See ENV.md for required environment variables.

### Run Locally

`+"```bash"+`
export SERVICE_NAME=%s
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

## Development

### Project Structure

- cmd/main.go - Service entry point
- internal/handlers/ - HTTP/gRPC handlers
- internal/services/ - Business logic
- internal/repositories/ - Data access layer

### Generate Repository

`+"```bash"+`
ebcctl new repo <model-name>
`+"```"+`

## License

Copyright © 2025 EggyByte Technology
`, name, name, serviceName, serviceName)
}

func generateDockerfile(serviceName string) string {
	return `FROM golang:1.24.5-alpine AS builder
WORKDIR /workspace
COPY . .
RUN go build -o /app/service cmd/main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/service .
EXPOSE 8080 9090
CMD ["./service"]
`
}

func generateENVDoc(serviceName string) string {
	return fmt.Sprintf(`# Environment Variables

## Required

- `+"`SERVICE_NAME`"+` - Service identifier (default: %s)
- `+"`PORT`"+` - HTTP server port (default: 8080)

## Optional

- `+"`METRICS_PORT`"+` - Metrics server port (default: 9090)
- `+"`LOG_LEVEL`"+` - Log level: debug, info, warn, error (default: info)
- `+"`LOG_FORMAT`"+` - Log format: json, console (default: json)
- `+"`DATABASE_DSN`"+` - Database connection string
- `+"`ENVIRONMENT`"+` - Environment: dev, staging, production
`, serviceName)
}

func generateGitignore(serviceName string) string {
	return `# Binaries
bin/
*.exe
*.exe~
*.dll
*.so
*.dylib

# Test binary
*.test

# Output of the go coverage tool
*.out

# Dependency directories
vendor/

# Go workspace file
go.work

# IDE
.idea/
.vscode/
*.swp
*.swo
*~

# OS
.DS_Store
`
}
