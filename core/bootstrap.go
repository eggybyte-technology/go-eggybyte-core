package core

import (
	"context"
	"fmt"

	"github.com/eggybyte-technology/go-eggybyte-core/config"
	"github.com/eggybyte-technology/go-eggybyte-core/db"
	"github.com/eggybyte-technology/go-eggybyte-core/health"
	"github.com/eggybyte-technology/go-eggybyte-core/log"
	"github.com/eggybyte-technology/go-eggybyte-core/metrics"
	"github.com/eggybyte-technology/go-eggybyte-core/service"
)

// Bootstrap is the single entry point for all EggyByte services.
// It orchestrates the complete application lifecycle including:
//   - Configuration loading
//   - Logging initialization
//   - Infrastructure setup (database, metrics, health)
//   - Business service registration and startup
//   - Graceful shutdown handling
//
// This function simplifies service creation to a single call with
// automatic handling of all boilerplate setup and teardown.
//
// Parameters:
//   - cfg: Service configuration loaded from environment variables
//   - businessServices: Application-specific services to run
//
// Returns:
//   - error: Returns error if any initialization or startup step fails
//
// Behavior:
//   - Initializes logging based on config
//   - Creates service launcher
//   - Conditionally registers database initializer if DSN provided
//   - Registers metrics and health services
//   - Registers provided business services
//   - Runs launcher with signal handling
//
// Example:
//
//	cfg := &config.Config{}
//	if err := config.ReadFromEnv(cfg); err != nil {
//	    log.Fatal("Failed to load config", log.Field{Key: "error", Value: err})
//	}
//
//	httpServer := myapp.NewHTTPServer(cfg.Port)
//	grpcServer := myapp.NewGRPCServer(9090)
//
//	if err := core.Bootstrap(cfg, httpServer, grpcServer); err != nil {
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

	// Phase 5: Register infrastructure services
	registerInfraServices(launcher, cfg)

	// Phase 6: Register business services
	for _, svc := range businessServices {
		launcher.AddService(svc)
	}

	// Phase 7: Run launcher with complete lifecycle management
	log.Info("Launching services",
		log.Field{Key: "service_count", Value: len(businessServices) + 2}) // +2 for metrics and health

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
// Only registers database initializer if DSN is provided in configuration.
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

// registerInfraServices registers core infrastructure services
// (metrics and health) with the launcher.
func registerInfraServices(launcher *service.Launcher, cfg *config.Config) {
	// Metrics service
	metricsService := metrics.NewMetricsService(cfg.MetricsPort)
	launcher.AddService(metricsService)
	log.Info("Metrics service registered", log.Field{Key: "port", Value: cfg.MetricsPort})

	// Health service
	healthService := health.NewHealthService(cfg.MetricsPort)
	launcher.AddService(healthService)
	log.Info("Health service registered", log.Field{Key: "port", Value: cfg.MetricsPort})
}
