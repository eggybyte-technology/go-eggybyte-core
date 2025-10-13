package db

import (
	"fmt"
	"sync"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/eggybyte-technology/go-eggybyte-core/pkg/log"
)

// global DB holds the singleton database connection.
// Initialized by Connect() and accessed via GetDB().
var (
	globalDB *gorm.DB
	dbMutex  sync.RWMutex
)

// Config holds database connection configuration parameters.
// These settings control connection pooling, timeouts, and behavior.
type Config struct {
	// DSN is the Data Source Name for database connection.
	// Format: "username:password@tcp(host:port)/dbname?charset=utf8mb4&parseTime=True"
	DSN string

	// MaxOpenConns sets the maximum number of open database connections.
	// Default: 100
	MaxOpenConns int

	// MaxIdleConns sets the maximum number of idle connections in the pool.
	// Default: 10
	MaxIdleConns int

	// ConnMaxLifetime sets the maximum lifetime of a connection.
	// Connections older than this duration are closed and recreated.
	// Default: 1 hour
	ConnMaxLifetime time.Duration

	// ConnMaxIdleTime sets the maximum idle time for a connection.
	// Idle connections exceeding this duration are closed.
	// Default: 10 minutes
	ConnMaxIdleTime time.Duration

	// LogLevel sets the GORM logger level.
	// Valid values: "silent", "error", "warn", "info"
	// Default: "warn"
	LogLevel string
}

// DefaultConfig returns database configuration with sensible defaults.
// Applications can override specific fields as needed.
//
// Returns:
//   - *Config: Configuration with default values
//
// Example:
//
//	cfg := db.DefaultConfig()
//	cfg.DSN = "user:pass@tcp(localhost:3306)/mydb"
//	cfg.MaxOpenConns = 200
func DefaultConfig() *Config {
	return &Config{
		MaxOpenConns:    100,
		MaxIdleConns:    10,
		ConnMaxLifetime: time.Hour,
		ConnMaxIdleTime: 10 * time.Minute,
		LogLevel:        "warn",
	}
}

// Connect establishes a database connection using the provided configuration.
// This function initializes the global database connection and configures
// connection pooling parameters.
//
// The connection uses MySQL driver which is compatible with TiDB.
//
// Parameters:
//   - cfg: Database configuration parameters
//
// Returns:
//   - *gorm.DB: The established database connection
//   - error: Returns error if connection fails
//
// Thread Safety: Safe for concurrent calls, but typically called once during startup.
//
// Example:
//
//	cfg := db.DefaultConfig()
//	cfg.DSN = os.Getenv("DATABASE_DSN")
//	db, err := db.Connect(cfg)
//	if err != nil {
//	    log.Fatal("Failed to connect to database", log.Field{Key: "error", Value: err})
//	}
func Connect(cfg *Config) (*gorm.DB, error) {
	// Parse log level
	gormLogLevel := parseLogLevel(cfg.LogLevel)

	// Configure GORM
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(gormLogLevel),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	}

	// Open database connection
	db, err := gorm.Open(mysql.Open(cfg.DSN), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Get underlying sql.DB for connection pool configuration
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	// Verify connection
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Set global DB
	SetDB(db)

	log.Info("Database connection established",
		log.Field{Key: "max_open_conns", Value: cfg.MaxOpenConns},
		log.Field{Key: "max_idle_conns", Value: cfg.MaxIdleConns})

	return db, nil
}

// GetDB returns the global database connection.
// If no connection has been established, returns nil.
//
// Returns:
//   - *gorm.DB: The global database connection, or nil if not initialized
//
// Thread Safety: Safe for concurrent access.
//
// Example:
//
//	db := db.GetDB()
//	if db == nil {
//	    log.Fatal("Database not initialized")
//	}
//	db.Find(&users)
func GetDB() *gorm.DB {
	dbMutex.RLock()
	defer dbMutex.RUnlock()
	return globalDB
}

// SetDB updates the global database connection.
// Useful for testing or advanced configuration scenarios.
//
// Parameters:
//   - db: Database connection to use as global instance
//
// Thread Safety: Safe for concurrent access.
func SetDB(db *gorm.DB) {
	dbMutex.Lock()
	defer dbMutex.Unlock()
	globalDB = db
}

// Close closes the global database connection and releases resources.
// Should be called during application shutdown.
//
// Returns:
//   - error: Returns error if closing fails
//
// Example:
//
//	defer db.Close()
func Close() error {
	db := GetDB()
	if db == nil {
		return nil
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}

	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("failed to close database: %w", err)
	}

	log.Info("Database connection closed")
	return nil
}

// parseLogLevel converts string log level to GORM logger level.
func parseLogLevel(level string) logger.LogLevel {
	switch level {
	case "silent":
		return logger.Silent
	case "error":
		return logger.Error
	case "warn":
		return logger.Warn
	case "info":
		return logger.Info
	default:
		return logger.Warn
	}
}
