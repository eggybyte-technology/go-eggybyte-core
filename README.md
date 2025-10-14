# 🥚 EggyByte Core

**Enterprise-Grade Go Microservice Foundation Library**

[![Go Version](https://img.shields.io/badge/Go-1.25.1+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![Go Reference](https://pkg.go.dev/badge/github.com/eggybyte-technology/go-eggybyte-core.svg)](https://pkg.go.dev/github.com/eggybyte-technology/go-eggybyte-core)
[![Test Coverage](https://img.shields.io/badge/Coverage-67.9%25-green.svg)](coverage.html)
[![Go Report Card](https://goreportcard.com/badge/github.com/eggybyte-technology/go-eggybyte-core)](https://goreportcard.com/report/github.com/eggybyte-technology/go-eggybyte-core)

---

## 📖 Overview

**EggyByte Core** is a powerful, production-ready Go foundation library designed for building scalable microservices with minimal boilerplate. Built on modern best practices and battle-tested patterns, it enables developers to focus on business logic while the framework handles infrastructure concerns.

### 🎯 Project Identity

**EggyByte Core** is a **pure Go library** that provides foundational infrastructure components for microservice development. It does **not** include CLI tools, code generation, or deployment configurations - those are handled by separate projects.

### 🔗 Related Projects

- **[ebcctl](https://github.com/eggybyte-technology/ebcctl)** - Standalone CLI tool for project management and code generation
- **[go-eggybyte-core](https://github.com/eggybyte-technology/go-eggybyte-core)** - This library (pure Go microservice foundation)

### 🎯 Philosophy

- **Zero Boilerplate**: Start your entire service with a single function call
- **Convention over Configuration**: Sensible defaults with flexibility when needed
- **Developer Experience First**: Intuitive APIs and comprehensive tooling
- **Production Ready**: Built-in observability, health checks, and graceful shutdown
- **Cloud Native**: Kubernetes-ready with modern deployment patterns

---

## ✨ Core Features

### 🚀 **Instant Bootstrap**
Start your entire microservice with one line of code. No configuration files, no complex setup—just pure productivity.

### 🔄 **Auto-Registration Magic**
Tables self-register and auto-migrate via `init()`. Write your models, import the package, and you're done.

### 📊 **Unified Monitoring**
Single port serves Prometheus metrics, health checks, liveness and readiness probes—Kubernetes-native from day one.

### 📝 **Structured Logging**
Context-aware logging with automatic request ID tracking. JSON or console output with configurable levels.

### 🗄️ **Database Integration**
Production-ready MySQL/TiDB support with connection pooling, transaction management, and GORM integration.

### 🛠️ **Extensible Architecture**
Modular design with clear interfaces for easy extension and customization. Built-in support for custom health checkers, initializers, and service implementations.

---

## 🚀 Quick Start

### Installation

```bash
go get github.com/eggybyte-technology/go-eggybyte-core
```

### The 2-Line Microservice

```go
package main

import (
    "github.com/eggybyte-technology/go-eggybyte-core/pkg/config"
    "github.com/eggybyte-technology/go-eggybyte-core/pkg/core"
    "github.com/eggybyte-technology/go-eggybyte-core/pkg/log"
)

func main() {
    cfg := &config.Config{}
    config.MustReadFromEnv(cfg)

    if err := core.Bootstrap(cfg); err != nil {
        log.Fatal("Bootstrap failed", log.Field{Key: "error", Value: err})
    }
}
```

**That's it!** Your service now has:

- ✅ Structured logging
- ✅ Request ID tracking
- ✅ Graceful shutdown
- ✅ Health checks (/healthz)
- ✅ Liveness probe (/livez)
- ✅ Readiness probe (/readyz)
- ✅ Prometheus metrics (/metrics)
- ✅ Database pooling
- ✅ Auto table migration

---

## 📦 Architecture

### Module Overview

```
go-eggybyte-core/
├── 🎯 pkg/core/         Bootstrap orchestrator & service lifecycle
├── ⚙️  pkg/config/      Environment-based configuration management
├── 📝 pkg/log/          Structured logging with context propagation
├── 🗄️  pkg/db/          Database with auto-registration & pooling
├── 🚀 pkg/service/      Service launcher & graceful shutdown
├── 📊 pkg/monitoring/   Unified metrics & health endpoints
├── 📊 pkg/metrics/      Prometheus metrics collection
├── 🏥 pkg/health/       Health check service implementation
└── 📚 docs/             Comprehensive documentation
```

### Bootstrap Flow

```mermaid
graph LR
    A[Load Config] --> B[Init Logging]
    B --> C[Setup Database]
    C --> D[Auto-Migrate Tables]
    D --> E[Register Services]
    E --> F[Start Monitoring]
    F --> G[Start Business Services]
    G --> H[Signal Handling]
    H --> I[Graceful Shutdown]
```

---

## 🗄️ Database with Auto-Registration

### Define Your Model

```go
package repositories

import (
    "context"
    "gorm.io/gorm"
    "github.com/eggybyte-technology/go-eggybyte-core/pkg/db"
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
import _ "myservice/internal/repositories" // Triggers auto-registration!

func main() {
    cfg := &config.Config{}
    config.MustReadFromEnv(cfg)

    core.Bootstrap(cfg) // Tables created automatically

    db := db.GetDB()
    var users []User
    db.Find(&users)
}
```

---

## ⚙️ Configuration

All configuration is managed through environment variables—no config files needed.

### Core Configuration

```bash
# Service Identity
SERVICE_NAME=user-service
ENVIRONMENT=production

# Network
PORT=8080
METRICS_PORT=9090

# Logging
LOG_LEVEL=info          # debug | info | warn | error | fatal
LOG_FORMAT=json         # json | console

# Database (Optional)
DATABASE_DSN=user:pass@tcp(localhost:3306)/mydb?charset=utf8mb4&parseTime=True
DATABASE_MAX_OPEN_CONNS=100
DATABASE_MAX_IDLE_CONNS=10

# Kubernetes Config Watch (Optional)
ENABLE_K8S_CONFIG_WATCH=false
K8S_NAMESPACE=default
K8S_CONFIGMAP_NAME=my-service-config
```

---

## 📊 Built-in Monitoring

### Unified Monitoring Server

All monitoring endpoints served on **port 9090** for Kubernetes compatibility:

| Endpoint | Purpose | Response |
|----------|---------|----------|
| `GET /metrics` | Prometheus metrics | Text format |
| `GET /healthz` | Combined health check | JSON status |
| `GET /livez` | Liveness probe | HTTP 200 |
| `GET /readyz` | Readiness probe | HTTP 200/503 |

### Health Check Response

```json
{
  "status": true,
  "checks": {
    "database": "OK",
    "external-api": "OK"
  }
}
```

### Prometheus Metrics

Auto-exposed metrics include:
- Go runtime (goroutines, memory, GC)
- HTTP request metrics (duration, status codes)
- Custom application metrics
- Database connection pool stats

---

## 📝 Logging

### Structured Logging

```go
import "github.com/eggybyte-technology/go-eggybyte-core/pkg/log"

log.Info("User created",
    log.Field{Key: "user_id", Value: userID},
    log.Field{Key: "email", Value: email},
)

log.Error("Payment failed",
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

// Request ID automatically included
log.InfoContext(ctx, "Processing request")
log.ErrorContext(ctx, "Request failed", log.Field{Key: "error", Value: err})
```

---

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

### Register with Bootstrap

```go
func main() {
    cfg := &config.Config{}
    config.MustReadFromEnv(cfg)

    httpServer := NewHTTPServer(cfg.Port)
    grpcServer := NewGRPCServer(9090)

    core.Bootstrap(cfg, httpServer, grpcServer)
}
```

---

## 🧪 Testing

### Unit Testing with Mock Database

```go
func TestUserRepository(t *testing.T) {
    db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
    require.NoError(t, err)

    repo := &UserRepository{}
    err = repo.InitTable(context.Background(), db)
    require.NoError(t, err)

    user := &User{Email: "test@example.com", Name: "Test User"}
    result := repo.db.Create(user)
    assert.NoError(t, result.Error)
    assert.NotZero(t, user.ID)
}
```

---

## 🔧 Advanced Usage

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

### Custom Initializers

```go
type CacheInitializer struct {
    redisAddr string
}

func (c *CacheInitializer) Init(ctx context.Context) error {
    log.Info("Initializing cache", log.Field{Key: "addr", Value: c.redisAddr})
    // Setup Redis connection
    return nil
}

// Register with launcher
launcher.AddInitializer(&CacheInitializer{redisAddr: "localhost:6379"})
```

---

## 📚 Documentation

### Official Resources

- **[Getting Started Guide](docs/GETTING_STARTED.md)** - Step-by-step tutorials
- **[API Reference](docs/API_REFERENCE.md)** - Complete API documentation
- **[Architecture Guide](docs/ARCHITECTURE.md)** - Design patterns and best practices
- **[Migration Guide](docs/MIGRATION.md)** - Upgrading from other frameworks

---

## 💡 Best Practices

1. **Always use context** - Pass context through all layers for tracing and cancellation
2. **Log with structured fields** - Use `log.Field` instead of string formatting
3. **Register repositories in init()** - Enable automatic table migration
4. **Keep methods under 50 lines** - Follow EggyByte code quality standards
5. **Document public APIs** - Write comprehensive English comments
6. **Follow Go best practices** - Use standard Go patterns and conventions for maintainable code

---

## 📋 Changelog

See [CHANGELOG.md](CHANGELOG.md) for version history and release notes.

---

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## 🌟 Why EggyByte Core?

### ⚡ **Lightning Fast**
From idea to production in minutes, not days. Minimal boilerplate with maximum functionality out of the box.

### 🛡️ **Battle Tested**
Used in production by EggyByte microservices. Proven reliability and performance at scale.

### 🎓 **Developer Friendly**
Intuitive APIs, comprehensive docs, and sensible defaults. Focus on features, not infrastructure.

---

<div align="center">

### Built with ❤️ by EggyByte Technology

**[⭐ Star us on GitHub](https://github.com/eggybyte-technology/go-eggybyte-core)** | **[📖 Read the Docs](docs/)** | **[🚀 Get Started](#-quick-start)**

</div>