package service

import "context"

// Service defines the interface for long-running service components.
// Services represent application components that run continuously until
// explicitly stopped, such as HTTP servers, gRPC servers, or background workers.
//
// Lifecycle:
//  1. Service is created and registered with Launcher
//  2. Launcher calls Start() to begin operation
//  3. Start() should block until service stops or context is cancelled
//  4. Launcher calls Stop() during graceful shutdown
//
// Implementations must be safe for concurrent Start/Stop calls.
type Service interface {
	// Start begins the service's operation and blocks until stopped.
	// The service should respect context cancellation for graceful shutdown.
	//
	// Parameters:
	//   - ctx: Context for cancellation and deadline control
	//
	// Returns:
	//   - error: Returns error if service fails to start or encounters
	//     a fatal runtime error. Returns nil on normal shutdown.
	//
	// Behavior:
	//   - Must block until service is stopped or context is cancelled
	//   - Should return promptly when context is cancelled
	//   - Clean up resources before returning
	//
	// Example implementation:
	//   func (s *HTTPServer) Start(ctx context.Context) error {
	//       errCh := make(chan error, 1)
	//       go func() {
	//           errCh <- s.server.ListenAndServe()
	//       }()
	//       select {
	//       case err := <-errCh:
	//           return err
	//       case <-ctx.Done():
	//           return s.server.Shutdown(context.Background())
	//       }
	//   }
	Start(ctx context.Context) error

	// Stop performs graceful shutdown of the service.
	// Called by Launcher when application is terminating.
	//
	// Parameters:
	//   - ctx: Context with timeout for shutdown completion
	//
	// Returns:
	//   - error: Returns error if shutdown fails or times out
	//
	// Behavior:
	//   - Must complete shutdown within context deadline
	//   - Should close connections and release resources
	//   - Should not start new work after being called
	//
	// Example implementation:
	//   func (s *HTTPServer) Stop(ctx context.Context) error {
	//       return s.server.Shutdown(ctx)
	//   }
	Stop(ctx context.Context) error
}

// Initializer defines the interface for one-time initialization tasks.
// Initializers run sequentially before services start, setting up
// dependencies like database connections, caches, and external clients.
//
// Use cases:
//   - Database connection and migration
//   - Cache warming
//   - External service client initialization
//   - Configuration validation
//   - Feature flag loading
//
// Initializers are executed in registration order, allowing controlled
// dependency setup sequences.
type Initializer interface {
	// Init performs one-time initialization logic.
	// Called by Launcher during application startup, before any services start.
	//
	// Parameters:
	//   - ctx: Context for timeout control and cancellation
	//
	// Returns:
	//   - error: Returns error if initialization fails. A failed initializer
	//     prevents application startup and triggers immediate shutdown.
	//
	// Behavior:
	//   - Must be idempotent (safe to call multiple times)
	//   - Must complete within reasonable time (use context timeout)
	//   - Should log initialization steps for debugging
	//   - Critical errors should return immediately
	//
	// Example implementation:
	//   func (d *DatabaseInitializer) Init(ctx context.Context) error {
	//       log.Info("Connecting to database...")
	//       db, err := sql.Open("mysql", d.dsn)
	//       if err != nil {
	//           return fmt.Errorf("failed to open database: %w", err)
	//       }
	//       if err := db.PingContext(ctx); err != nil {
	//           return fmt.Errorf("failed to ping database: %w", err)
	//       }
	//       d.db = db
	//       log.Info("Database connection established")
	//       return nil
	//   }
	Init(ctx context.Context) error
}
