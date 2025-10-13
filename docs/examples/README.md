# go-eggybyte-core Examples

This directory contains complete example projects demonstrating how to use go-eggybyte-core to build production-ready microservices.

## Available Examples

### 1. Task Service

A full-featured task management microservice showcasing:

- ✅ **Single-line bootstrap** - Start complete service with `core.Bootstrap()`
- ✅ **Auto-registration pattern** - Repositories self-register via `init()`
- ✅ **RESTful HTTP API** - Complete CRUD operations
- ✅ **Structured logging** - Context-aware logging with request IDs
- ✅ **Database integration** - MySQL/TiDB with automatic migrations
- ✅ **Prometheus metrics** - Built-in observability
- ✅ **Health checks** - Kubernetes-ready probes
- ✅ **Docker deployment** - Production-ready containerization
- ✅ **API testing** - Automated test scripts

**Location**: [`task-service/`](./task-service/)

**Quick Start**:

```bash
cd task-service

# Start with Docker Compose
docker-compose up

# Or run locally (requires MySQL)
source .env
go run cmd/main.go
```

**What you'll learn**:

1. How to use `ebcctl` to scaffold a new project
2. How to implement the auto-registration pattern
3. How to build a RESTful API with structured logging
4. How to integrate with databases using GORM
5. How to deploy with Docker and docker-compose

## Using These Examples

### Prerequisites

- Go 1.25.1+
- Docker & Docker Compose (for easiest setup)
- MySQL 8.0+ or TiDB (for local development without Docker)
- `ebcctl` CLI tool (optional, for code generation)

### Install ebcctl

```bash
go install github.com/eggybyte-technology/go-eggybyte-core/cmd/ebcctl@latest
```

### Running an Example

Each example is a complete, runnable project:

```bash
# Navigate to the example
cd task-service

# Option 1: Use Docker Compose (easiest)
docker-compose up

# Option 2: Run locally
source .env  # Load environment variables
go run cmd/main.go
```

### Testing the API

Each example includes test scripts:

```bash
# Task Service API tests
cd task-service
./test-api.sh
```

## Project Structure Pattern

All examples follow the standard go-eggybyte-core structure:

```
example-service/
├── cmd/
│   └── main.go                    # Service entry point with Bootstrap
├── internal/
│   ├── repositories/
│   │   └── *_repository.go       # Auto-registered repositories
│   └── services/
│       ├── *_service.go          # Business logic
│       └── http_server.go        # HTTP/gRPC servers
├── go.mod                         # Go module with core dependency
├── Dockerfile                     # Container image
├── docker-compose.yml            # Local development stack
├── .env                          # Environment configuration
└── test-api.sh                   # API test script
```

## Key Patterns Demonstrated

### 1. Bootstrap Pattern

All infrastructure initialized in one line:

```go
func main() {
    cfg := &config.Config{}
    config.MustReadFromEnv(cfg)
    
    httpServer := services.NewHTTPServer(cfg.Port)
    
    // Everything handled automatically
    core.Bootstrap(cfg, httpServer)
}
```

### 2. Auto-Registration Pattern

Repositories self-register via `init()`:

```go
package repositories

func init() {
    db.RegisterRepository(NewTaskRepository())
}
```

Just import the package and tables are auto-migrated!

### 3. Service Interface Pattern

Services implement simple `Start/Stop` interface:

```go
type HTTPServer struct { ... }

func (s *HTTPServer) Start(ctx context.Context) error { ... }
func (s *HTTPServer) Stop(ctx context.Context) error { ... }
```

Graceful shutdown handled automatically.

### 4. Structured Logging Pattern

Context-aware logging with automatic request IDs:

```go
log.InfoContext(ctx, "Task created", 
    log.Field{Key: "task_id", Value: id},
    log.Field{Key: "status", Value: status})
```

## Creating Your Own Service

Use `ebcctl` to scaffold a new service following these patterns:

```bash
# Initialize new service
ebcctl init my-service --module github.com/mycompany/my-service

cd my-service

# Generate repository
ebcctl new repo user

# Update main.go to import repositories
# Implement your business logic
# Run!
go run cmd/main.go
```

## Common Tasks

### Add a New Repository

```bash
ebcctl new repo order
```

Then import it in `cmd/main.go`:

```go
import _ "your-module/internal/repositories"
```

### Add Environment Variable

1. Update `.env` file
2. Update `ENV.md` documentation
3. Access via `cfg.YourVariable` in code

### Add API Endpoint

1. Add method to `*_service.go`
2. Add handler in `http_server.go`
3. Register route in `Start()` method

### Run with Different Database

Update `DATABASE_DSN` in `.env`:

```bash
# MySQL
DATABASE_DSN=user:pass@tcp(localhost:3306)/dbname?charset=utf8mb4&parseTime=True

# TiDB
DATABASE_DSN=user:pass@tcp(tidb-host:4000)/dbname?charset=utf8mb4&parseTime=True
```

## Troubleshooting

### "Failed to connect to database"

Ensure MySQL/TiDB is running:

```bash
docker run -d --name mysql -e MYSQL_ROOT_PASSWORD=rootpass -p 3306:3306 mysql:8.0
```

### "Port already in use"

Change ports in `.env`:

```bash
PORT=8081
METRICS_PORT=9091
```

### "Module not found"

Run `go mod tidy` in the service directory:

```bash
cd task-service
go mod tidy
```

## Resources

- [go-eggybyte-core Documentation](../README.md)
- [Task Service README](task-service/README.md)
- [ebcctl Command Reference](../README.md#-cli-tool-ebcctl)

## Contributing

Found an issue or have an improvement? Please open an issue or submit a pull request!

---

**Built with go-eggybyte-core** - Zero boilerplate, maximum productivity.

