package metrics

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/eggybyte-technology/go-eggybyte-core/pkg/log"
)

var (
	// metricsOnce ensures metrics collectors are registered only once
	metricsOnce sync.Once
)

// MetricsService implements service.Service interface to expose Prometheus metrics.
// It runs an HTTP server on a separate port (default 9090) to provide metrics
// collection without mixing with business API traffic.
//
// The service automatically registers default Go runtime metrics (goroutines,
// memory, GC) and provides an endpoint for custom application metrics.
//
// Usage:
//
//	metricsService := metrics.NewMetricsService(9090)
//	launcher.AddService(metricsService)
type MetricsService struct {
	port   int
	server *http.Server
	logger log.Logger
}

// NewMetricsService creates a new metrics service that listens on the specified port.
//
// Parameters:
//   - port: HTTP port to listen on (typically 9090 for Prometheus)
//
// Returns:
//   - *MetricsService: Service instance ready to be registered with launcher
//
// Example:
//
//	metricsService := metrics.NewMetricsService(9090)
//	launcher.AddService(metricsService)
func NewMetricsService(port int) *MetricsService {
	return &MetricsService{
		port:   port,
		logger: log.Default(),
	}
}

// Start begins the metrics HTTP server and blocks until stopped.
// Implements service.Service interface.
//
// The server exposes the following endpoints:
//   - /metrics: Prometheus metrics in text exposition format
//
// Parameters:
//   - ctx: Context for cancellation and shutdown coordination
//
// Returns:
//   - error: Returns error if server fails to start or bind to port
//
// Behavior:
//   - Registers default Go metrics collectors
//   - Starts HTTP server in separate goroutine
//   - Blocks until context is cancelled
//   - Performs graceful shutdown on cancellation
func (m *MetricsService) Start(ctx context.Context) error {
	// Register default Go runtime metrics (only once globally)
	metricsOnce.Do(func() {
		// Attempt to register Go collector
		// If already registered (e.g., from previous service restart), ignore the error
		if err := prometheus.Register(prometheus.NewGoCollector()); err != nil {
			m.logger.Debug("Go collector already registered, skipping",
				log.Field{Key: "error", Value: err.Error()})
		}
	})

	// Setup HTTP mux with metrics endpoint
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	// Create HTTP server
	m.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", m.port),
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	m.logger.Info("Starting metrics server",
		log.Field{Key: "port", Value: m.port},
		log.Field{Key: "endpoint", Value: "/metrics"})

	// Start server in goroutine
	errCh := make(chan error, 1)
	go func() {
		if err := m.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	// Wait for context cancellation or server error
	select {
	case err := <-errCh:
		return fmt.Errorf("metrics server failed: %w", err)
	case <-ctx.Done():
		return m.Stop(context.Background())
	}
}

// Stop performs graceful shutdown of the metrics server.
// Implements service.Service interface.
//
// Parameters:
//   - ctx: Context with timeout for shutdown completion
//
// Returns:
//   - error: Returns error if shutdown fails or times out
func (m *MetricsService) Stop(ctx context.Context) error {
	if m.server == nil {
		return nil
	}

	m.logger.Info("Stopping metrics server")

	if err := m.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown metrics server: %w", err)
	}

	m.logger.Info("Metrics server stopped successfully")
	return nil
}
