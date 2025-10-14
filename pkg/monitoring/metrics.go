// Package monitoring provides health check and metrics exposition for EggyByte services.

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/eggybyte-technology/go-eggybyte-core/pkg/log"
)


// MetricsService provides Prometheus metrics exposition on a dedicated port
// for security and monitoring isolation.
//
// Endpoints:
//   - /metrics: Prometheus metrics exposition format
//
// Thread Safety: MetricsService is safe for concurrent use after initialization.
// Metrics collectors are registered once and can be safely accessed concurrently.
//
// Implements the service.Service interface for integration with core.Bootstrap.
type MetricsService struct {
	// port is the HTTP port for metrics exposition
	port int

	// server is the underlying HTTP server instance
	server *http.Server

	// logger is the structured logger for this service
	logger log.Logger

	// registry is the Prometheus metrics registry
	registry *prometheus.Registry

	// serverMu protects concurrent access to server field
	serverMu sync.RWMutex
}

// NewMetricsService creates a new metrics exposition service with the specified port.
// This service provides Prometheus-compatible metrics endpoints.
//
// Parameters:
//   - port: HTTP port to listen on for metrics exposition
//
// Returns:
//   - *MetricsService: Service instance ready for registration
//
// Example:
//
//	metricsService := NewMetricsService(9091)
//	launcher.AddService(metricsService)
func NewMetricsService(port int) *MetricsService {
	// Create a new Prometheus registry
	registry := prometheus.NewRegistry()

	// Register default collectors
	registry.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)

	return &MetricsService{
		port:     port,
		logger:   log.Default(),
		registry: registry,
	}
}

// NewMetricsServiceWithRegistry creates a new metrics service with a custom registry.
// This constructor allows fine-grained control over metrics collection.
//
// Parameters:
//   - port: HTTP port to listen on for metrics exposition
//   - registry: Custom Prometheus registry
//
// Returns:
//   - *MetricsService: Service instance with custom registry
//
// Example:
//
//	registry := prometheus.NewRegistry()
//	registry.MustRegister(customCollector)
//	metricsService := NewMetricsServiceWithRegistry(9091, registry)
func NewMetricsServiceWithRegistry(port int, registry *prometheus.Registry) *MetricsService {
	return &MetricsService{
		port:     port,
		logger:   log.Default(),
		registry: registry,
	}
}

// RegisterCollector registers a Prometheus collector with the metrics registry.
// This method allows adding custom metrics collectors to the service.
//
// Parameters:
//   - collector: Prometheus collector to register
//
// Returns:
//   - error: Returns error if registration fails
//
// Thread Safety: This method is safe for concurrent use.
//
// Example:
//
//	counter := prometheus.NewCounterVec(
//	    prometheus.CounterOpts{
//	        Name: "http_requests_total",
//	        Help: "Total number of HTTP requests",
//	    },
//	    []string{"method", "endpoint"},
//	)
//	metricsService.RegisterCollector(counter)
func (m *MetricsService) RegisterCollector(collector prometheus.Collector) error {
	if collector == nil {
		return fmt.Errorf("nil collector")
	}

	if err := m.registry.Register(collector); err != nil {
		m.logger.Error("Failed to register metrics collector",
			log.Field{Key: "error", Value: err})
		return fmt.Errorf("failed to register metrics collector: %w", err)
	}

	m.logger.Info("Metrics collector registered successfully")
	return nil
}

// UnregisterCollector removes a Prometheus collector from the metrics registry.
// This method allows removing metrics collectors dynamically.
//
// Parameters:
//   - collector: Prometheus collector to unregister
//
// Returns:
//   - bool: True if collector was successfully unregistered
//
// Thread Safety: This method is safe for concurrent use.
func (m *MetricsService) UnregisterCollector(collector prometheus.Collector) bool {
	success := m.registry.Unregister(collector)
	if success {
		m.logger.Info("Metrics collector unregistered successfully")
	} else {
		m.logger.Warn("Failed to unregister metrics collector - collector not found")
	}
	return success
}

// Start begins the metrics HTTP server.
// This method blocks until the server is stopped or encounters an error.
// It should be called in a goroutine for non-blocking operation.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//
// Returns:
//   - error: Returns error if server fails to start or encounters a fatal error
//
// Behavior:
//   - Creates HTTP server with metrics exposition endpoint
//   - Logs server startup information
//   - Handles graceful shutdown on context cancellation
//   - Returns immediately if server is already running
//
// Example:
//
//	ctx, cancel := context.WithCancel(context.Background())
//	go func() {
//	    if err := metricsService.Start(ctx); err != nil {
//	        log.Error("Metrics service failed", log.Field{Key: "error", Value: err})
//	    }
//	}()
//	defer cancel()
func (m *MetricsService) Start(ctx context.Context) error {
	// Setup HTTP routes
	mux := http.NewServeMux()

	// Metrics endpoint with custom registry
	mux.Handle("/metrics", promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{
		EnableOpenMetrics: true,
		Timeout:           10 * time.Second,
	}))

	// Create HTTP server
	m.serverMu.Lock()
	m.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", m.port),
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second, // Longer timeout for metrics collection
		IdleTimeout:  120 * time.Second,
	}
	m.serverMu.Unlock()

	m.logger.Info("Starting metrics server",
		log.Field{Key: "port", Value: m.port},
		log.Field{Key: "endpoints", Value: "/metrics"})

	// Create a channel to receive server errors
	errChan := make(chan error, 1)

	// Start server in a goroutine
	go func() {
		m.serverMu.RLock()
		server := m.server
		m.serverMu.RUnlock()

		if server != nil {
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				errChan <- fmt.Errorf("metrics server failed: %w", err)
			}
		}
	}()

	// Wait for context cancellation or server error
	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		m.logger.Info("Shutting down metrics server",
			log.Field{Key: "reason", Value: "context_canceled"})

		// Create shutdown context with timeout
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Attempt graceful shutdown
		m.serverMu.RLock()
		server := m.server
		m.serverMu.RUnlock()

		if server != nil {
			if err := server.Shutdown(shutdownCtx); err != nil {
				m.logger.Error("Metrics server shutdown failed",
					log.Field{Key: "error", Value: err})
				return fmt.Errorf("metrics server shutdown failed: %w", err)
			}
		}

		m.logger.Info("Metrics server shutdown completed")
		return nil
	}
}

// Stop gracefully shuts down the metrics server.
// This method is provided for compatibility with the Service interface.
// In practice, shutdown is handled by the Start method when context is canceled.
//
// Parameters:
//   - ctx: Context for timeout control during shutdown
//
// Returns:
//   - error: Returns error if shutdown fails
func (m *MetricsService) Stop(ctx context.Context) error {
	m.serverMu.RLock()
	server := m.server
	m.serverMu.RUnlock()

	if server == nil {
		return nil
	}

	m.logger.Info("Stopping metrics server")

	if err := server.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown metrics server: %w", err)
	}

	m.logger.Info("Metrics server stopped successfully")
	return nil
}

// GetPort returns the configured port for this service.
// This method is useful for logging and monitoring purposes.
//
// Returns:
//   - int: The port number
func (m *MetricsService) GetPort() int {
	return m.port
}

// GetRegistry returns the Prometheus metrics registry.
// This method provides access to the underlying registry for advanced use cases.
//
// Returns:
//   - *prometheus.Registry: The Prometheus registry instance
//
// Example:
//
//	registry := metricsService.GetRegistry()
//	counter := prometheus.NewCounter(prometheus.CounterOpts{
//	    Name: "custom_metric_total",
//	    Help: "A custom metric",
//	})
//	registry.MustRegister(counter)
func (m *MetricsService) GetRegistry() *prometheus.Registry {
	return m.registry
}

// SetLogger sets the logger for this metrics service.
// This method allows customization of logging behavior.
//
// Parameters:
//   - logger: The logger instance to use for this service
func (m *MetricsService) SetLogger(logger interface{}) {
	if l, ok := logger.(log.Logger); ok {
		m.logger = l
	}
}

// GetCollectorCount returns the number of registered metrics collectors.
// This method is useful for monitoring and debugging purposes.
//
// Returns:
//   - int: The number of registered collectors
func (m *MetricsService) GetCollectorCount() int {
	// Note: Prometheus registry doesn't provide a direct way to count collectors
	// This is a placeholder for future implementation
	return 0
}
