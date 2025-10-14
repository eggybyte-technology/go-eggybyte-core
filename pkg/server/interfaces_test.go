package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewServerManager(t *testing.T) {
	httpServer := NewHTTPServer(":8080")
	grpcServer := NewGRPCServer(":9090")

	manager := NewServerManager(httpServer, grpcServer)

	assert.NotNil(t, manager)
	assert.Equal(t, httpServer, manager.GetHTTPServer())
	assert.Equal(t, grpcServer, manager.GetGRPCServer())
}

func TestNewServerManager_HTTPOnly(t *testing.T) {
	httpServer := NewHTTPServer(":8080")

	manager := NewServerManager(httpServer, nil)

	assert.NotNil(t, manager)
	assert.Equal(t, httpServer, manager.GetHTTPServer())
	assert.Nil(t, manager.GetGRPCServer())
}

func TestServerManager_Start_ContextCancellation(t *testing.T) {
	httpServer := NewHTTPServer(":0")
	grpcServer := NewGRPCServer(":0")

	manager := NewServerManager(httpServer, grpcServer)

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())

	// Start manager in goroutine
	errChan := make(chan error, 1)
	go func() {
		errChan <- manager.Start(ctx)
	}()

	// Give servers time to start
	time.Sleep(100 * time.Millisecond)

	// Cancel context
	cancel()

	// Wait for manager to stop
	select {
	case err := <-errChan:
		assert.NoError(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("Manager did not stop within timeout")
	}
}

func TestServerManager_Stop(t *testing.T) {
	httpServer := NewHTTPServer(":0")
	grpcServer := NewGRPCServer(":0")

	manager := NewServerManager(httpServer, grpcServer)

	// Start manager in goroutine
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		manager.Start(ctx)
	}()

	// Give servers time to start
	time.Sleep(100 * time.Millisecond)

	// Stop manager
	stopCtx, stopCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer stopCancel()

	err := manager.Stop(stopCtx)
	assert.NoError(t, err)
}

func TestServerManager_HTTPOnly_Start(t *testing.T) {
	httpServer := NewHTTPServer(":0")

	manager := NewServerManager(httpServer, nil)

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start manager in goroutine
	errChan := make(chan error, 1)
	go func() {
		errChan <- manager.Start(ctx)
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Cancel context
	cancel()

	// Wait for manager to stop
	select {
	case err := <-errChan:
		assert.NoError(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("Manager did not stop within timeout")
	}
}

func TestServerManager_SetLogger(t *testing.T) {
	httpServer := NewHTTPServer(":8080")
	grpcServer := NewGRPCServer(":9090")

	manager := NewServerManager(httpServer, grpcServer)

	// This test mainly ensures the method doesn't panic
	// In a real implementation, you might want to test logger functionality
	manager.SetLogger(nil)

	// Test passes if no panic occurs
	assert.True(t, true)
}

func TestServerManager_ConcurrentAccess(t *testing.T) {
	httpServer := NewHTTPServer(":8080")
	grpcServer := NewGRPCServer(":9090")

	manager := NewServerManager(httpServer, grpcServer)

	// Test concurrent access to getters
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func() {
			// Access getters concurrently
			_ = manager.GetHTTPServer()
			_ = manager.GetGRPCServer()
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Test passes if no panic occurs
	assert.True(t, true)
}

func TestServerManager_HTTPHandlerRegistration(t *testing.T) {
	httpServer := NewHTTPServer(":8080")
	grpcServer := NewGRPCServer(":9090")

	manager := NewServerManager(httpServer, grpcServer)
	assert.NotNil(t, manager)

	// Test that we can access the HTTP server through the manager
	retrievedHTTPServer := manager.GetHTTPServer()
	assert.NotNil(t, retrievedHTTPServer)
	assert.Equal(t, httpServer, retrievedHTTPServer)

	// Test that we can register handlers
	httpServer.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test"))
	})

	// Test the handler by getting the server and testing it
	server := httpServer.GetServer()
	assert.NotNil(t, server)
	assert.NotNil(t, server.Handler)

	// Test the handler using the server's handler
	req, _ := http.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	// Use the server's handler directly
	server.Handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "test", rec.Body.String())
}

func TestServerManager_GRPCServerAccess(t *testing.T) {
	httpServer := NewHTTPServer(":8080")
	grpcServer := NewGRPCServer(":9090")

	manager := NewServerManager(httpServer, grpcServer)

	// Access gRPC server through manager
	retrievedGRPCServer := manager.GetGRPCServer()

	assert.Equal(t, grpcServer, retrievedGRPCServer)
	assert.NotNil(t, retrievedGRPCServer.GetServer())
}

func TestServerManager_HTTPOnly_GRPCAccess(t *testing.T) {
	httpServer := NewHTTPServer(":8080")

	manager := NewServerManager(httpServer, nil)

	// Access gRPC server through manager (should be nil)
	retrievedGRPCServer := manager.GetGRPCServer()

	assert.Nil(t, retrievedGRPCServer)
}

func TestServerManager_Start_HTTPOnly(t *testing.T) {
	httpServer := NewHTTPServer(":0")

	manager := NewServerManager(httpServer, nil)

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start manager in goroutine
	errChan := make(chan error, 1)
	go func() {
		errChan <- manager.Start(ctx)
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Cancel context
	cancel()

	// Wait for manager to stop
	select {
	case err := <-errChan:
		assert.NoError(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("Manager did not stop within timeout")
	}
}

func TestServerManager_Stop_HTTPOnly(t *testing.T) {
	httpServer := NewHTTPServer(":0")

	manager := NewServerManager(httpServer, nil)

	// Start manager in goroutine
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		manager.Start(ctx)
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Stop manager
	stopCtx, stopCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer stopCancel()

	err := manager.Stop(stopCtx)
	assert.NoError(t, err)
}

func TestServerManager_Start_GRPCOnly(t *testing.T) {
	grpcServer := NewGRPCServer(":0")

	manager := NewServerManager(nil, grpcServer)

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start manager in goroutine
	errChan := make(chan error, 1)
	go func() {
		errChan <- manager.Start(ctx)
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Cancel context
	cancel()

	// Wait for manager to stop
	select {
	case err := <-errChan:
		assert.NoError(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("Manager did not stop within timeout")
	}
}

func TestServerManager_Stop_GRPCOnly(t *testing.T) {
	grpcServer := NewGRPCServer(":0")

	manager := NewServerManager(nil, grpcServer)

	// Start manager in goroutine
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		manager.Start(ctx)
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Stop manager
	stopCtx, stopCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer stopCancel()

	err := manager.Stop(stopCtx)
	assert.NoError(t, err)
}

func TestServerManager_NoServers(t *testing.T) {
	manager := NewServerManager(nil, nil)

	assert.NotNil(t, manager)
	assert.Nil(t, manager.GetHTTPServer())
	assert.Nil(t, manager.GetGRPCServer())

	// Start should not cause issues
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errChan := make(chan error, 1)
	go func() {
		errChan <- manager.Start(ctx)
	}()

	// Give time for start
	time.Sleep(100 * time.Millisecond)

	// Cancel context
	cancel()

	// Wait for manager to stop
	select {
	case err := <-errChan:
		assert.NoError(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("Manager did not stop within timeout")
	}
}
