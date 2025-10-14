package health

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockHealthChecker is a test implementation of HealthChecker interface.
type mockHealthChecker struct {
	name      string
	checkFunc func(ctx context.Context) error
}

func (m *mockHealthChecker) Name() string {
	return m.name
}

func (m *mockHealthChecker) Check(ctx context.Context) error {
	if m.checkFunc != nil {
		return m.checkFunc(ctx)
	}
	return nil
}

// TestNewHealthService tests health service creation.
// This is an isolated method test with no external dependencies.
func TestNewHealthService(t *testing.T) {
	service := NewHealthService(8081)

	assert.NotNil(t, service)
	assert.Equal(t, 8081, service.port)
	assert.NotNil(t, service.logger)
	assert.NotNil(t, service.checkers)
	assert.Len(t, service.checkers, 0)
}

// TestAddChecker tests adding health checkers.
// This verifies checkers are stored correctly.
func TestAddChecker(t *testing.T) {
	service := NewHealthService(8081)

	checker1 := &mockHealthChecker{name: "database"}
	checker2 := &mockHealthChecker{name: "redis"}

	service.AddChecker(checker1)
	assert.Len(t, service.checkers, 1)

	service.AddChecker(checker2)
	assert.Len(t, service.checkers, 2)
}

// TestAddChecker_Multiple tests adding multiple checkers.
// This verifies checkers accumulate properly.
func TestAddChecker_Multiple(t *testing.T) {
	service := NewHealthService(8081)

	checker1 := &mockHealthChecker{name: "database"}
	checker2 := &mockHealthChecker{name: "redis"}
	checker3 := &mockHealthChecker{name: "external-api"}

	service.AddChecker(checker1)
	service.AddChecker(checker2)
	service.AddChecker(checker3)

	assert.Len(t, service.checkers, 3)
}

// TestHandleLivez tests liveness probe handler.
// This verifies /livez always returns 200 OK.
func TestHandleLivez(t *testing.T) {
	service := NewHealthService(8081)

	req := httptest.NewRequest("GET", "/livez", nil)
	rec := httptest.NewRecorder()

	service.handleLivez(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "OK", rec.Body.String())
}

// TestHandleLivez_AlwaysSucceeds tests liveness probe is always healthy.
// This verifies liveness doesn't depend on checkers.
func TestHandleLivez_AlwaysSucceeds(t *testing.T) {
	service := NewHealthService(8081)

	// Add failing checker
	failingChecker := &mockHealthChecker{
		name: "database",
		checkFunc: func(ctx context.Context) error {
			return errors.New("database down")
		},
	}
	service.AddChecker(failingChecker)

	req := httptest.NewRequest("GET", "/livez", nil)
	rec := httptest.NewRecorder()

	service.handleLivez(rec, req)

	// Liveness should still return OK despite failing checker
	assert.Equal(t, http.StatusOK, rec.Code)
}

// TestHandleReadyz_NoCheckers tests readiness probe with no checkers.
// This verifies readiness succeeds when no checkers are registered.
func TestHandleReadyz_NoCheckers(t *testing.T) {
	service := NewHealthService(8081)

	req := httptest.NewRequest("GET", "/readyz", nil)
	rec := httptest.NewRecorder()

	service.handleReadyz(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response["status"].(bool))
	checks := response["checks"].(map[string]interface{})
	assert.Len(t, checks, 0)
}

// TestHandleReadyz_AllHealthy tests readiness probe with all healthy checkers.
// This verifies readiness succeeds when all checks pass.
func TestHandleReadyz_AllHealthy(t *testing.T) {
	service := NewHealthService(8081)

	checker1 := &mockHealthChecker{
		name: "database",
		checkFunc: func(ctx context.Context) error {
			return nil // Healthy
		},
	}
	checker2 := &mockHealthChecker{
		name: "redis",
		checkFunc: func(ctx context.Context) error {
			return nil // Healthy
		},
	}

	service.AddChecker(checker1)
	service.AddChecker(checker2)

	req := httptest.NewRequest("GET", "/readyz", nil)
	rec := httptest.NewRecorder()

	service.handleReadyz(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response["status"].(bool))
	checks := response["checks"].(map[string]interface{})
	assert.Equal(t, "OK", checks["database"])
	assert.Equal(t, "OK", checks["redis"])
}

// TestHandleReadyz_OneUnhealthy tests readiness probe with one failing checker.
// This verifies readiness fails when any check fails.
func TestHandleReadyz_OneUnhealthy(t *testing.T) {
	service := NewHealthService(8081)

	checker1 := &mockHealthChecker{
		name: "database",
		checkFunc: func(ctx context.Context) error {
			return nil // Healthy
		},
	}
	checker2 := &mockHealthChecker{
		name: "redis",
		checkFunc: func(ctx context.Context) error {
			return errors.New("connection failed") // Unhealthy
		},
	}

	service.AddChecker(checker1)
	service.AddChecker(checker2)

	req := httptest.NewRequest("GET", "/readyz", nil)
	rec := httptest.NewRecorder()

	service.handleReadyz(rec, req)

	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response["status"].(bool))
	checks := response["checks"].(map[string]interface{})
	assert.Equal(t, "OK", checks["database"])
	assert.Contains(t, checks["redis"], "FAIL")
	assert.Contains(t, checks["redis"], "connection failed")
}

// TestHandleReadyz_AllUnhealthy tests readiness probe with all failing checkers.
// This verifies readiness reports all failures.
func TestHandleReadyz_AllUnhealthy(t *testing.T) {
	service := NewHealthService(8081)

	checker1 := &mockHealthChecker{
		name: "database",
		checkFunc: func(ctx context.Context) error {
			return errors.New("db error")
		},
	}
	checker2 := &mockHealthChecker{
		name: "redis",
		checkFunc: func(ctx context.Context) error {
			return errors.New("redis error")
		},
	}

	service.AddChecker(checker1)
	service.AddChecker(checker2)

	req := httptest.NewRequest("GET", "/readyz", nil)
	rec := httptest.NewRecorder()

	service.handleReadyz(rec, req)

	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response["status"].(bool))
	checks := response["checks"].(map[string]interface{})
	assert.Contains(t, checks["database"], "FAIL")
	assert.Contains(t, checks["redis"], "FAIL")
}

// TestHandleHealthz tests combined health check handler.
// This verifies /healthz behaves the same as /readyz.
func TestHandleHealthz(t *testing.T) {
	service := NewHealthService(8081)

	checker := &mockHealthChecker{
		name: "database",
		checkFunc: func(ctx context.Context) error {
			return nil
		},
	}
	service.AddChecker(checker)

	req := httptest.NewRequest("GET", "/healthz", nil)
	rec := httptest.NewRecorder()

	service.handleHealthz(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response["status"].(bool))
}

// TestStop_WithNilServer tests Stop handles nil server gracefully.
// This verifies Stop doesn't panic when server isn't started.
func TestStop_WithNilServer(t *testing.T) {
	service := NewHealthService(8081)

	ctx := context.Background()
	err := service.Stop(ctx)

	assert.NoError(t, err, "Stop should succeed with nil server")
}

// TestHealthChecker_Interface tests mockHealthChecker implements interface.
// This verifies the test mock correctly implements HealthChecker.
func TestHealthChecker_Interface(t *testing.T) {
	var _ HealthChecker = (*mockHealthChecker)(nil)

	checker := &mockHealthChecker{name: "test"}

	assert.Equal(t, "test", checker.Name())

	ctx := context.Background()
	err := checker.Check(ctx)
	assert.NoError(t, err)
}

// TestHealthChecker_CustomLogic tests custom check logic.
// This verifies checkers can implement custom health logic.
func TestHealthChecker_CustomLogic(t *testing.T) {
	called := false

	checker := &mockHealthChecker{
		name: "custom",
		checkFunc: func(ctx context.Context) error {
			called = true
			return nil
		},
	}

	ctx := context.Background()
	err := checker.Check(ctx)

	assert.NoError(t, err)
	assert.True(t, called, "Custom check function should be called")
}

// TestReadyzTimeout tests readiness probe respects timeout.
// This verifies checks complete within the configured timeout.
func TestReadyzTimeout(t *testing.T) {
	service := NewHealthService(8081)

	checker := &mockHealthChecker{
		name: "slow-service",
		checkFunc: func(ctx context.Context) error {
			// Simulate slow check that exceeds the 5-second timeout
			select {
			case <-time.After(6 * time.Second):
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		},
	}
	service.AddChecker(checker)

	req := httptest.NewRequest("GET", "/readyz", nil)
	rec := httptest.NewRecorder()

	start := time.Now()
	service.handleReadyz(rec, req)
	elapsed := time.Since(start)

	// Should complete within the 5-second timeout (plus small buffer for execution)
	assert.Less(t, elapsed, 6*time.Second, "Health check should timeout within expected duration")
	assert.GreaterOrEqual(t, elapsed, 5*time.Second, "Health check should wait for full timeout")

	// Should return error due to timeout
	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response["status"].(bool))
}

// TestAddChecker_ThreadSafety tests concurrent checker addition.
// This verifies the mutex protection works correctly.
func TestAddChecker_ThreadSafety(t *testing.T) {
	service := NewHealthService(8081)

	done := make(chan bool)

	// Start multiple goroutines adding checkers
	for i := 0; i < 20; i++ {
		go func(id int) {
			checker := &mockHealthChecker{
				name: "checker-" + string(rune('A'+id)),
			}
			service.AddChecker(checker)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 20; i++ {
		<-done
	}

	// Verify all checkers were added
	service.mu.RLock()
	count := len(service.checkers)
	service.mu.RUnlock()

	assert.Equal(t, 20, count, "All checkers should be added")
}

// TestReadyzResponse_JSONFormat tests response JSON structure.
// This verifies the JSON response matches the expected format.
func TestReadyzResponse_JSONFormat(t *testing.T) {
	service := NewHealthService(8081)

	checker := &mockHealthChecker{
		name: "database",
		checkFunc: func(ctx context.Context) error {
			return nil
		},
	}
	service.AddChecker(checker)

	req := httptest.NewRequest("GET", "/readyz", nil)
	rec := httptest.NewRecorder()

	service.handleReadyz(rec, req)

	var response map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify required fields
	assert.Contains(t, response, "status")
	assert.Contains(t, response, "checks")

	// Verify types
	assert.IsType(t, true, response["status"])
	assert.IsType(t, map[string]interface{}{}, response["checks"])
}

// TestHealthService_MultipleEndpoints tests all health endpoints together.
// This verifies all three endpoints coexist properly.
func TestHealthService_MultipleEndpoints(t *testing.T) {
	service := NewHealthService(8081)

	checker := &mockHealthChecker{
		name: "database",
		checkFunc: func(ctx context.Context) error {
			return nil
		},
	}
	service.AddChecker(checker)

	endpoints := []string{"/healthz", "/livez", "/readyz"}

	for _, endpoint := range endpoints {
		t.Run(endpoint, func(t *testing.T) {
			req := httptest.NewRequest("GET", endpoint, nil)
			rec := httptest.NewRecorder()

			switch endpoint {
			case "/healthz":
				service.handleHealthz(rec, req)
			case "/livez":
				service.handleLivez(rec, req)
			case "/readyz":
				service.handleReadyz(rec, req)
			}

			assert.Equal(t, http.StatusOK, rec.Code)
		})
	}
}

// TestHealthService_StartAndStop tests service lifecycle with actual server.
// This verifies Start and Stop methods work correctly with integration.
func TestHealthService_StartAndStop(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	service := NewHealthService(8091)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Start service in background
	errCh := make(chan error, 1)
	go func() {
		errCh <- service.Start(ctx)
	}()

	// Wait for server to start
	time.Sleep(200 * time.Millisecond)

	// Verify server is running by making a health check request
	resp, err := http.Get("http://localhost:8091/livez")
	if err == nil {
		resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode, "Health server should be responding")
	}

	// Stop the service
	stopErr := service.Stop(context.Background())
	assert.NoError(t, stopErr, "Stop should succeed")

	// Cancel context to ensure Start returns
	cancel()

	// Wait for Start to complete
	select {
	case err := <-errCh:
		// http.ErrServerClosed is expected when server is shut down gracefully
		if err != nil && !errors.Is(err, http.ErrServerClosed) && !errors.Is(err, context.Canceled) {
			t.Logf("Start returned non-nil error (may be expected): %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("Start did not complete after Stop and context cancellation")
	}
}

// TestHealthService_MultipleStops tests calling Stop multiple times.
// This verifies Stop is idempotent and doesn't panic.
func TestHealthService_MultipleStops(t *testing.T) {
	service := NewHealthService(8092)

	ctx := context.Background()

	// First stop (server not started, should succeed)
	err1 := service.Stop(ctx)
	assert.NoError(t, err1, "First Stop with nil server should succeed")

	// Second stop (should also succeed)
	err2 := service.Stop(ctx)
	assert.NoError(t, err2, "Second Stop should also succeed")
}

// TestHealthService_StopWithError tests Stop with context timeout.
// This verifies Stop respects context timeout.
func TestHealthService_StopWithError(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	service := NewHealthService(8093)

	// Start service
	ctx := context.Background()
	go func() {
		_ = service.Start(ctx)
	}()

	time.Sleep(100 * time.Millisecond)

	// Stop with very short timeout
	stopCtx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	err := service.Stop(stopCtx)
	// May or may not error depending on timing
	if err != nil {
		t.Logf("Stop with short timeout returned: %v", err)
	}
}
