package db

import (
	"context"
	"fmt"

	"github.com/eggybyte-technology/go-eggybyte-core/pkg/log"
)

// TiDBInitializer implements service.Initializer interface for database setup.
// It handles database connection, table creation, and schema migration during
// application startup.
//
// This initializer:
//  1. Establishes database connection using provided configuration
//  2. Verifies connection health with ping
//  3. Initializes all registered repository tables
//
// The initializer works with both TiDB and MySQL databases through the
// MySQL-compatible driver.
//
// Usage:
//
//	cfg := db.DefaultConfig()
//	cfg.DSN = os.Getenv("DATABASE_DSN")
//	initializer := db.NewTiDBInitializer(cfg)
//	launcher.AddInitializer(initializer)
type TiDBInitializer struct {
	config *Config
}

// NewTiDBInitializer creates a new database initializer with the given configuration.
//
// Parameters:
//   - cfg: Database configuration including DSN and connection pool settings
//
// Returns:
//   - *TiDBInitializer: Initializer instance ready to be registered with launcher
//
// Example:
//
//	cfg := &db.Config{
//	    DSN: "root:password@tcp(localhost:4000)/mydb?charset=utf8mb4",
//	    MaxOpenConns: 100,
//	    MaxIdleConns: 10,
//	}
//	initializer := db.NewTiDBInitializer(cfg)
func NewTiDBInitializer(cfg *Config) *TiDBInitializer {
	return &TiDBInitializer{
		config: cfg,
	}
}

// Init establishes database connection and initializes all repository tables.
// This method is called by the service launcher during application startup.
//
// Initialization steps:
//  1. Connect to database using provided DSN
//  2. Verify connection health
//  3. Initialize all registered repository tables via registry
//
// Parameters:
//   - ctx: Context for timeout control and cancellation
//
// Returns:
//   - error: Returns error if connection fails or table initialization fails
//
// Behavior:
//   - Sets global database connection via SetDB()
//   - Calls InitializeAllTables() to create/migrate tables
//   - Returns immediately on any error
//
// Example:
//
//	initializer := db.NewTiDBInitializer(cfg)
//	if err := initializer.Init(ctx); err != nil {
//	    log.Fatal("Database initialization failed", log.Field{Key: "error", Value: err})
//	}
func (t *TiDBInitializer) Init(ctx context.Context) error {
	log.Info("Initializing TiDB/MySQL database connection")

	// Validate configuration
	if t.config == nil {
		return fmt.Errorf("config is nil")
	}

	if t.config.DSN == "" {
		return fmt.Errorf("database DSN is required")
	}

	// Establish database connection
	db, err := Connect(t.config)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Info("Database connection established successfully")

	// Initialize all registered repository tables
	if err := InitializeAllTables(ctx, db); err != nil {
		return fmt.Errorf("failed to initialize tables: %w", err)
	}

	log.Info("Database initialization completed successfully")
	return nil
}
