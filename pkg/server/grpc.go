// Package server provides HTTP and gRPC server implementations for EggyByte services.

import (
	"context"
	"fmt"
	"net"
	"sync"

	"github.com/eggybyte-technology/go-eggybyte-core/pkg/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// GRPCServer represents a business gRPC server for serving RPC APIs.
// This server is separate from HTTP servers to provide clear separation
// of concerns and protocol-specific optimizations.
//
// Thread Safety: GRPCServer is safe for concurrent use after initialization.
// The server should be started once and not modified after Start() is called.
//
// Usage:
//
//	server := NewGRPCServer(":9090")
//	pb.RegisterUserServiceServer(server.GetServer(), userService)
//	go server.Start(ctx)
type GRPCServer struct {
	// server is the underlying gRPC server instance
	server *grpc.Server

	// port is the listening port for this server
	port string

	// listener is the network listener for this server
	listener net.Listener

	// logger is the structured logger for this server
	logger log.Logger

	// enableReflection enables gRPC reflection for development and debugging
	enableReflection bool

	// mu protects concurrent access to enableReflection and listener fields
	mu sync.RWMutex
}

// NewGRPCServer creates a new business gRPC server with the specified port.
// The port should be in the format ":9090" or "0.0.0.0:9090".
//
// Parameters:
//   - port: The listening address and port for the gRPC server
//
// Returns:
//   - *GRPCServer: A new gRPC server instance ready for service registration
//
// Example:
//
//	server := NewGRPCServer(":9090")
//	pb.RegisterUserServiceServer(server.GetServer(), userService)
func NewGRPCServer(port string) *GRPCServer {
	// Create gRPC server with default options
	grpcServer := grpc.NewServer()

	return &GRPCServer{
		server:           grpcServer,
		port:             port,
		logger:           log.Default(),
		enableReflection: false, // Disabled by default for security
	}
}

// NewGRPCServerWithOptions creates a new business gRPC server with custom options.
// This constructor allows fine-grained control over gRPC server configuration.
//
// Parameters:
//   - port: The listening address and port for the gRPC server
//   - options: gRPC server options for custom configuration
//
// Returns:
//   - *GRPCServer: A new gRPC server instance with custom options
//
// Example:
//
//	server := NewGRPCServerWithOptions(":9090",
//	    grpc.ConnectionTimeout(60*time.Second),
//	    grpc.MaxRecvMsgSize(1024*1024),
//	)
func NewGRPCServerWithOptions(port string, options ...grpc.ServerOption) *GRPCServer {
	grpcServer := grpc.NewServer(options...)

	return &GRPCServer{
		server:           grpcServer,
		port:             port,
		logger:           log.Default(),
		enableReflection: false,
	}
}

// Start begins serving gRPC requests on the configured port.
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
//   - Creates network listener on the configured port
//   - Enables gRPC reflection if configured
//   - Logs server startup information
//   - Handles graceful shutdown on context cancellation
//   - Returns immediately if server is already running
//
// Example:
//
//	ctx, cancel := context.WithCancel(context.Background())
//	go func() {
//	    if err := server.Start(ctx); err != nil {
//	        log.Error("gRPC server failed", log.Field{Key: "error", Value: err})
//	    }
//	}()
//	defer cancel()
func (s *GRPCServer) Start(ctx context.Context) error {
	// Create network listener
	listener, err := net.Listen("tcp", s.port)
	if err != nil {
		return fmt.Errorf("failed to create gRPC listener on %s: %w", s.port, err)
	}

	s.mu.Lock()
	s.listener = listener
	s.mu.Unlock()

	// Enable reflection if configured
	s.mu.RLock()
	enableReflection := s.enableReflection
	s.mu.RUnlock()

	if enableReflection {
		reflection.Register(s.server)
		s.logger.Info("gRPC reflection enabled")
	}

	s.logger.Info("Starting business gRPC server",
		log.Field{Key: "port", Value: s.port},
		log.Field{Key: "address", Value: listener.Addr().String()})

	// Create a channel to receive server errors
	errChan := make(chan error, 1)

	// Start server in a goroutine
	go func() {
		if err := s.server.Serve(listener); err != nil {
			errChan <- fmt.Errorf("gRPC server failed: %w", err)
		}
	}()

	// Wait for context cancellation or server error
	select {
	case <-ctx.Done():
		s.logger.Info("Shutting down gRPC server",
			log.Field{Key: "reason", Value: "context_canceled"})

		// Attempt graceful shutdown
		s.server.GracefulStop()
		s.logger.Info("gRPC server shutdown completed")
		return nil

	case err := <-errChan:
		return err
	}
}

// Stop gracefully shuts down the gRPC server.
// This method is provided for compatibility with the Service interface.
// In practice, shutdown is handled by the Start method when context is canceled.
//
// Parameters:
//   - ctx: Context for timeout control during shutdown
//
// Returns:
//   - error: Returns error if shutdown fails
func (s *GRPCServer) Stop(ctx context.Context) error {
	s.logger.Info("Stopping gRPC server")
	s.server.GracefulStop()
	return nil
}

// GetPort returns the configured port for this server.
// This method is useful for logging and monitoring purposes.
//
// Returns:
//   - string: The port string (e.g., ":9090")
func (s *GRPCServer) GetPort() string {
	return s.port
}

// GetServer returns the underlying grpc.Server instance.
// This method is provided for service registration and advanced configuration.
//
// Returns:
//   - *grpc.Server: The underlying gRPC server instance
//
// Example:
//
//	pb.RegisterUserServiceServer(server.GetServer(), userService)
//
// Note: Modifying the returned server after Start() has been called
// may cause undefined behavior.
func (s *GRPCServer) GetServer() *grpc.Server {
	return s.server
}

// SetLogger sets the logger for this gRPC server.
// This method allows customization of logging behavior.
//
// Parameters:
//   - logger: The logger instance to use for this server
func (s *GRPCServer) SetLogger(logger interface{}) {
	if l, ok := logger.(log.Logger); ok {
		s.logger = l
	}
}

// EnableReflection enables gRPC reflection for this server.
// Reflection allows tools like grpcurl to discover and call services.
// This should only be enabled in development environments for security reasons.
//
// Example:
//
//	server.EnableReflection()
func (s *GRPCServer) EnableReflection() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.enableReflection = true
	s.logger.Info("gRPC reflection will be enabled on next start")
}

// DisableReflection disables gRPC reflection for this server.
// This is the default state and recommended for production environments.
//
// Example:
//
//	server.DisableReflection()
func (s *GRPCServer) DisableReflection() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.enableReflection = false
	s.logger.Info("gRPC reflection disabled")
}

// IsReflectionEnabled returns whether gRPC reflection is enabled.
// This method is useful for logging and monitoring purposes.
//
// Returns:
//   - bool: True if reflection is enabled, false otherwise
func (s *GRPCServer) IsReflectionEnabled() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.enableReflection
}
