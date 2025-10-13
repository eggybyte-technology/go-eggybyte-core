# Quick Start Guide - go-eggybyte-core

Get your first EggyByte microservice running in 5 minutes!

## Prerequisites

- Go 1.24.5 or higher
- Basic understanding of Go programming

## Step 1: Install ebcctl CLI Tool

```bash
# Clone the repository
git clone https://github.com/eggybyte-technology/go-eggybyte-core.git
cd go-eggybyte-core

# Build the CLI tool
go build -o ~/bin/ebcctl ./cmd/ebcctl

# Verify installation
ebcctl --help
```

## Step 2: Create Your First Service

```bash
# Create a new service project
ebcctl init my-service

# Navigate to the project
cd my-service
```

**What you get:**
- ✅ Complete project structure
- ✅ go.mod with core dependencies
- ✅ main.go with Bootstrap integration
- ✅ README and documentation
- ✅ Dockerfile for containerization

## Step 3: Generate a Repository

```bash
# Generate repository for a User model
ebcctl new repo user
```

**What happens:**
- ✅ Creates `internal/repositories/user_repository.go`
- ✅ Includes auto-registration via `init()`
- ✅ Implements CRUD operations
- ✅ Follows EggyByte standards

## Step 4: Update Repository Model

Edit `internal/repositories/user_repository.go`:

```go
// Update the User struct with your fields
type User struct {
	ID        uint      `gorm:"primaryKey"`
	Email     string    `gorm:"uniqueIndex;not null"`
	Name      string    `gorm:"not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
```

## Step 5: Import Repository in main.go

The generated repository is already set up for auto-registration. Just import it:

```go
package main

import (
	"github.com/eggybyte-technology/go-eggybyte-core/config"
	"github.com/eggybyte-technology/go-eggybyte-core/core"
	"github.com/eggybyte-technology/go-eggybyte-core/log"

	// This import triggers auto-registration via init()
	_ "github.com/<your-org>/my-service/internal/repositories"
)

func main() {
	cfg := &config.Config{}
	config.MustReadFromEnv(cfg)

	if err := core.Bootstrap(cfg); err != nil {
		log.Fatal("Bootstrap failed", log.Field{Key: "error", Value: err})
	}
}
```

## Step 6: Configure Environment

Create a `.env` file or export variables:

```bash
export SERVICE_NAME=my-service
export ENVIRONMENT=development
export PORT=8080
export METRICS_PORT=9090
export LOG_LEVEL=info
export LOG_FORMAT=console

# Optional: Database connection
export DATABASE_DSN="user:pass@tcp(localhost:3306)/mydb?charset=utf8mb4&parseTime=True"
export DATABASE_MAX_OPEN_CONNS=100
export DATABASE_MAX_IDLE_CONNS=10
```

## Step 7: Run Your Service

```bash
# Run directly
go run cmd/main.go

# Or build first
go build -o bin/my-service cmd/main.go
./bin/my-service
```

**What you'll see:**
```
{"level":"info","timestamp":"2025-01-15T10:30:00.000Z","msg":"Starting service bootstrap","service":"my-service"}
{"level":"info","timestamp":"2025-01-15T10:30:00.100Z","msg":"Database connection established"}
{"level":"info","timestamp":"2025-01-15T10:30:00.200Z","msg":"Table initialized successfully","table":"users"}
{"level":"info","timestamp":"2025-01-15T10:30:00.300Z","msg":"Starting metrics server","port":9090}
{"level":"info","timestamp":"2025-01-15T10:30:00.400Z","msg":"Starting health server","port":9090}
```

## Step 8: Check Health and Metrics

```bash
# Health check
curl http://localhost:9090/healthz

# Liveness probe
curl http://localhost:9090/livez

# Readiness probe
curl http://localhost:9090/readyz

# Prometheus metrics
curl http://localhost:9090/metrics
```

## Step 9: Use Your Repository

Create a simple HTTP handler to use the repository:

```go
// internal/handlers/user_handler.go
package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/<your-org>/my-service/internal/repositories"
	"github.com/eggybyte-technology/go-eggybyte-core/log"
)

type UserHandler struct {
	repo repositories.UserRepositoryInterface
}

func NewUserHandler(repo repositories.UserRepositoryInterface) *UserHandler {
	return &UserHandler{repo: repo}
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var user repositories.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.repo.Create(r.Context(), &user); err != nil {
		log.Error("Failed to create user", log.Field{Key: "error", Value: err})
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}
```

## Step 10: Add HTTP Server (Optional)

Update main.go to include an HTTP server:

```go
package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/eggybyte-technology/go-eggybyte-core/config"
	"github.com/eggybyte-technology/go-eggybyte-core/core"
	"github.com/eggybyte-technology/go-eggybyte-core/log"
	"github.com/eggybyte-technology/go-eggybyte-core/service"
	"github.com/eggybyte-technology/go-eggybyte-core/db"

	"github.com/<your-org>/my-service/internal/handlers"
	"github.com/<your-org>/my-service/internal/repositories"
	_ "github.com/<your-org>/my-service/internal/repositories"
)

// HTTPServer implements service.Service for HTTP API
type HTTPServer struct {
	port   int
	server *http.Server
}

func (s *HTTPServer) Start(ctx context.Context) error {
	mux := http.NewServeMux()
	
	// Setup routes
	userRepo := repositories.NewUserRepository()
	userHandler := handlers.NewUserHandler(userRepo)
	mux.HandleFunc("/users", userHandler.CreateUser)

	s.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.port),
		Handler: mux,
	}

	log.Info("Starting HTTP server", log.Field{Key: "port", Value: s.port})

	errCh := make(chan error, 1)
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		return s.Stop(context.Background())
	}
}

func (s *HTTPServer) Stop(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

func main() {
	cfg := &config.Config{}
	config.MustReadFromEnv(cfg)

	// Create HTTP server
	httpServer := &HTTPServer{port: cfg.Port}

	// Bootstrap with HTTP server
	if err := core.Bootstrap(cfg, httpServer); err != nil {
		log.Fatal("Bootstrap failed", log.Field{Key: "error", Value: err})
	}
}
```

## 🎉 Congratulations!

You now have a fully functional microservice with:

- ✅ **Auto-registering repositories** - Tables created automatically
- ✅ **Prometheus metrics** - Available at :9090/metrics
- ✅ **Health checks** - Kubernetes-ready probes
- ✅ **Structured logging** - JSON formatted logs
- ✅ **Graceful shutdown** - SIGTERM/SIGINT handling
- ✅ **Database integration** - TiDB/MySQL with connection pooling

## Next Steps

1. **Add more repositories**: `ebcctl new repo order`, `ebcctl new repo product`
2. **Implement business logic**: Create services and handlers
3. **Add tests**: Use Go's built-in testing
4. **Containerize**: `docker build -t my-service .`
5. **Deploy**: Use provided Kubernetes manifests

## Learn More

- [README.md](README.md) - Complete documentation
- [examples/user-service](examples/user-service) - Working example
- [API Documentation](#) - Auto-generated docs

## Getting Help

- GitHub Issues: Report bugs and request features
- Documentation: See `/docs` directory
- Examples: Check `examples/` for working code

---

**Happy Building! 🚀**

