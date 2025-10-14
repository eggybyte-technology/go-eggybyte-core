# API Reference

## Core Module (`pkg/core`)

### Bootstrap Function

```go
func Bootstrap(cfg *config.Config, businessServices ...service.Service) error
```

**Description:** Single entry point for all EggyByte services. Orchestrates the complete application lifecycle including configuration loading, logging initialization, infrastructure setup, and service startup.

**Parameters:**
- `cfg` - Service configuration loaded from environment variables
- `businessServices` - Application-specific services to run

**Returns:**
- `error` - Returns error if any initialization or startup step fails

**Example:**
```go
cfg := &config.Config{}
config.MustReadFromEnv(cfg)

httpServer := NewHTTPServer(cfg.Port)
grpcServer := NewGRPCServer(9090)

if err := core.Bootstrap(cfg, httpServer, grpcServer); err != nil {
    log.Fatal("Bootstrap failed", log.Field{Key: "error", Value: err})
}
```

## Configuration Module (`pkg/config`)

### Config Structure

```go
type Config struct {
    ServiceName string `envconfig:"SERVICE_NAME" required:"true"`
    Environment string `envconfig:"ENVIRONMENT" default:"development"`
    Port        int    `envconfig:"PORT" default:"8080"`
    MetricsPort int    `envconfig:"METRICS_PORT" default:"9090"`
    LogLevel    string `envconfig:"LOG_LEVEL" default:"info"`
    LogFormat   string `envconfig:"LOG_FORMAT" default:"json"`
    DatabaseDSN string `envconfig:"DATABASE_DSN"`
    // ... other fields
}
```

### Functions

#### ReadFromEnv
```go
func ReadFromEnv(cfg interface{}) error
```

**Description:** Loads configuration from environment variables into the provided struct.

**Parameters:**
- `cfg` - Pointer to configuration struct

**Returns:**
- `error` - Returns error if loading fails

#### MustReadFromEnv
```go
func MustReadFromEnv(cfg interface{})
```

**Description:** Loads configuration from environment variables, panicking on error.

**Parameters:**
- `cfg` - Pointer to configuration struct

#### Get
```go
func Get() *Config
```

**Description:** Returns the current global configuration.

**Returns:**
- `*Config` - The current configuration, or nil if not initialized

#### Set
```go
func Set(cfg *Config)
```

**Description:** Updates the global configuration with a new instance.

**Parameters:**
- `cfg` - The new configuration to set globally

## Logging Module (`pkg/log`)

### Field Structure

```go
type Field struct {
    Key   string
    Value interface{}
}
```

### Functions

#### Init
```go
func Init(level, format string) error
```

**Description:** Initializes the global logger with the specified level and format.

**Parameters:**
- `level` - Log level (debug, info, warn, error, fatal)
- `format` - Log format (json, console)

**Returns:**
- `error` - Returns error if initialization fails

#### Info
```go
func Info(msg string, fields ...Field)
```

**Description:** Logs an informational message with optional fields.

**Parameters:**
- `msg` - Log message
- `fields` - Optional structured fields

#### Error
```go
func Error(msg string, fields ...Field)
```

**Description:** Logs an error message with optional fields.

**Parameters:**
- `msg` - Log message
- `fields` - Optional structured fields

#### Fatal
```go
func Fatal(msg string, fields ...Field)
```

**Description:** Logs a fatal message and exits the program.

**Parameters:**
- `msg` - Log message
- `fields` - Optional structured fields

#### WithContext
```go
func WithContext(ctx context.Context, fields ...Field) (context.Context, *Logger)
```

**Description:** Creates a new context with attached logger and fields.

**Parameters:**
- `ctx` - Parent context
- `fields` - Fields to attach to context

**Returns:**
- `context.Context` - New context with logger
- `*Logger` - Logger instance

#### FromContext
```go
func FromContext(ctx context.Context) *Logger
```

**Description:** Retrieves logger from context.

**Parameters:**
- `ctx` - Context containing logger

**Returns:**
- `*Logger` - Logger instance, or default logger if not found

## Database Module (`pkg/db`)

### Config Structure

```go
type Config struct {
    DSN             string
    MaxOpenConns    int
    MaxIdleConns    int
    ConnMaxLifetime time.Duration
    ConnMaxIdleTime time.Duration
    LogLevel        string
}
```

### Repository Interface

```go
type Repository interface {
    TableName() string
    InitTable(ctx context.Context, db *gorm.DB) error
}
```

### Functions

#### DefaultConfig
```go
func DefaultConfig() *Config
```

**Description:** Returns database configuration with sensible defaults.

**Returns:**
- `*Config` - Configuration with default values

#### Connect
```go
func Connect(cfg *Config) (*gorm.DB, error)
```

**Description:** Establishes a database connection using the provided configuration.

**Parameters:**
- `cfg` - Database configuration parameters

**Returns:**
- `*gorm.DB` - The established database connection
- `error` - Returns error if connection fails

#### GetDB
```go
func GetDB() *gorm.DB
```

**Description:** Returns the global database connection.

**Returns:**
- `*gorm.DB` - The global database connection, or nil if not initialized

#### SetDB
```go
func SetDB(db *gorm.DB)
```

**Description:** Updates the global database connection.

**Parameters:**
- `db` - Database connection to use as global instance

#### RegisterRepository
```go
func RegisterRepository(repo Repository)
```

**Description:** Registers a repository for automatic table initialization.

**Parameters:**
- `repo` - Repository implementing the Repository interface

#### InitializeAllTables
```go
func InitializeAllTables(ctx context.Context, db *gorm.DB) error
```

**Description:** Initializes all registered repository tables.

**Parameters:**
- `ctx` - Context for timeout control
- `db` - Database connection

**Returns:**
- `error` - Returns error if initialization fails

#### Close
```go
func Close() error
```

**Description:** Closes the global database connection.

**Returns:**
- `error` - Returns error if closing fails

### TiDBInitializer

```go
type TiDBInitializer struct {
    config *Config
}
```

#### NewTiDBInitializer
```go
func NewTiDBInitializer(cfg *Config) *TiDBInitializer
```

**Description:** Creates a new database initializer with the given configuration.

**Parameters:**
- `cfg` - Database configuration including DSN and connection pool settings

**Returns:**
- `*TiDBInitializer` - Initializer instance ready to be registered with launcher

#### Init
```go
func (t *TiDBInitializer) Init(ctx context.Context) error
```

**Description:** Establishes database connection and initializes all repository tables.

**Parameters:**
- `ctx` - Context for timeout control and cancellation

**Returns:**
- `error` - Returns error if connection fails or table initialization fails

## Cache Module (`pkg/cache`)

### Config Structure

```go
type Config struct {
    Servers         []string
    MaxIdleConns    int
    Timeout         time.Duration
    ConnectTimeout  time.Duration
}
```

### Functions

#### DefaultConfig
```go
func DefaultConfig() *Config
```

**Description:** Returns cache configuration with sensible defaults.

**Returns:**
- `*Config` - Configuration with default values

#### Connect
```go
func Connect(cfg *Config) (*memcache.Client, error)
```

**Description:** Establishes a Memcached connection using the provided configuration.

**Parameters:**
- `cfg` - Cache configuration parameters

**Returns:**
- `*memcache.Client` - The established cache connection
- `error` - Returns error if connection fails

#### GetClient
```go
func GetClient() *memcache.Client
```

**Description:** Returns the global Memcached client.

**Returns:**
- `*memcache.Client` - The global cache client, or nil if not initialized

#### SetClient
```go
func SetClient(client *memcache.Client)
```

**Description:** Updates the global Memcached client.

**Parameters:**
- `client` - Cache client to use as global instance

#### Close
```go
func Close() error
```

**Description:** Closes the Memcached connection.

**Returns:**
- `error` - Returns error if closing fails

### CacheService

```go
type CacheService struct {
    client *memcache.Client
}
```

#### NewCacheService
```go
func NewCacheService(client *memcache.Client) *CacheService
```

**Description:** Creates a new cache service.

**Parameters:**
- `client` - Memcached client

**Returns:**
- `*CacheService` - Cache service instance

#### Get
```go
func (c *CacheService) Get(ctx context.Context, key string) ([]byte, error)
```

**Description:** Retrieves a value from cache.

**Parameters:**
- `ctx` - Context for timeout control
- `key` - Cache key

**Returns:**
- `[]byte` - Cached value, or nil if not found
- `error` - Returns error if operation fails

#### Set
```go
func (c *CacheService) Set(ctx context.Context, key string, value []byte, expiration time.Duration) error
```

**Description:** Stores a value in cache with expiration.

**Parameters:**
- `ctx` - Context for timeout control
- `key` - Cache key
- `value` - Value to store
- `expiration` - Expiration duration

**Returns:**
- `error` - Returns error if operation fails

#### Delete
```go
func (c *CacheService) Delete(ctx context.Context, key string) error
```

**Description:** Removes a value from cache.

**Parameters:**
- `ctx` - Context for timeout control
- `key` - Cache key

**Returns:**
- `error` - Returns error if operation fails

#### Exists
```go
func (c *CacheService) Exists(ctx context.Context, key string) (bool, error)
```

**Description:** Checks if a key exists in cache.

**Parameters:**
- `ctx` - Context for timeout control
- `key` - Cache key

**Returns:**
- `bool` - True if key exists, false otherwise
- `error` - Returns error if operation fails

### CacheInitializer

```go
type CacheInitializer struct {
    config *Config
}
```

#### NewCacheInitializer
```go
func NewCacheInitializer(cfg *Config) *CacheInitializer
```

**Description:** Creates a new cache initializer.

**Parameters:**
- `cfg` - Cache configuration

**Returns:**
- `*CacheInitializer` - Cache initializer instance

#### Init
```go
func (c *CacheInitializer) Init(ctx context.Context) error
```

**Description:** Establishes cache connection and verifies connectivity.

**Parameters:**
- `ctx` - Context for timeout control

**Returns:**
- `error` - Returns error if connection fails

## Service Module (`pkg/service`)

### Service Interface

```go
type Service interface {
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
}
```

### Initializer Interface

```go
type Initializer interface {
    Init(ctx context.Context) error
}
```

### Launcher

```go
type Launcher struct {
    // ... internal fields
}
```

#### NewLauncher
```go
func NewLauncher() *Launcher
```

**Description:** Creates a new service launcher.

**Returns:**
- `*Launcher` - Launcher instance

#### AddService
```go
func (l *Launcher) AddService(svc Service)
```

**Description:** Adds a service to the launcher.

**Parameters:**
- `svc` - Service to add

#### AddInitializer
```go
func (l *Launcher) AddInitializer(init Initializer)
```

**Description:** Adds an initializer to the launcher.

**Parameters:**
- `init` - Initializer to add

#### SetLogger
```go
func (l *Launcher) SetLogger(logger *log.Logger)
```

**Description:** Sets the logger for the launcher.

**Parameters:**
- `logger` - Logger instance

#### Run
```go
func (l *Launcher) Run(ctx context.Context) error
```

**Description:** Runs all services with lifecycle management.

**Parameters:**
- `ctx` - Context for cancellation

**Returns:**
- `error` - Returns error if any service fails

## Monitoring Module (`pkg/monitoring`)

### MonitoringService

```go
type MonitoringService struct {
    port int
    // ... internal fields
}
```

#### NewMonitoringService
```go
func NewMonitoringService(port int) *MonitoringService
```

**Description:** Creates a new monitoring service.

**Parameters:**
- `port` - Port to listen on

**Returns:**
- `*MonitoringService` - Monitoring service instance

#### AddChecker
```go
func (m *MonitoringService) AddChecker(checker HealthChecker)
```

**Description:** Adds a health checker to the monitoring service.

**Parameters:**
- `checker` - Health checker to add

### HealthChecker Interface

```go
type HealthChecker interface {
    Name() string
    Check(ctx context.Context) error
}
```

## Project Structure

### Manual Service Creation

Since EggyByte Core is a pure library, you'll need to create your service structure manually. Here's the recommended approach:

#### Create Service Directory
```bash
mkdir my-service
cd my-service
go mod init my-service
```

#### Add Core Dependency
```bash
go get github.com/eggybyte-technology/go-eggybyte-core
```

#### Create Basic Structure
```bash
mkdir -p cmd internal/handlers internal/services internal/repositories
```

#### Create Main Entry Point
Create `cmd/main.go`:
```go
package main

import (
    "github.com/eggybyte-technology/go-eggybyte-core/pkg/config"
    "github.com/eggybyte-technology/go-eggybyte-core/pkg/core"
)

func main() {
    cfg := &config.Config{}
    config.MustReadFromEnv(cfg)
    
    if err := core.Bootstrap(cfg); err != nil {
        panic(err)
    }
}
```

### Repository Pattern

For database integration, follow this pattern:

```go
// internal/repositories/user_repository.go
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

// Auto-register on import
func init() {
    db.RegisterRepository(&UserRepository{})
}
```

## Error Handling

### Common Error Types

#### Configuration Errors
- Invalid environment variables
- Missing required configuration
- Invalid configuration values

#### Database Errors
- Connection failures
- Query execution errors
- Transaction failures

#### Cache Errors
- Connection failures
- Operation timeouts
- Key not found

#### Service Errors
- Startup failures
- Shutdown failures
- Health check failures

### Error Handling Best Practices

1. **Use Structured Logging**
   ```go
   log.Error("Database operation failed",
       log.Field{Key: "operation", Value: "create_user"},
       log.Field{Key: "error", Value: err.Error()},
   )
   ```

2. **Return Meaningful Errors**
   ```go
   if err != nil {
       return fmt.Errorf("failed to create user %s: %w", userID, err)
   }
   ```

3. **Handle Context Cancellation**
   ```go
   select {
   case <-ctx.Done():
       return ctx.Err()
   case result := <-resultCh:
       return result
   }
   ```

4. **Implement Graceful Degradation**
   ```go
   if err := cacheService.Set(ctx, key, value, expiration); err != nil {
       log.Warn("Cache set failed, continuing without cache",
           log.Field{Key: "error", Value: err})
   }
   ```

## Performance Considerations

### Database Performance
- Use connection pooling
- Implement query optimization
- Use transactions appropriately
- Monitor database performance

### Cache Performance
- Cache frequently accessed data
- Use appropriate cache keys
- Set reasonable expiration times
- Monitor cache hit rates

### Logging Performance
- Use appropriate log levels
- Avoid logging sensitive data
- Use structured logging
- Monitor log volume

### Monitoring Performance
- Expose relevant metrics
- Implement health checks
- Monitor resource usage
- Set up alerts

## Security Considerations

### Configuration Security
- Use environment variables for secrets
- Validate all configuration inputs
- Use secure defaults
- Rotate secrets regularly

### Database Security
- Use connection encryption
- Implement proper authentication
- Use least privilege principles
- Monitor database access

### Logging Security
- Avoid logging sensitive data
- Use structured logging
- Implement log rotation
- Monitor log access

### Network Security
- Use TLS for all communications
- Implement proper authentication
- Use network policies
- Monitor network traffic