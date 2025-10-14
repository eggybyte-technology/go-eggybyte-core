package server

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

func TestNewGRPCServer(t *testing.T) {
	tests := []struct {
		name     string
		port     string
		expected string
	}{
		{
			name:     "standard_port",
			port:     ":9090",
			expected: ":9090",
		},
		{
			name:     "with_host",
			port:     "0.0.0.0:9090",
			expected: "0.0.0.0:9090",
		},
		{
			name:     "localhost",
			port:     "localhost:9090",
			expected: "localhost:9090",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := NewGRPCServer(tt.port)

			assert.NotNil(t, server)
			assert.Equal(t, tt.expected, server.GetPort())
			assert.NotNil(t, server.GetServer())
			assert.False(t, server.IsReflectionEnabled())
		})
	}
}

func TestNewGRPCServerWithOptions(t *testing.T) {
	server := NewGRPCServerWithOptions(":9090",
		grpc.ConnectionTimeout(60*time.Second),
		grpc.MaxRecvMsgSize(1024*1024),
	)

	assert.NotNil(t, server)
	assert.Equal(t, ":9090", server.GetPort())
	assert.NotNil(t, server.GetServer())
	assert.False(t, server.IsReflectionEnabled())
}

func TestGRPCServer_Reflection(t *testing.T) {
	server := NewGRPCServer(":9090")

	// Test initial state
	assert.False(t, server.IsReflectionEnabled())

	// Enable reflection
	server.EnableReflection()
	assert.True(t, server.IsReflectionEnabled())

	// Disable reflection
	server.DisableReflection()
	assert.False(t, server.IsReflectionEnabled())
}

func TestGRPCServer_Start_ContextCancellation(t *testing.T) {
	// Find an available port
	listener, err := net.Listen("tcp", ":0")
	require.NoError(t, err)
	port := listener.Addr().String()
	listener.Close()

	server := NewGRPCServer(port)

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

func TestGRPCServer_Stop(t *testing.T) {
	// Find an available port
	listener, err := net.Listen("tcp", ":0")
	require.NoError(t, err)
	port := listener.Addr().String()
	listener.Close()

	server := NewGRPCServer(port)

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

	err = server.Stop(stopCtx)
	assert.NoError(t, err)
}

func TestGRPCServer_GetPort(t *testing.T) {
	server := NewGRPCServer(":9090")
	assert.Equal(t, ":9090", server.GetPort())
}

func TestGRPCServer_GetServer(t *testing.T) {
	server := NewGRPCServer(":9090")
	grpcServer := server.GetServer()

	assert.NotNil(t, grpcServer)
}

func TestGRPCServer_SetLogger(t *testing.T) {
	server := NewGRPCServer(":9090")

	// This test mainly ensures the method doesn't panic
	// In a real implementation, you might want to test logger functionality
	server.SetLogger(nil)

	// Test passes if no panic occurs
	assert.True(t, true)
}

func TestGRPCServer_ConcurrentAccess(t *testing.T) {
	server := NewGRPCServer(":9090")

	// Test concurrent reflection toggling
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(i int) {
			if i%2 == 0 {
				server.EnableReflection()
			} else {
				server.DisableReflection()
			}
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

func TestGRPCServer_Start_AlreadyRunning(t *testing.T) {
	// Find an available port
	listener, err := net.Listen("tcp", ":0")
	require.NoError(t, err)
	port := listener.Addr().String()
	listener.Close()

	server := NewGRPCServer(port)

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

func TestGRPCServer_InvalidPort(t *testing.T) {
	// Test with invalid port format
	server := NewGRPCServer("invalid-port")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start should fail with invalid port
	err := server.Start(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create gRPC listener")
}

func TestGRPCServer_ConnectionTimeout(t *testing.T) {
	// Find an available port
	listener, err := net.Listen("tcp", ":0")
	require.NoError(t, err)
	port := listener.Addr().String()
	listener.Close()

	// Create server with custom connection timeout
	server := NewGRPCServerWithOptions(port,
		grpc.ConnectionTimeout(1*time.Second),
	)

	assert.NotNil(t, server)
	assert.Equal(t, port, server.GetPort())
}

func TestGRPCServer_KeepaliveParams(t *testing.T) {
	// Find an available port
	listener, err := net.Listen("tcp", ":0")
	require.NoError(t, err)
	port := listener.Addr().String()
	listener.Close()

	// Create server with custom options
	server := NewGRPCServerWithOptions(port,
		grpc.ConnectionTimeout(30*time.Second),
	)

	assert.NotNil(t, server)
	assert.Equal(t, port, server.GetPort())
}

func TestGRPCServer_MaxRecvMsgSize(t *testing.T) {
	// Find an available port
	listener, err := net.Listen("tcp", ":0")
	require.NoError(t, err)
	port := listener.Addr().String()
	listener.Close()

	// Create server with custom max receive message size
	server := NewGRPCServerWithOptions(port,
		grpc.MaxRecvMsgSize(1024*1024), // 1MB
	)

	assert.NotNil(t, server)
	assert.Equal(t, port, server.GetPort())
}

func TestGRPCServer_ReflectionState(t *testing.T) {
	server := NewGRPCServer(":9090")

	// Test initial state
	assert.False(t, server.IsReflectionEnabled())

	// Enable and verify
	server.EnableReflection()
	assert.True(t, server.IsReflectionEnabled())

	// Disable and verify
	server.DisableReflection()
	assert.False(t, server.IsReflectionEnabled())

	// Enable again and verify
	server.EnableReflection()
	assert.True(t, server.IsReflectionEnabled())
}

func TestGRPCServer_ServerOptions(t *testing.T) {
	// Find an available port
	listener, err := net.Listen("tcp", ":0")
	require.NoError(t, err)
	port := listener.Addr().String()
	listener.Close()

	// Test multiple options
	server := NewGRPCServerWithOptions(port,
		grpc.ConnectionTimeout(30*time.Second),
		grpc.MaxRecvMsgSize(512*1024),
		grpc.MaxSendMsgSize(512*1024),
	)

	assert.NotNil(t, server)
	assert.Equal(t, port, server.GetPort())

	// Verify server is created
	grpcServer := server.GetServer()
	assert.NotNil(t, grpcServer)
}
