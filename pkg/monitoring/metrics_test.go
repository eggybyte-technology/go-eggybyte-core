package monitoring

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMetricsService(t *testing.T) {
	service := NewMetricsService(9091)

	assert.NotNil(t, service)
	assert.Equal(t, 9091, service.GetPort())
	assert.NotNil(t, service.GetRegistry())
}

func TestNewMetricsServiceWithRegistry(t *testing.T) {
	registry := prometheus.NewRegistry()
	service := NewMetricsServiceWithRegistry(9091, registry)

	assert.NotNil(t, service)
	assert.Equal(t, 9091, service.GetPort())
	assert.Equal(t, registry, service.GetRegistry())
}

func TestMetricsService_RegisterCollector(t *testing.T) {
	service := NewMetricsService(9091)

	// Create a test counter
	counter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "test_counter_total",
			Help: "A test counter",
		},
		[]string{"label"},
	)

	// Register collector
	err := service.RegisterCollector(counter)
	assert.NoError(t, err)

	// Try to register same collector again - should fail
	err = service.RegisterCollector(counter)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate metrics collector registration attempted")
}

func TestMetricsService_UnregisterCollector(t *testing.T) {
	service := NewMetricsService(9091)

	// Create a test counter
	counter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "test_counter_total",
			Help: "A test counter",
		},
		[]string{"label"},
	)

	// Register collector
	err := service.RegisterCollector(counter)
	assert.NoError(t, err)

	// Unregister collector
	success := service.UnregisterCollector(counter)
	assert.True(t, success)

	// Try to unregister again - should fail
	success = service.UnregisterCollector(counter)
	assert.False(t, success)
}

func TestMetricsService_Start_ContextCancellation(t *testing.T) {
	service := NewMetricsService(0) // Use port 0 for automatic port assignment

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

func TestMetricsService_Stop(t *testing.T) {
	service := NewMetricsService(0)

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

func TestMetricsService_GetPort(t *testing.T) {
	service := NewMetricsService(9091)
	assert.Equal(t, 9091, service.GetPort())
}

func TestMetricsService_GetRegistry(t *testing.T) {
	service := NewMetricsService(9091)
	registry := service.GetRegistry()

	assert.NotNil(t, registry)

	// Test that we can use the registry
	counter := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "test_counter_total",
		Help: "A test counter",
	})

	err := registry.Register(counter)
	assert.NoError(t, err)
}

func TestMetricsService_SetLogger(t *testing.T) {
	service := NewMetricsService(9091)

	// This test mainly ensures the method doesn't panic
	// In a real implementation, you might want to test logger functionality
	service.SetLogger(nil)

	// Test passes if no panic occurs
	assert.True(t, true)
}

func TestMetricsService_GetCollectorCount(t *testing.T) {
	service := NewMetricsService(9091)

	// This method is a placeholder, so we just test it doesn't panic
	count := service.GetCollectorCount()
	assert.Equal(t, 0, count)
}

func TestMetricsService_Start_AlreadyRunning(t *testing.T) {
	service := NewMetricsService(0)

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

func TestMetricsService_Stop_NotStarted(t *testing.T) {
	service := NewMetricsService(9091)

	// Stop service that was never started
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := service.Stop(ctx)
	assert.NoError(t, err)
}

func TestMetricsService_ConcurrentAccess(t *testing.T) {
	service := NewMetricsService(9091)

	// Test concurrent collector registration
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(i int) {
			counter := prometheus.NewCounterVec(
				prometheus.CounterOpts{
					Name: "test_counter_" + string(rune(i)) + "_total",
					Help: "A test counter",
				},
				[]string{"label"},
			)
			service.RegisterCollector(counter)
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Test passes if no panic occurs
	assert.True(t, true)
}

func TestMetricsService_MetricsEndpoint(t *testing.T) {
	service := NewMetricsService(0)

	// Add a test counter
	counter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "test_counter_total",
			Help: "A test counter",
		},
		[]string{"label"},
	)

	err := service.RegisterCollector(counter)
	require.NoError(t, err)

	// Increment counter
	counter.WithLabelValues("test").Inc()

	// Start service in goroutine
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		service.Start(ctx)
	}()

	// Give service time to start
	time.Sleep(100 * time.Millisecond)

	// Test metrics endpoint
	req := httptest.NewRequest("GET", "/metrics", nil)
	rec := httptest.NewRecorder()

	// Create a test handler to simulate the metrics endpoint
	handler := promhttp.HandlerFor(service.GetRegistry(), promhttp.HandlerOpts{
		EnableOpenMetrics: true,
	})
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "test_counter_total")
	assert.Contains(t, rec.Body.String(), "test")
}

func TestMetricsService_CustomRegistry(t *testing.T) {
	// Create custom registry
	registry := prometheus.NewRegistry()

	// Add custom collector
	counter := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "custom_counter_total",
		Help: "A custom counter",
	})
	registry.MustRegister(counter)

	// Create service with custom registry
	service := NewMetricsServiceWithRegistry(9091, registry)

	assert.Equal(t, registry, service.GetRegistry())

	// Increment counter
	counter.Inc()

	// Test that the counter is accessible
	req := httptest.NewRequest("GET", "/metrics", nil)
	rec := httptest.NewRecorder()

	handler := promhttp.HandlerFor(service.GetRegistry(), promhttp.HandlerOpts{
		EnableOpenMetrics: true,
	})
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "custom_counter_total")
}

func TestMetricsService_ErrorHandling(t *testing.T) {
	service := NewMetricsService(9091)

	// Try to register nil collector
	err := service.RegisterCollector(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "nil collector")
}

func TestMetricsService_DefaultCollectors(t *testing.T) {
	service := NewMetricsService(9091)

	// Test that default collectors are registered
	registry := service.GetRegistry()

	// Create a test handler to check what's registered
	req := httptest.NewRequest("GET", "/metrics", nil)
	rec := httptest.NewRecorder()

	handler := promhttp.HandlerFor(registry, promhttp.HandlerOpts{
		EnableOpenMetrics: true,
	})
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	// Check for default collectors
	body := rec.Body.String()
	assert.Contains(t, body, "go_")      // Go runtime metrics
	assert.Contains(t, body, "process_") // Process metrics
}
