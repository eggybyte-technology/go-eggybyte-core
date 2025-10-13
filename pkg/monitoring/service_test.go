package monitoring

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// mockHealthChecker is a test implementation of HealthChecker
type mockHealthChecker struct {
	name    string
	healthy bool
	err     error
}

func (m *mockHealthChecker) Name() string {
	return m.name
}

func (m *mockHealthChecker) Check(ctx context.Context) error {
	if m.err != nil {
		return m.err
	}
	if !m.healthy {
		return assert.AnError
	}
	return nil
}

func TestNewMonitoringService(t *testing.T) {
	service := NewMonitoringService(9090)

	assert.NotNil(t, service)
	assert.Equal(t, 9090, service.port)
	assert.NotNil(t, service.logger)
	assert.Empty(t, service.checkers)
}

func TestAddHealthChecker(t *testing.T) {
	service := NewMonitoringService(9090)
	checker := &mockHealthChecker{name: "test", healthy: true}

	service.AddHealthChecker(checker)

	assert.Len(t, service.checkers, 1)
	assert.Equal(t, "test", service.checkers[0].Name())
}

func TestMonitoringService_Start_Stop(t *testing.T) {
	service := NewMonitoringService(0) // Use port 0 for testing

	// Test that we can stop without starting
	err := service.Stop(context.Background())
	assert.NoError(t, err)

	// Test basic functionality
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start service in goroutine
	errCh := make(chan error, 1)
	go func() {
		errCh <- service.Start(ctx)
	}()

	// Give it a moment to start
	time.Sleep(10 * time.Millisecond)

	// Cancel context to stop service
	cancel()

	// Wait for service to stop
	select {
	case err := <-errCh:
		assert.NoError(t, err) // Should stop gracefully
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Service did not stop within timeout")
	}
}

func TestMonitoringService_Endpoints(t *testing.T) {
	service := NewMonitoringService(0)

	// Add a healthy checker
	healthyChecker := &mockHealthChecker{name: "healthy", healthy: true}
	service.AddHealthChecker(healthyChecker)

	// Add an unhealthy checker
	unhealthyChecker := &mockHealthChecker{name: "unhealthy", healthy: false}
	service.AddHealthChecker(unhealthyChecker)

	// Test /livez endpoint
	req := httptest.NewRequest("GET", "/livez", nil)
	w := httptest.NewRecorder()
	service.handleLivez(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "OK", w.Body.String())

	// Test /readyz endpoint
	req = httptest.NewRequest("GET", "/readyz", nil)
	w = httptest.NewRecorder()
	service.handleReadyz(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)

	// Test /healthz endpoint (should behave like /readyz)
	req = httptest.NewRequest("GET", "/healthz", nil)
	w = httptest.NewRecorder()
	service.handleHealthz(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}

func TestMonitoringService_HealthyEndpoints(t *testing.T) {
	service := NewMonitoringService(0)

	// Add only healthy checkers
	healthyChecker := &mockHealthChecker{name: "healthy", healthy: true}
	service.AddHealthChecker(healthyChecker)

	// Test /readyz endpoint with healthy checkers
	req := httptest.NewRequest("GET", "/readyz", nil)
	w := httptest.NewRecorder()
	service.handleReadyz(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Test /healthz endpoint with healthy checkers
	req = httptest.NewRequest("GET", "/healthz", nil)
	w = httptest.NewRecorder()
	service.handleHealthz(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestMonitoringService_NoCheckers(t *testing.T) {
	service := NewMonitoringService(0)

	// Test /readyz endpoint with no checkers
	req := httptest.NewRequest("GET", "/readyz", nil)
	w := httptest.NewRecorder()
	service.handleReadyz(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestMonitoringService_StopWithoutStart(t *testing.T) {
	service := NewMonitoringService(9090)

	// Stop without starting should not error
	err := service.Stop(context.Background())
	assert.NoError(t, err)
}
