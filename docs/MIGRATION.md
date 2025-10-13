# Migration Guide

## Overview

This guide helps you migrate from other Go microservice frameworks to EggyByte Core. It covers common migration scenarios and provides step-by-step instructions.

## Migration from Gin/Echo/Fiber

### Before (Gin Example)

```go
package main

import (
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
)

func main() {
    r := gin.Default()
    
    // Database setup
    db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
    if err != nil {
        panic(err)
    }
    
    // Routes
    r.GET("/users", getUsers)
    r.POST("/users", createUser)
    
    r.Run(":8080")
}

func getUsers(c *gin.Context) {
    // Handler logic
}

func createUser(c *gin.Context) {
    // Handler logic
}
```

### After (EggyByte Core)

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
    
    httpServer := NewHTTPServer(cfg.Port)
    
    if err := core.Bootstrap(cfg, httpServer); err != nil {
        log.Fatal("Bootstrap failed", log.Field{Key: "error", Value: err})
    }
}

type HTTPServer struct {
    port   int
    server *http.Server
}

func (s *HTTPServer) Start(ctx context.Context) error {
    // Gin/Echo/Fiber setup and routes
    return s.server.ListenAndServe()
}

func (s *HTTPServer) Stop(ctx context.Context) error {
    return s.server.Shutdown(ctx)
}
```

### Migration Steps

1. **Replace Framework Imports**
   ```go
   // Remove
   import "github.com/gin-gonic/gin"
   
   // Add
   import "github.com/eggybyte-technology/go-eggybyte-core/pkg/core"
   ```

2. **Update Main Function**
   - Replace framework initialization with `core.Bootstrap()`
   - Move route setup to service implementation
   - Add graceful shutdown handling

3. **Implement Service Interface**
   - Create service struct implementing `Service` interface
   - Move HTTP server setup to `Start()` method
   - Add graceful shutdown to `Stop()` method

4. **Update Configuration**
   - Replace hardcoded values with environment variables
   - Use `config.Config` structure
   - Load configuration with `config.MustReadFromEnv()`

## Migration from Custom Frameworks

### Before (Custom Framework)

```go
package main

import (
    "log"
    "net/http"
    "os"
)

func main() {
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }
    
    // Custom initialization
    initDatabase()
    initCache()
    initLogging()
    
    // Start HTTP server
    http.HandleFunc("/health", healthHandler)
    http.HandleFunc("/users", usersHandler)
    
    log.Printf("Server starting on port %s", port)
    log.Fatal(http.ListenAndServe(":"+port, nil))
}

func initDatabase() {
    // Custom database initialization
}

func initCache() {
    // Custom cache initialization
}

func initLogging() {
    // Custom logging initialization
}
```

### After (EggyByte Core)

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
    
    httpServer := NewHTTPServer(cfg.Port)
    
    if err := core.Bootstrap(cfg, httpServer); err != nil {
        log.Fatal("Bootstrap failed", log.Field{Key: "error", Value: err})
    }
}

type HTTPServer struct {
    port   int
    server *http.Server
}

func (s *HTTPServer) Start(ctx context.Context) error {
    mux := http.NewServeMux()
    mux.HandleFunc("/health", healthHandler)
    mux.HandleFunc("/users", usersHandler)
    
    s.server = &http.Server{
        Addr:    fmt.Sprintf(":%d", s.port),
        Handler: mux,
    }
    
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
    if s.server != nil {
        return s.server.Shutdown(ctx)
    }
    return nil
}
```

### Migration Steps

1. **Remove Custom Initialization**
   - Delete custom `initDatabase()`, `initCache()`, `initLogging()` functions
   - Remove manual configuration loading
   - Remove custom error handling

2. **Add EggyByte Core**
   - Import EggyByte Core packages
   - Use `config.Config` for configuration
   - Use `core.Bootstrap()` for initialization

3. **Implement Service Interface**
   - Create service struct
   - Implement `Start()` and `Stop()` methods
   - Move HTTP server setup to service

4. **Update Environment Variables**
   - Replace hardcoded values with environment variables
   - Use standard EggyByte Core configuration names
   - Set up proper configuration structure

## Migration from Microservice Frameworks

### Before (Go-Kit Example)

```go
package main

import (
    "github.com/go-kit/kit/log"
    "github.com/go-kit/kit/log/level"
    "github.com/go-kit/kit/transport/http"
)

func main() {
    logger := log.NewJSONLogger(os.Stdout)
    logger = log.With(logger, "ts", log.DefaultTimestampUTC)
    logger = log.With(logger, "caller", log.DefaultCaller)
    
    // Service setup
    var svc UserService
    svc = userService{}
    svc = loggingMiddleware{logger, svc}
    
    // HTTP transport
    userHandler := http.NewServer(
        makeUserEndpoint(svc),
        decodeUserRequest,
        encodeResponse,
    )
    
    http.Handle("/users", userHandler)
    
    logger.Log("msg", "HTTP", "addr", ":8080")
    logger.Log("err", http.ListenAndServe(":8080", nil))
}
```

### After (EggyByte Core)

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
    
    userService := NewUserService()
    httpServer := NewHTTPServer(cfg.Port, userService)
    
    if err := core.Bootstrap(cfg, httpServer); err != nil {
        log.Fatal("Bootstrap failed", log.Field{Key: "error", Value: err})
    }
}

type UserService struct {
    // Service implementation
}

type HTTPServer struct {
    port        int
    userService *UserService
    server      *http.Server
}

func (s *HTTPServer) Start(ctx context.Context) error {
    mux := http.NewServeMux()
    mux.HandleFunc("/users", s.handleUsers)
    
    s.server = &http.Server{
        Addr:    fmt.Sprintf(":%d", s.port),
        Handler: mux,
    }
    
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
    if s.server != nil {
        return s.server.Shutdown(ctx)
    }
    return nil
}
```

### Migration Steps

1. **Replace Logging**
   - Remove Go-Kit logging
   - Use EggyByte Core structured logging
   - Update log calls to use `log.Info()`, `log.Error()`, etc.

2. **Simplify Service Structure**
   - Remove Go-Kit service interfaces
   - Implement direct service structs
   - Remove middleware complexity

3. **Update HTTP Handlers**
   - Replace Go-Kit HTTP transport
   - Use standard `http.HandlerFunc`
   - Implement direct request/response handling

4. **Add Bootstrap**
   - Use `core.Bootstrap()` for initialization
   - Implement `Service` interface
   - Add graceful shutdown

## Database Migration

### Before (Manual GORM Setup)

```go
package main

import (
    "gorm.io/driver/mysql"
    "gorm.io/gorm"
)

func main() {
    // Manual database setup
    db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
    if err != nil {
        panic(err)
    }
    
    // Manual table creation
    err = db.AutoMigrate(&User{}, &Order{})
    if err != nil {
        panic(err)
    }
    
    // Manual connection pool setup
    sqlDB, err := db.DB()
    if err != nil {
        panic(err)
    }
    sqlDB.SetMaxOpenConns(100)
    sqlDB.SetMaxIdleConns(10)
}
```

### After (EggyByte Core)

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
    
    // Database DSN from environment
    cfg.DatabaseDSN = os.Getenv("DATABASE_DSN")
    
    if err := core.Bootstrap(cfg); err != nil {
        log.Fatal("Bootstrap failed", log.Field{Key: "error", Value: err})
    }
}

// Repository with auto-registration
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

func init() {
    db.RegisterRepository(&UserRepository{})
}
```

### Migration Steps

1. **Remove Manual Database Setup**
   - Delete manual `gorm.Open()` calls
   - Remove manual table migration
   - Remove manual connection pool configuration

2. **Add Repository Pattern**
   - Create repository structs implementing `db.Repository`
   - Move table definitions to repositories
   - Add auto-registration in `init()` functions

3. **Update Configuration**
   - Add `DATABASE_DSN` environment variable
   - Use `config.Config` for database settings
   - Let EggyByte Core handle connection pooling

4. **Update Database Usage**
   - Use `db.GetDB()` to get database connection
   - Update repository methods to use context
   - Remove manual database management

## Cache Migration

### Before (Manual Redis Setup)

```go
package main

import (
    "github.com/go-redis/redis/v8"
)

func main() {
    // Manual Redis setup
    rdb := redis.NewClient(&redis.Options{
        Addr:     "localhost:6379",
        Password: "",
        DB:       0,
    })
    
    // Manual connection test
    ctx := context.Background()
    err := rdb.Ping(ctx).Err()
    if err != nil {
        panic(err)
    }
}

func getFromCache(key string) (string, error) {
    val, err := rdb.Get(ctx, key).Result()
    if err == redis.Nil {
        return "", nil
    }
    return val, err
}
```

### After (EggyByte Core)

```go
package main

import (
    "github.com/eggybyte-technology/go-eggybyte-core/pkg/cache"
    "github.com/eggybyte-technology/go-eggybyte-core/pkg/config"
    "github.com/eggybyte-technology/go-eggybyte-core/pkg/core"
)

func main() {
    cfg := &config.Config{}
    config.MustReadFromEnv(cfg)
    
    // Cache servers from environment
    cfg.CacheServers = []string{"localhost:11211"}
    
    if err := core.Bootstrap(cfg); err != nil {
        log.Fatal("Bootstrap failed", log.Field{Key: "error", Value: err})
    }
}

func getFromCache(key string) ([]byte, error) {
    cacheClient := cache.GetClient()
    cacheService := cache.NewCacheService(cacheClient)
    
    return cacheService.Get(context.Background(), key)
}
```

### Migration Steps

1. **Replace Redis with Memcached**
   - Remove Redis client setup
   - Add Memcached configuration
   - Update cache operations

2. **Update Cache Operations**
   - Replace Redis-specific methods
   - Use EggyByte Core cache service
   - Update data types (string to []byte)

3. **Update Configuration**
   - Add `CACHE_SERVERS` environment variable
   - Use `config.Config` for cache settings
   - Let EggyByte Core handle connection management

## Configuration Migration

### Before (Config Files)

```go
package main

import (
    "gopkg.in/yaml.v2"
    "io/ioutil"
)

type Config struct {
    Server struct {
        Port int `yaml:"port"`
    } `yaml:"server"`
    Database struct {
        DSN string `yaml:"dsn"`
    } `yaml:"database"`
}

func loadConfig() *Config {
    data, err := ioutil.ReadFile("config.yaml")
    if err != nil {
        panic(err)
    }
    
    var config Config
    err = yaml.Unmarshal(data, &config)
    if err != nil {
        panic(err)
    }
    
    return &config
}
```

### After (EggyByte Core)

```go
package main

import (
    "github.com/eggybyte-technology/go-eggybyte-core/pkg/config"
)

func main() {
    cfg := &config.Config{}
    config.MustReadFromEnv(cfg)
    
    // Configuration loaded from environment variables
    // No config files needed
}
```

### Migration Steps

1. **Remove Config Files**
   - Delete YAML/JSON config files
   - Remove config file parsing code
   - Remove file-based configuration

2. **Add Environment Variables**
   - Set up environment variables
   - Use standard EggyByte Core config names
   - Update deployment configurations

3. **Update Configuration Structure**
   - Use `config.Config` struct
   - Remove custom config structs
   - Use `config.MustReadFromEnv()` for loading

## Monitoring Migration

### Before (Manual Prometheus Setup)

```go
package main

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
    // Manual Prometheus setup
    http.Handle("/metrics", promhttp.Handler())
    
    // Manual metrics
    var (
        httpRequestsTotal = prometheus.NewCounterVec(
            prometheus.CounterOpts{
                Name: "http_requests_total",
                Help: "Total number of HTTP requests",
            },
            []string{"method", "endpoint"},
        )
    )
    
    prometheus.MustRegister(httpRequestsTotal)
}
```

### After (EggyByte Core)

```go
package main

import (
    "github.com/eggybyte-technology/go-eggybyte-core/pkg/config"
    "github.com/eggybyte-technology/go-eggybyte-core/pkg/core"
)

func main() {
    cfg := &config.Config{}
    config.MustReadFromEnv(cfg)
    
    // Monitoring automatically included
    if err := core.Bootstrap(cfg); err != nil {
        log.Fatal("Bootstrap failed", log.Field{Key: "error", Value: err})
    }
}
```

### Migration Steps

1. **Remove Manual Prometheus Setup**
   - Delete manual Prometheus configuration
   - Remove manual metrics registration
   - Remove manual HTTP handlers

2. **Add EggyByte Core Monitoring**
   - Use built-in monitoring service
   - Access metrics at `/metrics` endpoint
   - Use health checks at `/healthz`, `/readyz`

3. **Update Metrics**
   - Use EggyByte Core metrics patterns
   - Update metric names and labels
   - Remove custom metric definitions

## Testing Migration

### Before (Manual Test Setup)

```go
func TestUserService(t *testing.T) {
    // Manual test setup
    db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
    require.NoError(t, err)
    
    // Manual table creation
    err = db.AutoMigrate(&User{})
    require.NoError(t, err)
    
    // Manual service setup
    service := NewUserService(db)
    
    // Test logic
    user, err := service.CreateUser("test@example.com")
    assert.NoError(t, err)
    assert.NotNil(t, user)
}
```

### After (EggyByte Core)

```go
func TestUserService(t *testing.T) {
    // Use EggyByte Core test utilities
    cfg := &config.Config{
        DatabaseDSN: "sqlite://:memory:",
        LogLevel:    "silent",
    }
    
    // Bootstrap with test configuration
    err := core.Bootstrap(cfg)
    require.NoError(t, err)
    
    // Get database connection
    db := db.GetDB()
    require.NotNil(t, db)
    
    // Test logic
    userRepo := repositories.NewUserRepository(db)
    user := &repositories.User{
        Email: "test@example.com",
    }
    
    err = userRepo.Create(context.Background(), user)
    assert.NoError(t, err)
    assert.NotZero(t, user.ID)
}
```

### Migration Steps

1. **Update Test Setup**
   - Use EggyByte Core bootstrap for tests
   - Use `config.Config` for test configuration
   - Remove manual database setup

2. **Update Test Logic**
   - Use repository pattern for data access
   - Use context for all operations
   - Update assertions for new data structures

3. **Add Test Utilities**
   - Create test configuration helpers
   - Add test database setup utilities
   - Implement test service factories

## Deployment Migration

### Before (Manual Docker Setup)

```dockerfile
FROM golang:1.19-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
CMD ["./main"]
```

### After (EggyByte Core)

```dockerfile
FROM golang:1.25.1-alpine AS builder
WORKDIR /workspace
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /app/service cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /app/service .
EXPOSE 8080 9090
CMD ["./service"]
```

### Migration Steps

1. **Update Dockerfile**
   - Use multi-stage build
   - Add proper layer caching
   - Expose standard ports (8080, 9090)

2. **Update Environment Variables**
   - Add required environment variables
   - Use standard configuration names
   - Set up proper defaults

3. **Update Kubernetes Manifests**
   - Add health check endpoints
   - Configure proper probes
   - Set up monitoring

## Common Pitfalls

### 1. Not Implementing Service Interface

**Wrong:**
```go
func main() {
    // Direct HTTP server setup
    http.ListenAndServe(":8080", nil)
}
```

**Correct:**
```go
type HTTPServer struct {
    port   int
    server *http.Server
}

func (s *HTTPServer) Start(ctx context.Context) error {
    // HTTP server setup
}

func (s *HTTPServer) Stop(ctx context.Context) error {
    // Graceful shutdown
}
```

### 2. Not Using Context

**Wrong:**
```go
func (r *UserRepository) GetUser(id uint) (*User, error) {
    return r.db.First(&User{}, id).Error
}
```

**Correct:**
```go
func (r *UserRepository) GetUser(ctx context.Context, id uint) (*User, error) {
    var user User
    err := r.db.WithContext(ctx).First(&user, id).Error
    return &user, err
}
```

### 3. Not Using Structured Logging

**Wrong:**
```go
log.Printf("User created: %s", user.Email)
```

**Correct:**
```go
log.Info("User created",
    log.Field{Key: "user_id", Value: user.ID},
    log.Field{Key: "email", Value: user.Email},
)
```

### 4. Not Handling Errors Properly

**Wrong:**
```go
if err != nil {
    panic(err)
}
```

**Correct:**
```go
if err != nil {
    log.Error("Operation failed",
        log.Field{Key: "error", Value: err.Error()},
    )
    return err
}
```

## Migration Checklist

### Pre-Migration
- [ ] Analyze current codebase structure
- [ ] Identify framework-specific code
- [ ] Plan migration strategy
- [ ] Set up development environment

### During Migration
- [ ] Replace framework imports with EggyByte Core
- [ ] Update main function to use `core.Bootstrap()`
- [ ] Implement `Service` interface for HTTP servers
- [ ] Convert configuration to environment variables
- [ ] Update database operations to use repository pattern
- [ ] Replace logging with structured logging
- [ ] Update error handling
- [ ] Add context to all operations

### Post-Migration
- [ ] Test all functionality
- [ ] Update deployment configurations
- [ ] Update documentation
- [ ] Train team on new patterns
- [ ] Monitor performance and errors

## Support

If you encounter issues during migration:

1. **Check Documentation** - Review the [API Reference](API_REFERENCE.md)
2. **Review Examples** - Look at the [examples directory](examples/)
3. **Ask Questions** - Use [GitHub Discussions](https://github.com/eggybyte-technology/go-eggybyte-core/discussions)
4. **Report Issues** - Create [GitHub Issues](https://github.com/eggybyte-technology/go-eggybyte-core/issues)
5. **Contact Support** - Email support@eggybyte.com