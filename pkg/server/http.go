// Package server provides HTTP and gRPC server implementations for EggyByte services.
// It includes business server implementations with handler registration,
// lifecycle management, and graceful shutdown capabilities.
package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/eggybyte-technology/go-eggybyte-core/pkg/log"
)

// HTTPServer represents a business HTTP server for serving REST APIs.
// This server is separate from health check and metrics servers to provide
// clear separation of concerns and security isolation.
//
// Thread Safety: HTTPServer is safe for concurrent use after initialization.
// The server should be started once and not modified after Start() is called.
//
// Usage:
//
//	server := NewHTTPServer(":8080")
//	server.HandleFunc("/api/v1/users", userHandler)
//	go server.Start(ctx)
type HTTPServer struct {
	// server is the underlying HTTP server instance
	server *http.Server

	// port is the listening port for this server
	port string

	// mux is the HTTP request multiplexer
	mux *http.ServeMux

	// logger is the structured logger for this server
	logger log.Logger
}

// NewHTTPServer creates a new business HTTP server with the specified port.
// The port should be in the format ":8080" or "0.0.0.0:8080".
//
// Parameters:
//   - port: The listening address and port for the HTTP server
//
// Returns:
//   - *HTTPServer: A new HTTP server instance ready for configuration
//
// Example:
//
//	server := NewHTTPServer(":8080")
//	server.HandleFunc("/api/v1/users", userHandler)
func NewHTTPServer(port string) *HTTPServer {
	mux := http.NewServeMux()

	return &HTTPServer{
		server: &http.Server{
			Addr:         port,
			Handler:      mux,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
			IdleTimeout:  120 * time.Second,
		},
		port:   port,
		mux:    mux,
		logger: log.Default(),
	}
}

// HandleFunc registers a handler function for the given pattern.
// This method follows the standard http.ServeMux pattern.
//
// Parameters:
//   - pattern: The URL pattern to match (e.g., "/api/v1/users")
//   - handler: The handler function to execute for matching requests
//
// Example:
//
//	server.HandleFunc("/api/v1/users", func(w http.ResponseWriter, r *http.Request) {
//	    w.WriteHeader(http.StatusOK)
//	    w.Write([]byte("Hello, World!"))
//	})
func (s *HTTPServer) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	s.mux.HandleFunc(pattern, handler)
	s.logger.Info("HTTP route registered",
		log.Field{Key: "pattern", Value: pattern},
		log.Field{Key: "port", Value: s.port})
}

// Handle registers a handler for the given pattern.
// This method follows the standard http.ServeMux pattern.
//
// Parameters:
//   - pattern: The URL pattern to match (e.g., "/api/v1/users")
//   - handler: The handler to execute for matching requests
//
// Example:
//
//	server.Handle("/api/v1/users", http.HandlerFunc(userHandler))
func (s *HTTPServer) Handle(pattern string, handler http.Handler) {
	s.mux.Handle(pattern, handler)
	s.logger.Info("HTTP route registered",
		log.Field{Key: "pattern", Value: pattern},
		log.Field{Key: "port", Value: s.port})
}

// Start begins serving HTTP requests on the configured port.
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
//   - Logs server startup information
//   - Handles graceful shutdown on context cancellation
//   - Returns immediately if server is already running
//
// Example:
//
//	ctx, cancel := context.WithCancel(context.Background())
//	go func() {
//	    if err := server.Start(ctx); err != nil {
//	        log.Error("HTTP server failed", log.Field{Key: "error", Value: err})
//	    }
//	}()
//	defer cancel()
func (s *HTTPServer) Start(ctx context.Context) error {
	s.logger.Info("Starting business HTTP server",
		log.Field{Key: "port", Value: s.port},
		log.Field{Key: "address", Value: s.server.Addr})

	// Create a channel to receive server errors
	errChan := make(chan error, 1)

	// Start server in a goroutine
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- fmt.Errorf("HTTP server failed: %w", err)
		}
	}()

	// Wait for context cancellation or server error
	select {
	case <-ctx.Done():
		s.logger.Info("Shutting down HTTP server",
			log.Field{Key: "reason", Value: "context_canceled"})

		// Create shutdown context with timeout
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Attempt graceful shutdown
		if err := s.server.Shutdown(shutdownCtx); err != nil {
			s.logger.Error("HTTP server shutdown failed",
				log.Field{Key: "error", Value: err})
			return fmt.Errorf("HTTP server shutdown failed: %w", err)
		}

		s.logger.Info("HTTP server shutdown completed")
		return nil

	case err := <-errChan:
		return err
	}
}

// Stop gracefully shuts down the HTTP server.
// This method is provided for compatibility with the Service interface.
// In practice, shutdown is handled by the Start method when context is canceled.
//
// Parameters:
//   - ctx: Context for timeout control during shutdown
//
// Returns:
//   - error: Returns error if shutdown fails
func (s *HTTPServer) Stop(ctx context.Context) error {
	s.logger.Info("Stopping HTTP server")
	return s.server.Shutdown(ctx)
}

// GetPort returns the configured port for this server.
// This method is useful for logging and monitoring purposes.
//
// Returns:
//   - string: The port string (e.g., ":8080")
func (s *HTTPServer) GetPort() string {
	return s.port
}

// GetServer returns the underlying http.Server instance.
// This method is provided for advanced use cases where direct access
// to the http.Server is needed for custom configuration.
//
// Returns:
//   - *http.Server: The underlying HTTP server instance
//
// Note: Modifying the returned server after Start() has been called
// may cause undefined behavior.
func (s *HTTPServer) GetServer() *http.Server {
	return s.server
}

// SetLogger sets the logger for this HTTP server.
// This method allows customization of logging behavior.
//
// Parameters:
//   - logger: The logger instance to use for this server
func (s *HTTPServer) SetLogger(logger interface{}) {
	if l, ok := logger.(log.Logger); ok {
		s.logger = l
	}
}
