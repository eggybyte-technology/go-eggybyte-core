# go-eggybyte-core Examples

This directory contains complete working examples demonstrating how to use go-eggybyte-core.

## 📁 Examples

### demo-platform/

A complete full-stack application generated with `ebcctl init project demo-platform`.

**Structure**:
```
demo-platform/
├── backend/
│   └── services/
│       ├── auth/       # Authentication service
│       └── user/       # User management service
├── frontend/           # Flutter application
├── api/                # API definitions
├── Makefile            # Build automation
└── docker-compose.yml  # Local development
```

**Features Demonstrated**:
- ✅ Complete project structure
- ✅ Multiple backend microservices
- ✅ Flutter frontend integration
- ✅ Local development with Docker Compose
- ✅ Repository auto-registration
- ✅ Unified monitoring endpoints
- ✅ Local go-eggybyte-core dependency

## 🚀 Running the Examples

### Backend Services

#### 1. Run Auth Service

```bash
cd demo-platform/backend/services/auth

# Set environment variables
export SERVICE_NAME=auth-service
export PORT=8080
export METRICS_PORT=9090
export LOG_LEVEL=info
export LOG_FORMAT=console

# Run the service
go run cmd/main.go
```

#### 2. Access Monitoring Endpoints

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

#### 3. Run with Database

```bash
# Start TiDB with Docker
docker run -d --name tidb -p 4000:4000 pingcap/tidb:latest

# Set DATABASE_DSN
export DATABASE_DSN="root@tcp(localhost:4000)/test?charset=utf8mb4&parseTime=True"

# Run service (will auto-migrate tables)
go run cmd/main.go
```

### Frontend Application

```bash
cd demo-platform/frontend

# Install dependencies
flutter pub get

# Run on web
flutter run -d chrome

# Run on mobile
flutter run
```

## 📝 What Each Example Demonstrates

### demo-platform/backend/services/auth

**Demonstrates**:
- Minimal service with Bootstrap
- Repository auto-registration (Session repository)
- No database configuration (runs without DB)
- Monitoring endpoints
- Graceful shutdown

**Key Files**:
- `cmd/main.go` - 2-line Bootstrap pattern
- `internal/repositories/session_repository.go` - Auto-registered repository
- `go.mod` - Local replace directive for go-eggybyte-core

### demo-platform/backend/services/user

**Demonstrates**:
- User management service
- Repository auto-registration (User repository)
- CRUD operations pattern
- Same Bootstrap pattern as auth service

**Key Files**:
- `cmd/main.go` - Identical Bootstrap to auth service
- `internal/repositories/user_repository.go` - User model and CRUD

## 🔧 Development Tips

### 1. Local Development with go-eggybyte-core

All generated projects use local replace directive:

```go
// go.mod
replace github.com/eggybyte-technology/go-eggybyte-core => ../../../../../
```

This allows you to:
- Modify core library and see changes immediately
- Debug into core library code
- Test core changes without publishing

### 2. Adding New Repositories

```bash
cd demo-platform/backend/services/user
ebcctl new repo order
```

This generates:
- `internal/repositories/order_repository.go`
- Auto-registers via `init()`
- Implements CRUD operations
- Includes full documentation

### 3. Database Auto-Migration

When you start a service with `DATABASE_DSN`:

1. Bootstrap calls db initializer
2. All registered repositories' `InitTable()` are called
3. GORM AutoMigrate creates/updates tables
4. Service starts successfully

**No manual migration needed!**

### 4. Monitoring in Kubernetes

```yaml
apiVersion: v1
kind: Service
metadata:
  name: auth-service
spec:
  ports:
  - name: http
    port: 8080
  - name: monitoring
    port: 9090  # Prometheus and health probes

---
apiVersion: v1
kind: Pod
metadata:
  name: auth
spec:
  containers:
  - name: auth
    image: auth-service:latest
    ports:
    - containerPort: 8080
      name: http
    - containerPort: 9090
      name: monitoring
    livenessProbe:
      httpGet:
        path: /livez
        port: 9090
    readinessProbe:
      httpGet:
        path: /readyz
        port: 9090
```

## 🎯 Common Patterns

### Minimal Service (No Database)

```go
package main

import (
    "github.com/eggybyte-technology/go-eggybyte-core/config"
    "github.com/eggybyte-technology/go-eggybyte-core/core"
    "github.com/eggybyte-technology/go-eggybyte-core/log"
)

func main() {
    cfg := &config.Config{}
    config.MustReadFromEnv(cfg)
    
    if err := core.Bootstrap(cfg); err != nil {
        log.Fatal("Bootstrap failed", log.Field{Key: "error", Value: err})
    }
}
```

Provides:
- ✅ Logging
- ✅ Monitoring (/metrics, /healthz, /livez, /readyz)
- ✅ Graceful shutdown

### Service with Database

```go
package main

import (
    "github.com/eggybyte-technology/go-eggybyte-core/config"
    "github.com/eggybyte-technology/go-eggybyte-core/core"
    "github.com/eggybyte-technology/go-eggybyte-core/log"
    
    // Import repositories for auto-registration
    _ "myservice/internal/repositories"
)

func main() {
    cfg := &config.Config{}
    config.MustReadFromEnv(cfg)
    
    if err := core.Bootstrap(cfg); err != nil {
        log.Fatal("Bootstrap failed", log.Field{Key: "error", Value: err})
    }
}
```

Provides everything above **plus**:
- ✅ Database connection
- ✅ Automatic table migration
- ✅ Connection pooling

### Service with Business Logic

```go
package main

import (
    "github.com/eggybyte-technology/go-eggybyte-core/config"
    "github.com/eggybyte-technology/go-eggybyte-core/core"
    "github.com/eggybyte-technology/go-eggybyte-core/log"
    
    _ "myservice/internal/repositories"
)

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

Provides everything above **plus**:
- ✅ Your HTTP server
- ✅ Your gRPC server
- ✅ Concurrent service management

## 🐛 Troubleshooting

### Port Already in Use

```bash
# Kill existing service
pkill -f "./bin/auth"

# Or use different port
export METRICS_PORT=9091
go run cmd/main.go
```

### Database Connection Failed

```bash
# Check TiDB is running
docker ps | grep tidb

# Test connection
mysql -h 127.0.0.1 -P 4000 -u root -e "SELECT 1"

# Check DSN format
export DATABASE_DSN="root@tcp(localhost:4000)/test?charset=utf8mb4&parseTime=True"
```

### go.mod Replace Path Issues

```bash
# From service directory, verify path to core
cd backend/services/auth
ls ../../../../../go.mod  # Should exist

# If wrong, update go.mod replace directive
```

## 📖 Next Steps

1. **Explore the code**: Read through generated files to understand patterns
2. **Modify repositories**: Add fields to models and see auto-migration
3. **Add business logic**: Implement handlers, services following three-layer pattern
4. **Deploy to Kubernetes**: Use provided health probes and metrics
5. **Generate more code**: Use `ebcctl new repo` to add repositories

## 💡 Key Takeaways

- **2 lines to start a service**: Bootstrap handles everything
- **Auto-registration**: Import repositories to activate them
- **Unified monitoring**: Single port for all observability
- **Production-ready**: Metrics, health checks, graceful shutdown built-in
- **Developer-friendly**: Local core dependency for rapid iteration

## 🤝 Contributing

Found an issue or have improvements? Contributions welcome!

1. Test changes in examples/demo-platform
2. Ensure all services compile and run
3. Update this documentation
4. Submit PR with examples demonstrating new features

