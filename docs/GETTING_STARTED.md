# Getting Started Guide

## Quick Start

### Prerequisites

- Go 1.24.5 or later
- Docker (optional, for local development)
- Make (optional, for build automation)

### Installation

#### Install as Library

```bash
go get github.com/eggybyte-technology/go-eggybyte-core
```

#### Install CLI Tool

```bash
go install github.com/eggybyte-technology/go-eggybyte-core/cmd/ebcctl@latest
```

Verify installation:

```bash
ebcctl version
```

## Your First Microservice

### Step 1: Create a New Service

```bash
ebcctl init backend my-service
cd my-service
```

This creates a complete microservice structure:

```
my-service/
├── cmd/main.go              # Service entry point
├── internal/
│   ├── handlers/            # HTTP/gRPC handlers
│   ├── services/            # Business logic
│   └── repositories/        # Data access layer
├── go.mod                   # Go module definition
├── README.md                # Service documentation
├── ENV.md                   # Configuration guide
├── Dockerfile               # Container configuration
└── .gitignore               # Git ignore rules
```

### Step 2: Configure Your Service

Set environment variables:

```bash
export SERVICE_NAME=my-service
export PORT=8080
export METRICS_PORT=9090
export LOG_LEVEL=info
export LOG_FORMAT=console
```

### Step 3: Run Your Service

```bash
go run cmd/main.go
```

Your service is now running with:
- HTTP server on port 8080
- Metrics and health checks on port 9090
- Structured logging
- Graceful shutdown handling

### Step 4: Test Your Service

Check health status:

```bash
curl http://localhost:9090/healthz
```

View metrics:

```bash
curl http://localhost:9090/metrics
```

## Adding Database Support

### Step 1: Start MySQL Database

```bash
docker run -d \
  --name mysql \
  -e MYSQL_ROOT_PASSWORD=root \
  -e MYSQL_DATABASE=test \
  -e MYSQL_USER=test \
  -e MYSQL_PASSWORD=test \
  -p 3306:3306 \
  mysql:8.0
```

### Step 2: Configure Database Connection

```bash
export DATABASE_DSN="test:test@tcp(localhost:3306)/test?charset=utf8mb4&parseTime=True&loc=Local"
```

### Step 3: Generate Repository Code

```bash
ebcctl new repo user
```

This generates `internal/repositories/user_repository.go` with:
- User model definition
- Repository with CRUD operations
- Auto-registration for table migration

### Step 4: Import Repository Package

Add to `cmd/main.go`:

```go
import _ "my-service/internal/repositories"
```

### Step 5: Use Repository in Your Code

```go
import (
    "github.com/eggybyte-technology/go-eggybyte-core/pkg/db"
    "my-service/internal/repositories"
)

func main() {
    // ... bootstrap code ...
    
    // Get database connection
    database := db.GetDB()
    
    // Create repository instance
    userRepo := repositories.NewUserRepository(database)
    
    // Use repository
    user := &repositories.User{
        Name:  "John Doe",
        Email: "john@example.com",
    }
    
    err := userRepo.Create(context.Background(), user)
    if err != nil {
        log.Error("Failed to create user", log.Field{Key: "error", Value: err})
    }
}
```

## Adding Cache Support

### Step 1: Start Memcached

```bash
docker run -d \
  --name memcached \
  -p 11211:11211 \
  memcached:1.6-alpine
```

### Step 2: Configure Cache Connection

```bash
export CACHE_SERVERS="localhost:11211"
```

### Step 3: Use Cache in Your Code

```go
import "github.com/eggybyte-technology/go-eggybyte-core/pkg/cache"

func main() {
    // ... bootstrap code ...
    
    // Get cache client
    cacheClient := cache.GetClient()
    cacheService := cache.NewCacheService(cacheClient)
    
    // Use cache
    ctx := context.Background()
    err := cacheService.Set(ctx, "key", []byte("value"), time.Minute)
    if err != nil {
        log.Error("Failed to set cache", log.Field{Key: "error", Value: err})
    }
    
    value, err := cacheService.Get(ctx, "key")
    if err != nil {
        log.Error("Failed to get cache", log.Field{Key: "error", Value: err})
    }
}
```

## Adding Custom Services

### Step 1: Implement Service Interface

```go
type MyCustomService struct {
    port int
    server *http.Server
}

func (s *MyCustomService) Start(ctx context.Context) error {
    s.server = &http.Server{
        Addr:    fmt.Sprintf(":%d", s.port),
        Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            w.WriteHeader(http.StatusOK)
            w.Write([]byte("Hello from custom service!"))
        }),
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

func (s *MyCustomService) Stop(ctx context.Context) error {
    if s.server != nil {
        return s.server.Shutdown(ctx)
    }
    return nil
}
```

### Step 2: Register Service with Bootstrap

```go
func main() {
    cfg := &config.Config{}
    config.MustReadFromEnv(cfg)
    
    // Create custom service
    customService := &MyCustomService{port: 8081}
    
    // Bootstrap with custom service
    if err := core.Bootstrap(cfg, customService); err != nil {
        log.Fatal("Bootstrap failed", log.Field{Key: "error", Value: err})
    }
}
```

## Adding Custom Health Checks

### Step 1: Implement Health Checker Interface

```go
type MyHealthChecker struct {
    service string
}

func (h *MyHealthChecker) Name() string {
    return h.service
}

func (h *MyHealthChecker) Check(ctx context.Context) error {
    // Implement your health check logic
    // Return nil if healthy, error if unhealthy
    return nil
}
```

### Step 2: Register Health Checker

```go
func main() {
    cfg := &config.Config{}
    config.MustReadFromEnv(cfg)
    
    // Create health checker
    healthChecker := &MyHealthChecker{service: "my-service"}
    
    // Register with monitoring service
    monitoringService := monitoring.NewMonitoringService(cfg.MetricsPort)
    monitoringService.AddChecker(healthChecker)
    
    // Bootstrap with monitoring service
    if err := core.Bootstrap(cfg, monitoringService); err != nil {
        log.Fatal("Bootstrap failed", log.Field{Key: "error", Value: err})
    }
}
```

## Development Workflow

### Local Development

1. **Start Dependencies**
   ```bash
   docker-compose up -d
   ```

2. **Set Environment Variables**
   ```bash
   export SERVICE_NAME=my-service
   export DATABASE_DSN="test:test@tcp(localhost:3306)/test?charset=utf8mb4&parseTime=True&loc=Local"
   export CACHE_SERVERS="localhost:11211"
   export LOG_LEVEL=debug
   export LOG_FORMAT=console
   ```

3. **Run Service**
   ```bash
   go run cmd/main.go
   ```

4. **Test Endpoints**
   ```bash
   curl http://localhost:8080/health
   curl http://localhost:9090/healthz
   curl http://localhost:9090/metrics
   ```

### Building and Testing

1. **Build Service**
   ```bash
   go build -o bin/my-service cmd/main.go
   ```

2. **Run Tests**
   ```bash
   go test ./...
   ```

3. **Run Tests with Coverage**
   ```bash
   go test -coverprofile=coverage.out ./...
   go tool cover -html=coverage.out
   ```

### Docker Development

1. **Build Docker Image**
   ```bash
   docker build -t my-service .
   ```

2. **Run Container**
   ```bash
   docker run -d \
     --name my-service \
     -p 8080:8080 \
     -p 9090:9090 \
     -e SERVICE_NAME=my-service \
     -e DATABASE_DSN="test:test@tcp(host.docker.internal:3306)/test?charset=utf8mb4&parseTime=True&loc=Local" \
     -e CACHE_SERVERS="host.docker.internal:11211" \
     my-service
   ```

## Configuration Reference

### Environment Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `SERVICE_NAME` | Service identifier | - | Yes |
| `ENVIRONMENT` | Deployment environment | `development` | No |
| `PORT` | HTTP server port | `8080` | No |
| `METRICS_PORT` | Metrics server port | `9090` | No |
| `LOG_LEVEL` | Log level (debug, info, warn, error, fatal) | `info` | No |
| `LOG_FORMAT` | Log format (json, console) | `json` | No |
| `DATABASE_DSN` | Database connection string | - | No |
| `DATABASE_MAX_OPEN_CONNS` | Max open database connections | `100` | No |
| `DATABASE_MAX_IDLE_CONNS` | Max idle database connections | `10` | No |
| `CACHE_SERVERS` | Memcached server addresses | - | No |
| `CACHE_MAX_IDLE_CONNS` | Max idle cache connections | `10` | No |
| `CACHE_TIMEOUT` | Cache operation timeout | `5s` | No |
| `CACHE_CONNECT_TIMEOUT` | Cache connection timeout | `5s` | No |

### Database DSN Format

```
username:password@tcp(host:port)/database?charset=utf8mb4&parseTime=True&loc=Local
```

### Cache Servers Format

```
server1:port1,server2:port2,server3:port3
```

## Troubleshooting

### Common Issues

1. **Service Won't Start**
   - Check if ports are available
   - Verify environment variables
   - Check logs for errors

2. **Database Connection Failed**
   - Verify DSN format
   - Check database server status
   - Verify credentials

3. **Cache Connection Failed**
   - Check server addresses
   - Verify network connectivity
   - Check server configuration

4. **Health Checks Failing**
   - Check dependency status
   - Verify health check logic
   - Review error logs

### Debug Mode

Enable debug logging:

```bash
export LOG_LEVEL=debug
export LOG_FORMAT=console
```

### Health Check Endpoints

- `GET /healthz` - Combined health check
- `GET /livez` - Liveness probe
- `GET /readyz` - Readiness probe
- `GET /metrics` - Prometheus metrics

### Log Analysis

Look for these log patterns:

- `Starting service bootstrap` - Service startup
- `Database connection established` - Database connected
- `Cache connection established` - Cache connected
- `Service shutdown completed` - Graceful shutdown

## Next Steps

1. **Explore Examples** - Check out the [examples directory](examples/)
2. **Read Architecture Guide** - Learn about [architecture patterns](ARCHITECTURE.md)
3. **API Reference** - Review the [API documentation](API_REFERENCE.md)
4. **Contributing** - See how to [contribute](../CONTRIBUTING.md)

## Support

- **Documentation** - Check the [docs directory](.)
- **Issues** - Report bugs on [GitHub Issues](https://github.com/eggybyte-technology/go-eggybyte-core/issues)
- **Discussions** - Ask questions on [GitHub Discussions](https://github.com/eggybyte-technology/go-eggybyte-core/discussions)
- **Email** - Contact us at support@eggybyte.com