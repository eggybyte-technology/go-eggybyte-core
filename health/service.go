package health

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/eggybyte-technology/go-eggybyte-core/log"
)

// HealthChecker defines the interface for health check implementations.
// Health checkers verify the operational status of dependencies like
// databases, caches, and external services.
type HealthChecker interface {
	// Name returns the identifier for this health checker.
	// Used in health check response JSON.
	Name() string

	// Check performs the health check and returns error if unhealthy.
	// Should complete quickly (< 1 second) to avoid timeout.
	Check(ctx context.Context) error
}

// HealthService implements service.Service interface to provide health endpoints.
// It exposes standard Kubernetes health check endpoints on the metrics port.
//
// Endpoints:
//   - /healthz: Combined health check
//   - /livez: Liveness probe (always returns 200 when service is running)
//   - /readyz: Readiness probe (returns 200 only when all checks pass)
type HealthService struct {
	port     int
	server   *http.Server
	logger   log.Logger
	checkers []HealthChecker
	mu       sync.RWMutex
}

// NewHealthService creates a new health service on the specified port.
//
// Parameters:
//   - port: HTTP port to listen on (typically same as metrics port)
//
// Returns:
//   - *HealthService: Service instance ready for registration
func NewHealthService(port int) *HealthService {
	return &HealthService{
		port:     port,
		logger:   log.Default(),
		checkers: make([]HealthChecker, 0),
	}
}

// AddChecker registers a health checker to be included in readiness checks.
//
// Parameters:
//   - checker: Health checker implementation
func (h *HealthService) AddChecker(checker HealthChecker) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.checkers = append(h.checkers, checker)
}

// Start begins the health check HTTP server and blocks until stopped.
//
// Parameters:
//   - ctx: Context for cancellation
//
// Returns:
//   - error: Returns error if server fails to start
func (h *HealthService) Start(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", h.handleHealthz)
	mux.HandleFunc("/livez", h.handleLivez)
	mux.HandleFunc("/readyz", h.handleReadyz)

	h.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", h.port),
		Handler: mux,
	}

	h.logger.Info("Starting health server",
		log.Field{Key: "port", Value: h.port})

	errCh := make(chan error, 1)
	go func() {
		if err := h.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return fmt.Errorf("health server failed: %w", err)
	case <-ctx.Done():
		return h.Stop(context.Background())
	}
}

// Stop performs graceful shutdown of the health server.
//
// Parameters:
//   - ctx: Context with timeout for shutdown
//
// Returns:
//   - error: Returns error if shutdown fails
func (h *HealthService) Stop(ctx context.Context) error {
	if h.server == nil {
		return nil
	}

	h.logger.Info("Stopping health server")
	return h.server.Shutdown(ctx)
}

// handleLivez handles liveness probe requests.
// Always returns 200 OK when service is running.
func (h *HealthService) handleLivez(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// handleReadyz handles readiness probe requests.
// Returns 200 only if all health checkers pass.
func (h *HealthService) handleReadyz(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	h.mu.RLock()
	checkers := h.checkers
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
func (h *HealthService) handleHealthz(w http.ResponseWriter, r *http.Request) {
	h.handleReadyz(w, r)
}
