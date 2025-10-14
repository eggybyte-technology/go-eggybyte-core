// Package core provides the main bootstrap functionality for EggyByte services.
// It orchestrates the complete application lifecycle including configuration loading,
// logging initialization, infrastructure setup, and service startup with graceful shutdown.
package core

import (
	"context"
	"fmt"
	"strconv"

	"github.com/eggybyte-technology/go-eggybyte-core/pkg/config"
	"github.com/eggybyte-technology/go-eggybyte-core/pkg/db"
	"github.com/eggybyte-technology/go-eggybyte-core/pkg/log"
	"github.com/eggybyte-technology/go-eggybyte-core/pkg/monitoring"
	"github.com/eggybyte-technology/go-eggybyte-core/pkg/server"
	"github.com/eggybyte-technology/go-eggybyte-core/pkg/service"
)

// BootstrapWithContext is the same as Bootstrap but accepts a context for cancellation.
// It orchestrates the complete application lifecycle including:
//   - Configuration loading and validation
//   - Logging initialization with structured output
//   - Infrastructure setup (database, cache, monitoring)
//   - Business server creation (HTTP/gRPC based on configuration)
//   - Health check and metrics service registration
//   - Graceful shutdown handling with proper signal management
//
// This function simplifies service creation to a single call with
// automatic handling of all boilerplate setup and teardown.
//
// Parameters:
//   - cfg: Service configuration loaded from environment variables
//   - businessServices: Application-specific services to run (optional)
//
// Returns:
//   - error: Returns error if any initialization or startup step fails
//
// Behavior:
//   - Initializes logging based on config (level, format)
//   - Creates service launcher with proper lifecycle management
//   - Conditionally registers database initializer if DSN provided
//   - Creates and registers business HTTP/gRPC servers based on configuration
//   - Registers health check and metrics services on separate ports
//   - Registers any additional business services provided
//   - Runs launcher with signal handling and graceful shutdown
//
// Environment Variables:
//   - SERVICE_NAME: Required service identifier
//   - BUSINESS_HTTP_PORT: HTTP server port (default: 8080)
//   - BUSINESS_GRPC_PORT: gRPC server port (default: 9090)
//   - HEALTH_CHECK_PORT: Health check port (default: 8081)
//   - METRICS_PORT: Metrics exposition port (default: 9091)
//   - ENABLE_BUSINESS_HTTP: Enable HTTP server (default: true)
//   - ENABLE_BUSINESS_GRPC: Enable gRPC server (default: true)
//   - ENABLE_HEALTH_CHECK: Enable health check server (default: true)
//   - ENABLE_METRICS: Enable metrics server (default: true)
//
// Example:
//
//	cfg := &config.Config{}
//	if err := config.ReadFromEnv(cfg); err != nil {
//	    log.Fatal("Failed to load config", log.Field{Key: "error", Value: err})
//	}
//
//	// Bootstrap will automatically create HTTP/gRPC servers based on config
//	if err := core.Bootstrap(cfg); err != nil {
//	    log.Fatal("Bootstrap failed", log.Field{Key: "error", Value: err})
//	}
//
// Example with custom business services:
//
//	cfg := &config.Config{}
//	config.MustReadFromEnv(cfg)
//
//	customService := myapp.NewCustomService()
//	if err := core.Bootstrap(cfg, customService); err != nil {
//	    log.Fatal("Bootstrap failed", log.Field{Key: "error", Value: err})
//	}
//
// BootstrapWithContext is the same as Bootstrap but accepts a context for cancellation.
// This is useful for testing and scenarios where you need to control the lifecycle.
func BootstrapWithContext(ctx context.Context, cfg *config.Config, businessServices ...service.Service) error {
	// Phase 1: Initialize logging system
	if err := initializeLogging(cfg); err != nil {
		return fmt.Errorf("failed to initialize logging: %w", err)
	}

	log.Info("Starting service bootstrap",
		log.Field{Key: "service", Value: cfg.ServiceName},
		log.Field{Key: "environment", Value: cfg.Environment})

	// Phase 2: Set global configuration
	config.Set(cfg)

	// Phase 3: Create service launcher
	launcher := service.NewLauncher()
	launcher.SetLogger(log.Default())

	// Phase 4: Register infrastructure initializers
	if err := registerInitializers(launcher, cfg); err != nil {
		return err
	}

	// Phase 5: Create and register business servers
	if err := registerBusinessServers(launcher, cfg); err != nil {
		return err
	}

	// Phase 6: Register infrastructure services
	registerInfraServices(launcher, cfg)

	// Phase 7: Register additional business services
	for _, svc := range businessServices {
		launcher.AddService(svc)
	}

	// Phase 8: Run launcher with complete lifecycle management
	log.Info("Launching services",
		log.Field{Key: "service_count", Value: len(businessServices) + 3}) // +3 for business servers and monitoring

	if err := launcher.Run(ctx); err != nil {
		return fmt.Errorf("service launcher failed: %w", err)
	}

	log.Info("Service shutdown completed")
	return nil
}

// Bootstrap is the single entry point for all EggyByte services.
// It orchestrates the complete application lifecycle including:
//   - Configuration loading and validation
//   - Logging initialization with structured output
//   - Infrastructure setup (database, cache, monitoring)
//   - Business server creation (HTTP/gRPC based on configuration)
//   - Health check and metrics service registration
//   - Graceful shutdown handling with proper signal management
//
// This function simplifies service creation to a single call with
// automatic handling of all boilerplate setup and teardown.
//
// Parameters:
//   - cfg: Service configuration loaded from environment variables
//   - businessServices: Application-specific services to run (optional)
//
// Returns:
//   - error: Returns error if any initialization or startup step fails
//
// Behavior:
//   - Initializes logging based on config (level, format)
//   - Creates service launcher with proper lifecycle management
//   - Conditionally registers database initializer if DSN provided
//   - Creates and registers business HTTP/gRPC servers based on configuration
//   - Registers health check and metrics services on separate ports
//   - Registers any additional business services provided
//   - Runs launcher with signal handling and graceful shutdown
//
// Environment Variables:
//   - SERVICE_NAME: Required service identifier
//   - BUSINESS_HTTP_PORT: HTTP server port (default: 8080)
//   - BUSINESS_GRPC_PORT: gRPC server port (default: 9090)
//   - HEALTH_CHECK_PORT: Health check port (default: 8081)
//   - METRICS_PORT: Metrics exposition port (default: 9091)
//   - ENABLE_BUSINESS_HTTP: Enable HTTP server (default: true)
//   - ENABLE_BUSINESS_GRPC: Enable gRPC server (default: true)
//   - ENABLE_HEALTH_CHECK: Enable health check server (default: true)
//   - ENABLE_METRICS: Enable metrics server (default: true)
//
// Example:
//
//	cfg := &config.Config{}
//	if err := config.ReadFromEnv(cfg); err != nil {
//	    log.Fatal("Failed to load config", log.Field{Key: "error", Value: err})
//	}
//
//	// Bootstrap will automatically create HTTP/gRPC servers based on config
//	if err := core.Bootstrap(cfg); err != nil {
//	    log.Fatal("Bootstrap failed", log.Field{Key: "error", Value: err})
//	}
//
// Example with custom business services:
//
//	cfg := &config.Config{}
//	config.MustReadFromEnv(cfg)
//
//	customService := myapp.NewCustomService()
//	if err := core.Bootstrap(cfg, customService); err != nil {
//	    log.Fatal("Bootstrap failed", log.Field{Key: "error", Value: err})
//	}
func Bootstrap(cfg *config.Config, businessServices ...service.Service) error {
	// Phase 1: Initialize logging system
	if err := initializeLogging(cfg); err != nil {
		return fmt.Errorf("failed to initialize logging: %w", err)
	}

	log.Info("Starting service bootstrap",
		log.Field{Key: "service", Value: cfg.ServiceName},
		log.Field{Key: "environment", Value: cfg.Environment})

	// Phase 2: Set global configuration
	config.Set(cfg)

	// Phase 3: Create service launcher
	launcher := service.NewLauncher()
	launcher.SetLogger(log.Default())

	// Phase 4: Register infrastructure initializers
	if err := registerInitializers(launcher, cfg); err != nil {
		return err
	}

	// Phase 5: Create and register business servers
	if err := registerBusinessServers(launcher, cfg); err != nil {
		return err
	}

	// Phase 6: Register infrastructure services
	registerInfraServices(launcher, cfg)

	// Phase 7: Register additional business services
	for _, svc := range businessServices {
		launcher.AddService(svc)
	}

	// Phase 8: Run launcher with complete lifecycle management
	log.Info("Launching services",
		log.Field{Key: "service_count", Value: len(businessServices) + 3}) // +3 for business servers and monitoring

	if err := launcher.Run(context.Background()); err != nil {
		return fmt.Errorf("service launcher failed: %w", err)
	}

	log.Info("Service shutdown completed")
	return nil
}

// initializeLogging configures the global logger based on configuration.
// Returns error if log level or format is invalid.
func initializeLogging(cfg *config.Config) error {
	if err := log.Init(cfg.LogLevel, cfg.LogFormat); err != nil {
		return fmt.Errorf("invalid log configuration: %w", err)
	}
	return nil
}

// registerInitializers registers infrastructure initializers with the launcher.
// Registers database initializer if configuration is provided.
func registerInitializers(launcher *service.Launcher, cfg *config.Config) error {
	// Database initializer (conditional)
	if cfg.DatabaseDSN != "" {
		dbConfig := &db.Config{
			DSN:             cfg.DatabaseDSN,
			MaxOpenConns:    cfg.DatabaseMaxOpenConns,
			MaxIdleConns:    cfg.DatabaseMaxIdleConns,
			ConnMaxLifetime: 0, // Use default
			ConnMaxIdleTime: 0, // Use default
			LogLevel:        cfg.LogLevel,
		}

		dbInit := db.NewTiDBInitializer(dbConfig)
		launcher.AddInitializer(dbInit)

		log.Info("Database initializer registered")
	} else {
		log.Info("No database DSN provided, skipping database initialization")
	}

	return nil
}

// registerBusinessServers creates and registers business HTTP/gRPC servers based on configuration.
// This function creates servers only if they are enabled in the configuration.
//
// Parameters:
//   - launcher: The service launcher to register servers with
//   - cfg: Service configuration containing server settings
//
// Returns:
//   - error: Returns error if server creation fails
//
// Behavior:
//   - Creates HTTP server if ENABLE_BUSINESS_HTTP is true
//   - Creates gRPC server if ENABLE_BUSINESS_GRPC is true
//   - Registers servers with the launcher for lifecycle management
//   - Logs server creation and configuration details
func registerBusinessServers(launcher *service.Launcher, cfg *config.Config) error {
	var serverCount int

	// Create HTTP server if enabled
	if cfg.EnableBusinessHTTP {
		httpPort := ":" + strconv.Itoa(cfg.BusinessHTTPPort)
		httpServer := server.NewHTTPServer(httpPort)
		httpServer.SetLogger(log.Default())
		if launcher != nil {
			launcher.AddService(httpServer)
		}
		serverCount++

		log.Info("Business HTTP server registered",
			log.Field{Key: "port", Value: cfg.BusinessHTTPPort},
			log.Field{Key: "address", Value: httpPort})
	}

	// Create gRPC server if enabled
	if cfg.EnableBusinessGRPC {
		grpcPort := ":" + strconv.Itoa(cfg.BusinessGRPCPort)
		grpcServer := server.NewGRPCServer(grpcPort)
		grpcServer.SetLogger(log.Default())
		if launcher != nil {
			launcher.AddService(grpcServer)
		}
		serverCount++

		log.Info("Business gRPC server registered",
			log.Field{Key: "port", Value: cfg.BusinessGRPCPort},
			log.Field{Key: "address", Value: grpcPort})
	}

	if serverCount == 0 {
		log.Info("No business servers enabled - only infrastructure services will run")
	} else {
		log.Info("Business servers registered",
			log.Field{Key: "count", Value: serverCount},
			log.Field{Key: "http_enabled", Value: cfg.EnableBusinessHTTP},
			log.Field{Key: "grpc_enabled", Value: cfg.EnableBusinessGRPC})
	}

	return nil
}

// registerInfraServices registers core infrastructure services
// (health check and metrics services) with the launcher.
// These services run on separate ports for security and monitoring isolation.
//
// Parameters:
//   - launcher: The service launcher to register services with
//   - cfg: Service configuration containing service settings
//
// Behavior:
//   - Registers health check service if ENABLE_HEALTH_CHECK is true
//   - Registers metrics service if ENABLE_METRICS is true
//   - Logs service registration and endpoint information
func registerInfraServices(launcher *service.Launcher, cfg *config.Config) {
	var serviceCount int

	// Register health check service if enabled
	if cfg.EnableHealthCheck {
		healthService := monitoring.NewHealthService(cfg.HealthCheckPort)
		launcher.AddService(healthService)
		serviceCount++

		log.Info("Health check service registered",
			log.Field{Key: "port", Value: cfg.HealthCheckPort},
			log.Field{Key: "endpoints", Value: "/healthz, /livez, /readyz"})
	}

	// Register metrics service if enabled
	if cfg.EnableMetrics {
		metricsService := monitoring.NewMetricsService(cfg.MetricsPort)
		launcher.AddService(metricsService)
		serviceCount++

		log.Info("Metrics service registered",
			log.Field{Key: "port", Value: cfg.MetricsPort},
			log.Field{Key: "endpoints", Value: "/metrics"})
	}

	if serviceCount == 0 {
		log.Info("No infrastructure services enabled")
	} else {
		log.Info("Infrastructure services registered",
			log.Field{Key: "count", Value: serviceCount},
			log.Field{Key: "health_enabled", Value: cfg.EnableHealthCheck},
			log.Field{Key: "metrics_enabled", Value: cfg.EnableMetrics})
	}
}
