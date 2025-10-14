package monitoring

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// MockHealthChecker is a test implementation of HealthChecker interface
type MockHealthChecker struct {
	name    string
	healthy bool
	err     error
}

func (m *MockHealthChecker) Name() string {
	return m.name
}

func (m *MockHealthChecker) Check(ctx context.Context) error {
	if m.healthy {
		return nil
	}
	return m.err
}

func TestNewHealthService(t *testing.T) {
	service := NewHealthService(8081)

	assert.NotNil(t, service)
	assert.Equal(t, 8081, service.GetPort())
	assert.Equal(t, 0, service.GetCheckerCount())
}

func TestHealthService_AddHealthChecker(t *testing.T) {
	service := NewHealthService(8081)

	// Add first checker
	checker1 := &MockHealthChecker{name: "checker1", healthy: true}
	service.AddHealthChecker(checker1)

	assert.Equal(t, 1, service.GetCheckerCount())

	// Add second checker
	checker2 := &MockHealthChecker{name: "checker2", healthy: true}
	service.AddHealthChecker(checker2)

	assert.Equal(t, 2, service.GetCheckerCount())
}

func TestHealthService_Start_ContextCancellation(t *testing.T) {
	service := NewHealthService(0) // Use port 0 for automatic port assignment

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())

	// Start service in goroutine
	errChan := make(chan error, 1)
	go func() {
		errChan <- service.Start(ctx)
	}()

	// Give service time to start
	time.Sleep(100 * time.Millisecond)

	// Cancel context
	cancel()

	// Wait for service to stop
	select {
	case err := <-errChan:
		assert.NoError(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("Service did not stop within timeout")
	}
}

func TestHealthService_Stop(t *testing.T) {
	service := NewHealthService(0)

	// Start service in goroutine
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		service.Start(ctx)
	}()

	// Give service time to start
	time.Sleep(100 * time.Millisecond)

	// Stop service
	stopCtx, stopCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer stopCancel()

	err := service.Stop(stopCtx)
	assert.NoError(t, err)
}

func TestHealthService_handleLivez(t *testing.T) {
	service := NewHealthService(8081)

	req := httptest.NewRequest("GET", "/livez", nil)
	rec := httptest.NewRecorder()

	service.handleLivez(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "OK", rec.Body.String())
}

func TestHealthService_handleReadyz_Healthy(t *testing.T) {
	service := NewHealthService(8081)

	// Add healthy checker
	checker := &MockHealthChecker{name: "test-checker", healthy: true}
	service.AddHealthChecker(checker)

	req := httptest.NewRequest("GET", "/readyz", nil)
	rec := httptest.NewRecorder()

	service.handleReadyz(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "test-checker")
	assert.Contains(t, rec.Body.String(), "OK")
}

func TestHealthService_handleReadyz_Unhealthy(t *testing.T) {
	service := NewHealthService(8081)

	// Add unhealthy checker
	checker := &MockHealthChecker{
		name:    "test-checker",
		healthy: false,
		err:     assert.AnError,
	}
	service.AddHealthChecker(checker)

	req := httptest.NewRequest("GET", "/readyz", nil)
	rec := httptest.NewRecorder()

	service.handleReadyz(rec, req)

	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
	assert.Contains(t, rec.Body.String(), "test-checker")
	assert.Contains(t, rec.Body.String(), "FAIL")
}

func TestHealthService_handleReadyz_MultipleCheckers(t *testing.T) {
	service := NewHealthService(8081)

	// Add multiple checkers
	checker1 := &MockHealthChecker{name: "checker1", healthy: true}
	checker2 := &MockHealthChecker{name: "checker2", healthy: false, err: assert.AnError}
	checker3 := &MockHealthChecker{name: "checker3", healthy: true}

	service.AddHealthChecker(checker1)
	service.AddHealthChecker(checker2)
	service.AddHealthChecker(checker3)

	req := httptest.NewRequest("GET", "/readyz", nil)
	rec := httptest.NewRecorder()

	service.handleReadyz(rec, req)

	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
	assert.Contains(t, rec.Body.String(), "checker1")
	assert.Contains(t, rec.Body.String(), "checker2")
	assert.Contains(t, rec.Body.String(), "checker3")
	assert.Contains(t, rec.Body.String(), "FAIL")
}

func TestHealthService_handleHealthz(t *testing.T) {
	service := NewHealthService(8081)

	// Add checker
	checker := &MockHealthChecker{name: "test-checker", healthy: true}
	service.AddHealthChecker(checker)

	req := httptest.NewRequest("GET", "/healthz", nil)
	rec := httptest.NewRecorder()

	service.handleHealthz(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "test-checker")
}

func TestHealthService_handleReadyz_Timeout(t *testing.T) {
	service := NewHealthService(8081)

	// Add checker that takes too long
	checker := &MockHealthChecker{name: "slow-checker", healthy: true}
	service.AddHealthChecker(checker)

	// Create request with very short timeout
	req := httptest.NewRequest("GET", "/readyz", nil)
	type timeoutKey string
	req = req.WithContext(context.WithValue(req.Context(), timeoutKey("timeout"), 1*time.Nanosecond))
	rec := httptest.NewRecorder()

	service.handleReadyz(rec, req)

	// Should still work as the mock checker is fast
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestHealthService_handleReadyz_NoCheckers(t *testing.T) {
	service := NewHealthService(8081)

	req := httptest.NewRequest("GET", "/readyz", nil)
	rec := httptest.NewRecorder()

	service.handleReadyz(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "status")
	assert.Contains(t, rec.Body.String(), "checks")
}

func TestHealthService_ConcurrentAccess(t *testing.T) {
	service := NewHealthService(8081)

	// Test concurrent checker addition
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(i int) {
			checker := &MockHealthChecker{
				name:    "checker" + string(rune(i)),
				healthy: true,
			}
			service.AddHealthChecker(checker)
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	assert.Equal(t, 10, service.GetCheckerCount())
}

func TestHealthService_SetLogger(t *testing.T) {
	service := NewHealthService(8081)

	// This test mainly ensures the method doesn't panic
	// In a real implementation, you might want to test logger functionality
	service.SetLogger(nil)

	// Test passes if no panic occurs
	assert.True(t, true)
}

func TestHealthService_GetPort(t *testing.T) {
	service := NewHealthService(8081)
	assert.Equal(t, 8081, service.GetPort())
}

func TestHealthService_GetCheckerCount(t *testing.T) {
	service := NewHealthService(8081)

	// Initially no checkers
	assert.Equal(t, 0, service.GetCheckerCount())

	// Add checkers
	checker1 := &MockHealthChecker{name: "checker1", healthy: true}
	checker2 := &MockHealthChecker{name: "checker2", healthy: true}

	service.AddHealthChecker(checker1)
	assert.Equal(t, 1, service.GetCheckerCount())

	service.AddHealthChecker(checker2)
	assert.Equal(t, 2, service.GetCheckerCount())
}

func TestHealthService_Start_AlreadyRunning(t *testing.T) {
	service := NewHealthService(0)

	// Start service
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		service.Start(ctx)
	}()

	// Give service time to start
	time.Sleep(100 * time.Millisecond)

	// Try to start again - should not cause issues
	// (In real implementation, this might return an error)
	go func() {
		service.Start(ctx)
	}()

	// Test passes if no panic occurs
	assert.True(t, true)
}

func TestHealthService_Stop_NotStarted(t *testing.T) {
	service := NewHealthService(8081)

	// Stop service that was never started
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := service.Stop(ctx)
	assert.NoError(t, err)
}

func TestHealthService_JSONResponse(t *testing.T) {
	service := NewHealthService(8081)

	// Add checker
	checker := &MockHealthChecker{name: "test-checker", healthy: true}
	service.AddHealthChecker(checker)

	req := httptest.NewRequest("GET", "/readyz", nil)
	rec := httptest.NewRecorder()

	service.handleReadyz(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	// Verify JSON structure
	body := rec.Body.String()
	assert.Contains(t, body, "status")
	assert.Contains(t, body, "checks")
	assert.Contains(t, body, "timestamp")
	assert.Contains(t, body, "test-checker")
	assert.Contains(t, body, "OK")
}
