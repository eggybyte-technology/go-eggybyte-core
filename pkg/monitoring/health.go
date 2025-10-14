// Package monitoring provides health check and metrics exposition for EggyByte services.
// It includes Kubernetes-compatible health check endpoints and Prometheus metrics
// exposition with support for custom health checkers and metrics collectors.
package monitoring

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/eggybyte-technology/go-eggybyte-core/pkg/log"
)

// HealthChecker defines the interface for health check implementations.
type HealthChecker interface {
	// Name returns the identifier for this health checker.
	Name() string

	// Check performs the health check and returns error if unhealthy.
	Check(ctx context.Context) error
}

// HealthService provides Kubernetes-compatible health check endpoints
// on a dedicated port for security and monitoring isolation.
//
// Endpoints:
//   - /healthz: Combined health check with all registered checkers
//   - /livez: Liveness probe (always returns OK if service is running)
//   - /readyz: Readiness probe with dependency checks
//
// Thread Safety: HealthService is safe for concurrent use after initialization.
// Health checkers can be added concurrently and checks are performed safely.
//
// Implements the service.Service interface for integration with core.Bootstrap.
type HealthService struct {
	// port is the HTTP port for health check endpoints
	port int

	// server is the underlying HTTP server instance
	server *http.Server

	// logger is the structured logger for this service
	logger log.Logger

	// checkers holds all registered health checkers
	checkers []HealthChecker

	// mu protects concurrent access to checkers slice
	mu sync.RWMutex

	// serverMu protects concurrent access to server field
	serverMu sync.RWMutex
}

// NewHealthService creates a new health check service with the specified port.
// This service provides Kubernetes-compatible health check endpoints.
//
// Parameters:
//   - port: HTTP port to listen on for health check endpoints
//
// Returns:
//   - *HealthService: Service instance ready for registration
//
// Example:
//
//	healthService := NewHealthService(8081)
//	healthService.AddHealthChecker(databaseChecker)
//	launcher.AddService(healthService)
func NewHealthService(port int) *HealthService {
	return &HealthService{
		port:     port,
		logger:   log.Default(),
		checkers: make([]HealthChecker, 0),
	}
}

// AddHealthChecker registers a health checker with this service.
// Health checkers are used by the /readyz endpoint to verify service dependencies.
//
// Parameters:
//   - checker: Health checker implementing the HealthChecker interface
//
// Thread Safety: This method is safe for concurrent use.
//
// Example:
//
//	healthService.AddHealthChecker(&DatabaseHealthChecker{})
//	healthService.AddHealthChecker(&RedisHealthChecker{})
func (h *HealthService) AddHealthChecker(checker HealthChecker) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.checkers = append(h.checkers, checker)

	h.logger.Info("Health checker registered",
		log.Field{Key: "name", Value: checker.Name()},
		log.Field{Key: "total_checkers", Value: len(h.checkers)})
}

// Start begins the health check HTTP server.
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
//   - Creates HTTP server with health check endpoints
//   - Logs server startup information
//   - Handles graceful shutdown on context cancellation
//   - Returns immediately if server is already running
//
// Example:
//
//	ctx, cancel := context.WithCancel(context.Background())
//	go func() {
//	    if err := healthService.Start(ctx); err != nil {
//	        log.Error("Health service failed", log.Field{Key: "error", Value: err})
//	    }
//	}()
//	defer cancel()
func (h *HealthService) Start(ctx context.Context) error {
	// Setup HTTP routes
	mux := http.NewServeMux()

	// Health check endpoints
	mux.HandleFunc("/healthz", h.handleHealthz)
	mux.HandleFunc("/livez", h.handleLivez)
	mux.HandleFunc("/readyz", h.handleReadyz)

	// Create HTTP server
	h.serverMu.Lock()
	h.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", h.port),
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
	h.serverMu.Unlock()

	h.logger.Info("Starting health check server",
		log.Field{Key: "port", Value: h.port},
		log.Field{Key: "endpoints", Value: "/healthz, /livez, /readyz"})

	// Create a channel to receive server errors
	errChan := make(chan error, 1)

	// Start server in a goroutine
	go func() {
		h.serverMu.RLock()
		server := h.server
		h.serverMu.RUnlock()

		if server != nil {
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				errChan <- fmt.Errorf("health check server failed: %w", err)
			}
		}
	}()

	// Wait for context cancellation or server error
	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		h.logger.Info("Shutting down health check server",
			log.Field{Key: "reason", Value: "context_canceled"})

		// Create shutdown context with timeout
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Attempt graceful shutdown
		h.serverMu.RLock()
		server := h.server
		h.serverMu.RUnlock()

		if server != nil {
			if err := server.Shutdown(shutdownCtx); err != nil {
				h.logger.Error("Health check server shutdown failed",
					log.Field{Key: "error", Value: err})
				return fmt.Errorf("health check server shutdown failed: %w", err)
			}
		}

		h.logger.Info("Health check server shutdown completed")
		return nil
	}
}

// Stop gracefully shuts down the health check server.
// This method is provided for compatibility with the Service interface.
// In practice, shutdown is handled by the Start method when context is canceled.
//
// Parameters:
//   - ctx: Context for timeout control during shutdown
//
// Returns:
//   - error: Returns error if shutdown fails
func (h *HealthService) Stop(ctx context.Context) error {
	h.serverMu.RLock()
	server := h.server
	h.serverMu.RUnlock()

	if server == nil {
		return nil
	}

	h.logger.Info("Stopping health check server")

	if err := server.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown health check server: %w", err)
	}

	h.logger.Info("Health check server stopped successfully")
	return nil
}

// GetPort returns the configured port for this service.
// This method is useful for logging and monitoring purposes.
//
// Returns:
//   - int: The port number
func (h *HealthService) GetPort() int {
	return h.port
}

// GetCheckerCount returns the number of registered health checkers.
// This method is useful for monitoring and debugging purposes.
//
// Returns:
//   - int: The number of registered health checkers
func (h *HealthService) GetCheckerCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.checkers)
}

// SetLogger sets the logger for this health service.
// This method allows customization of logging behavior.
//
// Parameters:
//   - logger: The logger instance to use for this service
func (h *HealthService) SetLogger(logger interface{}) {
	if l := log.SetLoggerHelper(logger); l != nil {
		h.logger = l
	}
}

// handleLivez handles liveness probe requests.
// This endpoint always returns OK if the service is running,
// indicating that the service process is alive.
//
// Parameters:
//   - w: HTTP response writer
//   - r: HTTP request
func (h *HealthService) handleLivez(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("OK")); err != nil {
		h.logger.Error("Failed to write livez response", log.Field{Key: "error", Value: err})
	}
}

// handleReadyz handles readiness probe requests.
// This endpoint checks all registered health checkers to determine
// if the service is ready to accept traffic.
//
// Parameters:
//   - w: HTTP response writer
//   - r: HTTP request
//
// Behavior:
//   - Runs all registered health checkers with timeout
//   - Returns 200 OK if all checkers pass
//   - Returns 503 Service Unavailable if any checker fails
//   - Includes detailed results in JSON response
func (h *HealthService) handleReadyz(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	h.mu.RLock()
	checkers := make([]HealthChecker, len(h.checkers))
	copy(checkers, h.checkers)
	h.mu.RUnlock()

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

	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    healthy,
		"checks":    results,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}); err != nil {
		h.logger.Error("Failed to encode health response", log.Field{Key: "error", Value: err})
	}
}

// handleHealthz handles combined health check requests.
// This endpoint provides the same functionality as /readyz for compatibility.
//
// Parameters:
//   - w: HTTP response writer
//   - r: HTTP request
func (h *HealthService) handleHealthz(w http.ResponseWriter, r *http.Request) {
	h.handleReadyz(w, r)
}
