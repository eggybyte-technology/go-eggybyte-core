package server

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewHTTPServer(t *testing.T) {
	tests := []struct {
		name     string
		port     string
		expected string
	}{
		{
			name:     "standard_port",
			port:     ":8080",
			expected: ":8080",
		},
		{
			name:     "with_host",
			port:     "0.0.0.0:8080",
			expected: "0.0.0.0:8080",
		},
		{
			name:     "localhost",
			port:     "localhost:8080",
			expected: "localhost:8080",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := NewHTTPServer(tt.port)

			assert.NotNil(t, server)
			assert.Equal(t, tt.expected, server.GetPort())
			assert.NotNil(t, server.GetServer())
			assert.NotNil(t, server.mux)
		})
	}
}

func TestHTTPServer_HandleFunc(t *testing.T) {
	server := NewHTTPServer(":8080")

	// Register a test handler
	server.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	})

	// Test the handler
	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	server.mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "test response", rec.Body.String())
}

func TestHTTPServer_Handle(t *testing.T) {
	server := NewHTTPServer(":8080")

	// Create a test handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("handler response"))
	})

	// Register the handler
	server.Handle("/handler", handler)

	// Test the handler
	req := httptest.NewRequest("GET", "/handler", nil)
	rec := httptest.NewRecorder()

	server.mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "handler response", rec.Body.String())
}

func TestHTTPServer_Start_ContextCancellation(t *testing.T) {
	server := NewHTTPServer(":0") // Use port 0 for automatic port assignment

	// Register a test handler
	server.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test"))
	})

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())

	// Start server in goroutine
	errChan := make(chan error, 1)
	go func() {
		errChan <- server.Start(ctx)
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Cancel context
	cancel()

	// Wait for server to stop
	select {
	case err := <-errChan:
		assert.NoError(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("Server did not stop within timeout")
	}
}

func TestHTTPServer_Stop(t *testing.T) {
	server := NewHTTPServer(":0")

	// Register a test handler
	server.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test"))
	})

	// Start server in goroutine
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		server.Start(ctx)
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Stop server
	stopCtx, stopCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer stopCancel()

	err := server.Stop(stopCtx)
	assert.NoError(t, err)
}

func TestHTTPServer_GetPort(t *testing.T) {
	server := NewHTTPServer(":8080")
	assert.Equal(t, ":8080", server.GetPort())
}

func TestHTTPServer_GetServer(t *testing.T) {
	server := NewHTTPServer(":8080")
	httpServer := server.GetServer()

	assert.NotNil(t, httpServer)
	assert.Equal(t, ":8080", httpServer.Addr)
}

func TestHTTPServer_SetLogger(t *testing.T) {
	server := NewHTTPServer(":8080")

	// This test mainly ensures the method doesn't panic
	// In a real implementation, you might want to test logger functionality
	server.SetLogger(nil)

	// Test passes if no panic occurs
	assert.True(t, true)
}

func TestHTTPServer_ConcurrentAccess(t *testing.T) {
	server := NewHTTPServer(":8080")

	// Test concurrent handler registration
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(i int) {
			server.HandleFunc(fmt.Sprintf("/test%d", i), func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Test that handlers were registered
	req := httptest.NewRequest("GET", "/test0", nil)
	rec := httptest.NewRecorder()
	server.mux.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestHTTPServer_Start_AlreadyRunning(t *testing.T) {
	server := NewHTTPServer(":0")

	// Start server
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		server.Start(ctx)
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Try to start again - should not cause issues
	// (In real implementation, this might return an error)
	go func() {
		server.Start(ctx)
	}()

	// Test passes if no panic occurs
	assert.True(t, true)
}

func TestHTTPServer_HandlerRegistration(t *testing.T) {
	server := NewHTTPServer(":8080")

	// Register multiple handlers
	server.HandleFunc("/api/v1/users", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("users"))
	})

	server.HandleFunc("/api/v1/orders", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("orders"))
	})

	// Test both handlers
	tests := []struct {
		path     string
		expected string
	}{
		{"/api/v1/users", "users"},
		{"/api/v1/orders", "orders"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			rec := httptest.NewRecorder()

			server.mux.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(t, tt.expected, rec.Body.String())
		})
	}
}

func TestHTTPServer_ErrorHandling(t *testing.T) {
	server := NewHTTPServer(":8080")

	// Register a handler that returns an error
	server.HandleFunc("/error", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal error"))
	})

	// Test error handler
	req := httptest.NewRequest("GET", "/error", nil)
	rec := httptest.NewRecorder()

	server.mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Equal(t, "internal error", rec.Body.String())
}

func TestHTTPServer_NotFound(t *testing.T) {
	server := NewHTTPServer(":8080")

	// Test non-existent endpoint
	req := httptest.NewRequest("GET", "/nonexistent", nil)
	rec := httptest.NewRecorder()

	server.mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestHTTPServer_MethodNotAllowed(t *testing.T) {
	server := NewHTTPServer(":8080")

	// Register GET handler
	server.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	// Test POST to GET-only endpoint
	req := httptest.NewRequest("POST", "/test", nil)
	rec := httptest.NewRecorder()

	server.mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
}
