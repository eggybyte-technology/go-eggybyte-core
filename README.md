# go-eggybyte-core

A powerful Go foundation library for building EggyByte microservices with minimal boilerplate and maximum productivity.

## 🎯 Features

- **Single-Line Bootstrap**: Start your entire service with one function call
- **Automatic Repository Registration**: Tables self-register and auto-migrate via init()
- **Service Lifecycle Management**: Graceful startup and shutdown with signal handling
- **Built-in Observability**: Prometheus metrics and health checks out of the box
- **Structured Logging**: Context-aware logging with request ID tracking
- **TiDB/MySQL Support**: Production-ready database integration with connection pooling
- **Kubernetes-Ready**: Health probes and configuration watching
- **Zero Boilerplate**: Focus on business logic, not infrastructure code

## 📦 Installation

```bash
go get github.com/eggybyte-technology/go-eggybyte-core
```

## 🚀 Quick Start

### Minimal Service (2 Lines!)

```go
package main

import (
    "github.com/eggybyte-technology/go-eggybyte-core/config"
    "github.com/eggybyte-technology/go-eggybyte-core/core"
    "github.com/eggybyte-technology/go-eggybyte-core/log"
)

func main() {
    // Load configuration from environment
    cfg := &config.Config{}
    config.MustReadFromEnv(cfg)

    // Bootstrap entire service in one call
    if err := core.Bootstrap(cfg); err != nil {
        log.Fatal("Bootstrap failed", log.Field{Key: "error", Value: err})
    }
}
```

That's it! Your service now has:
- ✅ Structured logging
- ✅ Prometheus metrics on port 9090
- ✅ Health checks (/healthz, /livez, /readyz)
- ✅ Graceful shutdown on SIGTERM/SIGINT
- ✅ Database connection (if DSN provided)
- ✅ Automatic table migration

### With Business Services

```go
func main() {
    cfg := &config.Config{}
    config.MustReadFromEnv(cfg)

    // Create your business services
    httpServer := NewHTTPServer(cfg.Port)
    grpcServer := NewGRPCServer(9090)

    // Bootstrap with business services
    if err := core.Bootstrap(cfg, httpServer, grpcServer); err != nil {
        log.Fatal("Bootstrap failed", log.Field{Key: "error", Value: err})
    }
}
```

## 🗄️ Database with Auto-Registration

### Define Your Repository

```go
package repositories

import (
    "context"
    "gorm.io/gorm"
    "github.com/eggybyte-technology/go-eggybyte-core/db"
)

type User struct {
    ID    uint   `gorm:"primaryKey"`
    Email string `gorm:"uniqueIndex;not null"`
    Name  string
}

type UserRepository struct {
    db *gorm.DB
}

func (r *UserRepository) TableName() string {
    return "users"
}

func (r *UserRepository) InitTable(ctx context.Context, db *gorm.DB) error {
    r.db = db
    return db.WithContext(ctx).AutoMigrate(&User{})
}

// Magic: Auto-register on import!
func init() {
    db.RegisterRepository(&UserRepository{})
}
```

### Use Your Repository

```go
import _ "myservice/internal/repositories" // Import triggers init()

func main() {
    cfg := &config.Config{}
    config.MustReadFromEnv(cfg)

    // Bootstrap automatically creates your tables!
    core.Bootstrap(cfg)

    // Access database
    db := db.GetDB()
    var users []User
    db.Find(&users)
}
```

## 🔧 Configuration

All configuration via environment variables:

```bash
# Service Identity
SERVICE_NAME=user-service
ENVIRONMENT=production

# Network
PORT=8080
METRICS_PORT=9090

# Logging
LOG_LEVEL=info      # debug, info, warn, error, fatal
LOG_FORMAT=json     # json, console

# Database (Optional)
DATABASE_DSN=user:pass@tcp(localhost:4000)/mydb?charset=utf8mb4&parseTime=True
DATABASE_MAX_OPEN_CONNS=100
DATABASE_MAX_IDLE_CONNS=10

# Kubernetes Config Watching (Optional)
ENABLE_K8S_CONFIG_WATCH=false
K8S_NAMESPACE=default
K8S_CONFIGMAP_NAME=my-service-config
```

## 📊 Built-in Endpoints

### Metrics (Port 9090)
- `GET /metrics` - Prometheus metrics

### Health Checks (Port 9090)
- `GET /healthz` - Combined health check
- `GET /livez` - Liveness probe (always returns 200)
- `GET /readyz` - Readiness probe (checks dependencies)

## 🏗️ Architecture

### Module Overview

```
go-eggybyte-core/
├── config/     # Configuration management
├── log/        # Structured logging with context
├── db/         # Database with auto-registration
├── service/    # Service lifecycle management
├── metrics/    # Prometheus metrics
├── health/     # Health check endpoints
└── core/       # Bootstrap orchestrator
```

### Design Patterns

**Registry Pattern**: Repositories self-register via init()
```go
func init() {
    db.RegisterRepository(&MyRepo{})
}
```

**Dependency Injection**: Components injected through launcher
```go
launcher.AddInitializer(dbInit)
launcher.AddService(httpServer)
```

**Graceful Shutdown**: Signal handling with timeout
```go
// Automatically handles SIGINT and SIGTERM
```

**Context Propagation**: Thread-safe context passing
```go
logger := log.FromContext(ctx)
requestID := log.GetRequestID(ctx)
```

## 📝 Logging Examples

### Basic Logging

```go
import "github.com/eggybyte-technology/go-eggybyte-core/log"

log.Info("User created",
    log.Field{Key: "user_id", Value: userID},
    log.Field{Key: "email", Value: email},
)

log.Error("Failed to process payment",
    log.Field{Key: "order_id", Value: orderID},
    log.Field{Key: "error", Value: err.Error()},
)
```

### Context-Aware Logging

```go
// Attach logger to context
ctx, logger := log.WithLogger(ctx, "",
    log.Field{Key: "user_id", Value: userID},
)

// Use throughout request lifecycle
log.InfoContext(ctx, "Processing request")
log.ErrorContext(ctx, "Request failed", log.Field{Key: "error", Value: err})

// Request ID automatically included in all logs
```

## 🎯 Service Implementation

### Implement Service Interface

```go
type HTTPServer struct {
    port   int
    server *http.Server
}

func (s *HTTPServer) Start(ctx context.Context) error {
    s.server = &http.Server{Addr: fmt.Sprintf(":%d", s.port)}

    errCh := make(chan error, 1)
    go func() {
        errCh <- s.server.ListenAndServe()
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
```

## 🧪 Testing

### Unit Testing with Mock DB

```go
func TestUserRepository(t *testing.T) {
    // Setup test database
    db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
    require.NoError(t, err)

    // Initialize repository
    repo := &UserRepository{}
    err = repo.InitTable(context.Background(), db)
    require.NoError(t, err)

    // Test operations
    user := &User{Email: "test@example.com", Name: "Test User"}
    result := repo.db.Create(user)
    assert.NoError(t, result.Error)
}
```

## 🔍 Best Practices

1. **Always use context**: Pass context through all layers
2. **Log with fields**: Use structured logging, not string formatting
3. **Register repositories in init()**: Enable automatic migration
4. **Keep methods under 50 lines**: Follow code quality standards
5. **Document public APIs**: Write comprehensive English comments
6. **Use Bootstrap**: Let the core handle infrastructure setup

## 🛠️ Advanced Usage

### Custom Initializers

```go
type CacheInitializer struct {
    redisAddr string
}

func (c *CacheInitializer) Init(ctx context.Context) error {
    // Setup Redis connection
    log.Info("Initializing cache", log.Field{Key: "addr", Value: c.redisAddr})
    // ...
    return nil
}

// Register with bootstrap
func main() {
    cfg := &config.Config{}
    config.MustReadFromEnv(cfg)

    launcher := service.NewLauncher()
    launcher.AddInitializer(&CacheInitializer{redisAddr: "localhost:6379"})

    // Or use core.Bootstrap and add initializers after
}
```

### Custom Health Checkers

```go
type DatabaseHealthChecker struct {
    db *gorm.DB
}

func (d *DatabaseHealthChecker) Name() string {
    return "database"
}

func (d *DatabaseHealthChecker) Check(ctx context.Context) error {
    sqlDB, err := d.db.DB()
    if err != nil {
        return err
    }
    return sqlDB.PingContext(ctx)
}

// Register with health service
healthService.AddChecker(&DatabaseHealthChecker{db: db.GetDB()})
```

## 📄 License

Copyright © 2025 EggyByte Technology. All rights reserved.

## 🤝 Contributing

1. Follow EggyByte code quality standards
2. All public APIs must have English comments
3. Methods must be under 50 lines
4. Run `go test ./...` before submitting
5. Ensure `go build ./...` succeeds

## 📞 Support

For issues and questions:
- GitHub Issues: [github.com/eggybyte-technology/go-eggybyte-core/issues](https://github.com/eggybyte-technology/go-eggybyte-core/issues)
- Documentation: See `/docs` directory

---

Built with ❤️ by EggyByte Technology

