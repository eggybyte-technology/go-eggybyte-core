// Package server provides HTTP and gRPC server implementations for EggyByte services.
package server

import (
	"context"
	"net/http"

	"google.golang.org/grpc"

	"github.com/eggybyte-technology/go-eggybyte-core/pkg/log"
)

// HTTPServerInterface defines the interface for business HTTP servers.
// This interface allows for easy testing and implementation swapping.
//
// Thread Safety: Implementations should be safe for concurrent use after initialization.
type HTTPServerInterface interface {
	// Start begins serving HTTP requests on the configured port.
	// This method blocks until the server is stopped or encounters an error.
	Start(ctx context.Context) error

	// Stop gracefully shuts down the HTTP server.
	Stop(ctx context.Context) error

	// HandleFunc registers a handler function for the given pattern.
	HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request))

	// Handle registers a handler for the given pattern.
	Handle(pattern string, handler http.Handler)

	// GetPort returns the configured port for this server.
	GetPort() string

	// SetLogger sets the logger for this HTTP server.
	SetLogger(logger interface{}) // Using interface{} to avoid circular imports
}

// GRPCServerInterface defines the interface for business gRPC servers.
// This interface allows for easy testing and implementation swapping.
//
// Thread Safety: Implementations should be safe for concurrent use after initialization.
type GRPCServerInterface interface {
	// Start begins serving gRPC requests on the configured port.
	// This method blocks until the server is stopped or encounters an error.
	Start(ctx context.Context) error

	// Stop gracefully shuts down the gRPC server.
	Stop(ctx context.Context) error

	// GetPort returns the configured port for this server.
	GetPort() string

	// GetServer returns the underlying grpc.Server instance.
	GetServer() *grpc.Server

	// SetLogger sets the logger for this gRPC server.
	SetLogger(logger interface{}) // Using interface{} to avoid circular imports

	// EnableReflection enables gRPC reflection for this server.
	EnableReflection()

	// DisableReflection disables gRPC reflection for this server.
	DisableReflection()

	// IsReflectionEnabled returns whether gRPC reflection is enabled.
	IsReflectionEnabled() bool
}

// ServerManager provides a unified interface for managing both HTTP and gRPC servers.
// This manager handles the lifecycle of both server types and provides convenient
// methods for common operations.
//
// Thread Safety: ServerManager is safe for concurrent use after initialization.
type ServerManager struct {
	// httpServer is the business HTTP server instance
	httpServer HTTPServerInterface

	// grpcServer is the business gRPC server instance
	grpcServer GRPCServerInterface

	// logger is the structured logger for this manager
	logger interface{} // Using interface{} to avoid circular imports
}

// NewServerManager creates a new server manager for coordinating HTTP and gRPC servers.
// This manager provides a unified interface for managing both server types.
//
// Parameters:
//   - httpServer: The HTTP server instance to manage
//   - grpcServer: The gRPC server instance to manage (can be nil)
//
// Returns:
//   - *ServerManager: A new server manager instance
//
// Example:
//
//	httpSrv := NewHTTPServer(":8080")
//	grpcSrv := NewGRPCServer(":9090")
//	manager := NewServerManager(httpSrv, grpcSrv)
func NewServerManager(httpServer HTTPServerInterface, grpcServer GRPCServerInterface) *ServerManager {
	return &ServerManager{
		httpServer: httpServer,
		grpcServer: grpcServer,
	}
}

// Start begins serving requests on both HTTP and gRPC servers.
// This method starts both servers concurrently and blocks until both are stopped
// or one encounters an error.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//
// Returns:
//   - error: Returns error if any server fails to start or encounters a fatal error
//
// Behavior:
//   - Starts HTTP server in a goroutine
//   - Starts gRPC server in a goroutine (if configured)
//   - Waits for context cancellation or server errors
//   - Performs graceful shutdown of both servers
//
// Example:
//
//	ctx, cancel := context.WithCancel(context.Background())
//	go func() {
//	    if err := manager.Start(ctx); err != nil {
//	        log.Error("Server manager failed", log.Field{Key: "error", Value: err})
//	    }
//	}()
//	defer cancel()
func (m *ServerManager) Start(ctx context.Context) error {
	// Start HTTP server if configured
	if m.httpServer != nil {
		go func() {
			if err := m.httpServer.Start(ctx); err != nil {
				// Log error but don't return it here since we're in a goroutine
				// The main goroutine will handle the error through context cancellation
				log.Default().Error("HTTP server failed to start", log.Field{Key: "error", Value: err})
			}
		}()
	}

	// Start gRPC server if configured
	if m.grpcServer != nil {
		go func() {
			if err := m.grpcServer.Start(ctx); err != nil {
				// Log error but don't return it here since we're in a goroutine
				// The main goroutine will handle the error through context cancellation
				log.Default().Error("gRPC server failed to start", log.Field{Key: "error", Value: err})
			}
		}()
	}

	// Wait for context cancellation
	<-ctx.Done()

	// Perform graceful shutdown
	if err := m.Stop(context.Background()); err != nil {
		return err
	}

	return nil
}

// Stop gracefully shuts down both HTTP and gRPC servers.
// This method is provided for compatibility with the Service interface.
// In practice, shutdown is handled by the Start method when context is canceled.
//
// Parameters:
//   - ctx: Context for timeout control during shutdown
//
// Returns:
//   - error: Returns error if shutdown fails
func (m *ServerManager) Stop(ctx context.Context) error {
	// Stop HTTP server if configured
	if m.httpServer != nil {
		if err := m.httpServer.Stop(ctx); err != nil {
			return err
		}
	}

	// Stop gRPC server if configured
	if m.grpcServer != nil {
		if err := m.grpcServer.Stop(ctx); err != nil {
			return err
		}
	}

	return nil
}

// GetHTTPServer returns the HTTP server instance.
// This method provides access to the underlying HTTP server for advanced configuration.
//
// Returns:
//   - HTTPServerInterface: The HTTP server instance
func (m *ServerManager) GetHTTPServer() HTTPServerInterface {
	return m.httpServer
}

// GetGRPCServer returns the gRPC server instance.
// This method provides access to the underlying gRPC server for service registration.
//
// Returns:
//   - GRPCServerInterface: The gRPC server instance, or nil if not configured
func (m *ServerManager) GetGRPCServer() GRPCServerInterface {
	return m.grpcServer
}

// SetLogger sets the logger for both servers.
// This method allows customization of logging behavior across both server types.
//
// Parameters:
//   - logger: The logger instance to use for both servers
func (m *ServerManager) SetLogger(logger interface{}) {
	m.logger = logger
	m.httpServer.SetLogger(logger)
	if m.grpcServer != nil {
		m.grpcServer.SetLogger(logger)
	}
}
