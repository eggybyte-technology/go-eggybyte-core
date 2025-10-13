package service

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/eggybyte-technology/go-eggybyte-core/log"
)

// Launcher orchestrates the application lifecycle by managing initializers and services.
// It provides a standardized pattern for:
//   - Sequential initialization of dependencies
//   - Concurrent service startup with error handling
//   - Graceful shutdown on termination signals
//   - Coordinated resource cleanup
//
// Usage pattern:
//  1. Create launcher: launcher := service.NewLauncher()
//  2. Register initializers: launcher.AddInitializer(dbInit, cacheInit)
//  3. Register services: launcher.AddService(httpServer, grpcServer)
//  4. Run: launcher.Run(context.Background())
//
// The launcher handles SIGINT and SIGTERM signals for graceful shutdown.
type Launcher struct {
	initializers    []Initializer
	services        []Service
	logger          log.Logger
	shutdownTimeout time.Duration
}

// NewLauncher creates a new service launcher with default configuration.
// The launcher is ready to accept initializers and services.
//
// Returns:
//   - *Launcher: New launcher instance with 30-second shutdown timeout
//
// Example:
//
//	launcher := service.NewLauncher()
//	launcher.AddService(httpServer)
//	launcher.Run(ctx)
func NewLauncher() *Launcher {
	return &Launcher{
		initializers:    make([]Initializer, 0),
		services:        make([]Service, 0),
		logger:          log.Default(),
		shutdownTimeout: 30 * time.Second,
	}
}

// AddInitializer registers one or more initializers to run before services start.
// Initializers execute sequentially in registration order.
//
// Parameters:
//   - inits: One or more Initializer instances
//
// Example:
//
//	launcher.AddInitializer(dbInit, cacheInit, clientInit)
func (l *Launcher) AddInitializer(inits ...Initializer) {
	l.initializers = append(l.initializers, inits...)
}

// AddService registers one or more services to run concurrently.
// Services start after all initializers complete successfully.
//
// Parameters:
//   - svcs: One or more Service instances
//
// Example:
//
//	launcher.AddService(httpServer, grpcServer, metricsServer)
func (l *Launcher) AddService(svcs ...Service) {
	l.services = append(l.services, svcs...)
}

// SetLogger configures a custom logger for launcher operations.
// By default, uses the global logger from log.Default().
//
// Parameters:
//   - logger: Logger instance to use for launcher messages
func (l *Launcher) SetLogger(logger log.Logger) {
	l.logger = logger
}

// SetShutdownTimeout configures the maximum time to wait for graceful shutdown.
// If services don't stop within this timeout, shutdown proceeds forcefully.
//
// Parameters:
//   - timeout: Maximum duration for shutdown completion
//
// Default: 30 seconds
func (l *Launcher) SetShutdownTimeout(timeout time.Duration) {
	l.shutdownTimeout = timeout
}

// Init runs all registered initializers sequentially.
// If any initializer fails, subsequent initializers are skipped and
// the error is returned immediately.
//
// Parameters:
//   - ctx: Context for timeout control and cancellation
//
// Returns:
//   - error: First initialization error encountered, or nil if all succeed
//
// Example:
//
//	if err := launcher.Init(ctx); err != nil {
//	    log.Fatal("Initialization failed", log.Field{Key: "error", Value: err})
//	}
func (l *Launcher) Init(ctx context.Context) error {
	l.logger.Info("Starting initialization phase",
		log.Field{Key: "initializer_count", Value: len(l.initializers)})

	for i, init := range l.initializers {
		l.logger.Debug("Running initializer",
			log.Field{Key: "index", Value: i},
			log.Field{Key: "type", Value: fmt.Sprintf("%T", init)})

		if err := init.Init(ctx); err != nil {
			return fmt.Errorf("initializer %d (%T) failed: %w", i, init, err)
		}
	}

	l.logger.Info("Initialization phase completed successfully")
	return nil
}

// Run executes the complete application lifecycle:
//  1. Runs all initializers sequentially
//  2. Starts all services concurrently
//  3. Waits for termination signal or service error
//  4. Performs graceful shutdown
//
// This method blocks until shutdown completes or an error occurs.
//
// Parameters:
//   - ctx: Root context for the application
//
// Returns:
//   - error: Returns error if initialization, service startup, or shutdown fails
//
// Signal handling:
//   - SIGINT (Ctrl+C): Triggers graceful shutdown
//   - SIGTERM: Triggers graceful shutdown (from orchestrators)
//
// Example:
//
//	if err := launcher.Run(context.Background()); err != nil {
//	    log.Error("Application failed", log.Field{Key: "error", Value: err})
//	    os.Exit(1)
//	}
func (l *Launcher) Run(ctx context.Context) error {
	// Phase 1: Run initializers
	if err := l.Init(ctx); err != nil {
		return fmt.Errorf("initialization failed: %w", err)
	}

	// Phase 2: Setup signal handling for graceful shutdown
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Phase 3: Start all services concurrently
	if err := l.startServices(ctx); err != nil {
		return fmt.Errorf("service startup failed: %w", err)
	}

	return nil
}

// startServices launches all registered services concurrently using errgroup.
// If any service fails to start or crashes, all other services are stopped.
func (l *Launcher) startServices(ctx context.Context) error {
	l.logger.Info("Starting services",
		log.Field{Key: "service_count", Value: len(l.services)})

	g, ctx := errgroup.WithContext(ctx)

	// Start each service in its own goroutine
	for i, svc := range l.services {
		// Capture loop variables
		index := i
		service := svc

		g.Go(func() error {
			l.logger.Info("Starting service",
				log.Field{Key: "index", Value: index},
				log.Field{Key: "type", Value: fmt.Sprintf("%T", service)})

			if err := service.Start(ctx); err != nil {
				return fmt.Errorf("service %d (%T) failed: %w", index, service, err)
			}
			return nil
		})
	}

	// Wait for all services to complete or context cancellation
	if err := g.Wait(); err != nil {
		// Context cancellation is expected during shutdown
		if ctx.Err() != nil {
			l.logger.Info("Services stopped due to context cancellation")
			return l.shutdown()
		}
		// Actual service error - perform emergency shutdown
		l.logger.Error("Service error occurred", log.Field{Key: "error", Value: err})
		return l.shutdown()
	}

	return nil
}

// shutdown performs graceful shutdown of all services with timeout.
// Services are stopped in reverse registration order for proper dependency cleanup.
func (l *Launcher) shutdown() error {
	l.logger.Info("Initiating graceful shutdown",
		log.Field{Key: "timeout", Value: l.shutdownTimeout})

	// Create context with timeout for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), l.shutdownTimeout)
	defer cancel()

	// Stop services in reverse order
	for i := len(l.services) - 1; i >= 0; i-- {
		svc := l.services[i]
		l.logger.Debug("Stopping service",
			log.Field{Key: "index", Value: i},
			log.Field{Key: "type", Value: fmt.Sprintf("%T", svc)})

		if err := svc.Stop(ctx); err != nil {
			l.logger.Error("Failed to stop service",
				log.Field{Key: "index", Value: i},
				log.Field{Key: "error", Value: err})
			// Continue stopping other services despite error
		}
	}

	l.logger.Info("Graceful shutdown completed")
	return nil
}
