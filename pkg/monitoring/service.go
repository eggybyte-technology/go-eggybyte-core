package monitoring

import (
	"context"
	"encoding/json"
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

// HealthChecker defines the interface for health check implementations.
type HealthChecker interface {
	// Name returns the identifier for this health checker.
	Name() string

	// Check performs the health check and returns error if unhealthy.
	Check(ctx context.Context) error
}

// MonitoringService provides both metrics and health check endpoints
// on a single port, following Kubernetes best practices.
//
// Endpoints:
//   - /metrics: Prometheus metrics exposition
//   - /healthz: Combined health check
//   - /livez: Liveness probe
//   - /readyz: Readiness probe
//
// Implements the service.Service interface for integration with core.Bootstrap.
type MonitoringService struct {
	port     int
	server   *http.Server
	logger   log.Logger
	checkers []HealthChecker
	mu       sync.RWMutex
}

// NewMonitoringService creates a new unified monitoring service.
//
// Parameters:
//   - port: HTTP port to listen on (typically 9090)
//
// Returns:
//   - *MonitoringService: Service instance ready for registration
func NewMonitoringService(port int) *MonitoringService {
	return &MonitoringService{
		port:     port,
		logger:   log.Default(),
		checkers: make([]HealthChecker, 0),
	}
}

// AddHealthChecker registers a health checker.
func (m *MonitoringService) AddHealthChecker(checker HealthChecker) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.checkers = append(m.checkers, checker)
}

// Start begins the monitoring HTTP server.
func (m *MonitoringService) Start(ctx context.Context) error {
	// Register Prometheus collectors once
	metricsOnce.Do(func() {
		if err := prometheus.Register(prometheus.NewGoCollector()); err != nil {
			m.logger.Debug("Go collector already registered",
				log.Field{Key: "error", Value: err.Error()})
		}
	})

	// Setup HTTP routes
	mux := http.NewServeMux()

	// Metrics endpoint
	mux.Handle("/metrics", promhttp.Handler())

	// Health endpoints
	mux.HandleFunc("/healthz", m.handleHealthz)
	mux.HandleFunc("/livez", m.handleLivez)
	mux.HandleFunc("/readyz", m.handleReadyz)

	// Create HTTP server
	m.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", m.port),
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	m.logger.Info("Starting monitoring server",
		log.Field{Key: "port", Value: m.port},
		log.Field{Key: "endpoints", Value: "/metrics, /healthz, /livez, /readyz"})

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
		return fmt.Errorf("monitoring server failed: %w", err)
	case <-ctx.Done():
		return m.Stop(context.Background())
	}
}

// Stop performs graceful shutdown.
func (m *MonitoringService) Stop(ctx context.Context) error {
	if m.server == nil {
		return nil
	}

	m.logger.Info("Stopping monitoring server")

	if err := m.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown monitoring server: %w", err)
	}

	m.logger.Info("Monitoring server stopped successfully")
	return nil
}

// handleLivez handles liveness probe requests.
func (m *MonitoringService) handleLivez(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// handleReadyz handles readiness probe requests.
func (m *MonitoringService) handleReadyz(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	m.mu.RLock()
	checkers := m.checkers
	m.mu.RUnlock()

	results := make(map[string]string)
	healthy := true

	for _, checker := range checkers {
		if err := checker.Check(ctx); err != nil {
			results[checker.Name()] = fmt.Sprintf("FAIL: %v", err)
			healthy = false
		} else {
			results[checker.Name()] = "OK"
		}
	}

	w.Header().Set("Content-Type", "application/json")

	if healthy {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": healthy,
		"checks": results,
	})
}

// handleHealthz handles combined health check requests.
func (m *MonitoringService) handleHealthz(w http.ResponseWriter, r *http.Request) {
	m.handleReadyz(w, r)
}
