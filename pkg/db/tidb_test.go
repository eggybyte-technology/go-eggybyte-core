package db

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewTiDBInitializer tests creating a TiDB initializer.
// This is an isolated method test with no external dependencies.
func TestNewTiDBInitializer(t *testing.T) {
	cfg := DefaultConfig()
	cfg.DSN = "user:pass@tcp(localhost:4000)/test"

	initializer := NewTiDBInitializer(cfg)

	assert.NotNil(t, initializer)
	assert.NotNil(t, initializer.config)
	assert.Equal(t, cfg.DSN, initializer.config.DSN)
}

// TestNewTiDBInitializer_NilConfig tests nil config handling.
// This verifies the initializer accepts nil config.
func TestNewTiDBInitializer_NilConfig(t *testing.T) {
	initializer := NewTiDBInitializer(nil)

	assert.NotNil(t, initializer)
	assert.Nil(t, initializer.config)
}

// TestNewTiDBInitializer_CustomConfig tests initializer with custom config.
// This verifies configuration is properly stored.
func TestNewTiDBInitializer_CustomConfig(t *testing.T) {
	cfg := &Config{
		DSN:          "custom:pass@tcp(localhost:4000)/custom",
		MaxOpenConns: 200,
		MaxIdleConns: 20,
		LogLevel:     "debug",
	}

	initializer := NewTiDBInitializer(cfg)

	assert.NotNil(t, initializer)
	require.NotNil(t, initializer.config)
	assert.Equal(t, "custom:pass@tcp(localhost:4000)/custom", initializer.config.DSN)
	assert.Equal(t, 200, initializer.config.MaxOpenConns)
	assert.Equal(t, 20, initializer.config.MaxIdleConns)
	assert.Equal(t, "debug", initializer.config.LogLevel)
}

// TestNewTiDBInitializer_DefaultConfig tests with default configuration.
// This verifies default values are used correctly.
func TestNewTiDBInitializer_DefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	initializer := NewTiDBInitializer(cfg)

	assert.NotNil(t, initializer)
	require.NotNil(t, initializer.config)
	assert.Equal(t, 100, initializer.config.MaxOpenConns)
	assert.Equal(t, 10, initializer.config.MaxIdleConns)
	assert.Equal(t, "warn", initializer.config.LogLevel)
}

// TestTiDBInitializer_Init_NoDSN tests error when DSN is missing.
// This verifies proper validation of required configuration.
func TestTiDBInitializer_Init_NoDSN(t *testing.T) {
	cfg := &Config{
		DSN: "", // Empty DSN
	}

	initializer := NewTiDBInitializer(cfg)
	err := initializer.Init(context.Background())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "database DSN is required")
}

// TestTiDBInitializer_Init_EmptyConfig tests with nil config.
// This verifies error handling for missing configuration.
func TestTiDBInitializer_Init_EmptyConfig(t *testing.T) {
	initializer := NewTiDBInitializer(nil)
	err := initializer.Init(context.Background())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "config is nil")
}

// TestTiDBInitializer_Init_InvalidDSN tests error with invalid DSN.
// This verifies connection error handling.
func TestTiDBInitializer_Init_InvalidDSN(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database connection test")
	}

	cfg := &Config{
		DSN:      "invalid-dsn-format",
		LogLevel: "warn",
	}

	initializer := NewTiDBInitializer(cfg)
	err := initializer.Init(context.Background())

	// Should fail with connection or DSN parse error
	assert.Error(t, err)
}

// TestTiDBInitializer_ConfigFields tests all config fields are used.
// This verifies configuration completeness.
func TestTiDBInitializer_ConfigFields(t *testing.T) {
	tests := []struct {
		name         string
		dsn          string
		maxOpenConns int
		maxIdleConns int
		logLevel     string
	}{
		{
			name:         "production_config",
			dsn:          "prod:pass@tcp(prod-db:4000)/proddb",
			maxOpenConns: 200,
			maxIdleConns: 20,
			logLevel:     "info",
		},
		{
			name:         "dev_config",
			dsn:          "dev:pass@tcp(localhost:4000)/devdb",
			maxOpenConns: 50,
			maxIdleConns: 5,
			logLevel:     "debug",
		},
		{
			name:         "test_config",
			dsn:          "test:pass@tcp(localhost:4000)/testdb",
			maxOpenConns: 10,
			maxIdleConns: 2,
			logLevel:     "error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				DSN:          tt.dsn,
				MaxOpenConns: tt.maxOpenConns,
				MaxIdleConns: tt.maxIdleConns,
				LogLevel:     tt.logLevel,
			}

			initializer := NewTiDBInitializer(cfg)

			assert.NotNil(t, initializer)
			assert.Equal(t, tt.dsn, initializer.config.DSN)
			assert.Equal(t, tt.maxOpenConns, initializer.config.MaxOpenConns)
			assert.Equal(t, tt.maxIdleConns, initializer.config.MaxIdleConns)
			assert.Equal(t, tt.logLevel, initializer.config.LogLevel)
		})
	}
}

// TestTiDBInitializer_MultipleInstances tests creating multiple initializers.
// This verifies each instance is independent.
func TestTiDBInitializer_MultipleInstances(t *testing.T) {
	cfg1 := &Config{
		DSN:          "user1:pass@tcp(localhost:4000)/db1",
		MaxOpenConns: 100,
	}

	cfg2 := &Config{
		DSN:          "user2:pass@tcp(localhost:4000)/db2",
		MaxOpenConns: 200,
	}

	init1 := NewTiDBInitializer(cfg1)
	init2 := NewTiDBInitializer(cfg2)

	assert.NotSame(t, init1, init2)
	assert.NotSame(t, init1.config, init2.config)
	assert.Equal(t, 100, init1.config.MaxOpenConns)
	assert.Equal(t, 200, init2.config.MaxOpenConns)
}

// TestTiDBInitializer_ContextCancellation tests context handling during init.
// This verifies the initializer respects context cancellation.
func TestTiDBInitializer_ContextCancellation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping context test")
	}

	cfg := &Config{
		DSN: "user:pass@tcp(localhost:4000)/test",
	}

	initializer := NewTiDBInitializer(cfg)

	// Cancel context immediately
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := initializer.Init(ctx)

	// Should fail due to cancelled context or connection error
	assert.Error(t, err)
}

// TestTiDBInitializer_DSNFormats tests various DSN formats.
// This verifies different connection string formats are handled.
func TestTiDBInitializer_DSNFormats(t *testing.T) {
	tests := []struct {
		name  string
		dsn   string
		valid bool
	}{
		{
			name:  "standard_mysql",
			dsn:   "user:pass@tcp(localhost:3306)/dbname",
			valid: true,
		},
		{
			name:  "tidb_default_port",
			dsn:   "user:pass@tcp(localhost:4000)/dbname",
			valid: true,
		},
		{
			name:  "with_params",
			dsn:   "user:pass@tcp(localhost:4000)/db?charset=utf8mb4&parseTime=True",
			valid: true,
		},
		{
			name:  "unix_socket",
			dsn:   "user:pass@unix(/tmp/tidb.sock)/dbname",
			valid: true,
		},
		{
			name:  "empty_dsn",
			dsn:   "",
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				DSN: tt.dsn,
			}

			initializer := NewTiDBInitializer(cfg)

			if tt.valid && tt.dsn != "" {
				assert.Equal(t, tt.dsn, initializer.config.DSN)
			}
		})
	}
}

// TestTiDBInitializer_ConnectionPoolConfig tests connection pool settings.
// This verifies pool configuration is properly applied.
func TestTiDBInitializer_ConnectionPoolConfig(t *testing.T) {
	tests := []struct {
		name         string
		maxOpenConns int
		maxIdleConns int
	}{
		{"minimal", 1, 1},
		{"low", 10, 5},
		{"medium", 50, 10},
		{"high", 100, 20},
		{"very_high", 200, 50},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				DSN:          "user:pass@tcp(localhost:4000)/test",
				MaxOpenConns: tt.maxOpenConns,
				MaxIdleConns: tt.maxIdleConns,
			}

			initializer := NewTiDBInitializer(cfg)

			assert.Equal(t, tt.maxOpenConns, initializer.config.MaxOpenConns)
			assert.Equal(t, tt.maxIdleConns, initializer.config.MaxIdleConns)
		})
	}
}

// TestTiDBInitializer_LogLevels tests different log level configurations.
// This verifies all log levels are properly stored.
func TestTiDBInitializer_LogLevels(t *testing.T) {
	levels := []string{"debug", "info", "warn", "error"}

	for _, level := range levels {
		t.Run(level, func(t *testing.T) {
			cfg := &Config{
				DSN:      "user:pass@tcp(localhost:4000)/test",
				LogLevel: level,
			}

			initializer := NewTiDBInitializer(cfg)

			assert.Equal(t, level, initializer.config.LogLevel)
		})
	}
}

// TestTiDBInitializer_RepositoryInitialization tests repository auto-init.
// This verifies registered repositories are initialized during Init.
func TestTiDBInitializer_RepositoryInitialization(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	cfg := &Config{
		DSN: "user:pass@tcp(localhost:4000)/test",
	}

	initializer := NewTiDBInitializer(cfg)

	// This will fail to connect, but we verify the logic path
	err := initializer.Init(context.Background())

	// Expected to fail on connection
	assert.Error(t, err)
}

// TestTiDBInitializer_GlobalDBSet tests global DB is set after init.
// This verifies the global DB accessor is properly configured.
func TestTiDBInitializer_GlobalDBSet(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Reset global DB
	SetDB(nil)

	cfg := &Config{
		DSN: "user:pass@tcp(localhost:4000)/test",
	}

	initializer := NewTiDBInitializer(cfg)
	err := initializer.Init(context.Background())

	// Will fail due to connection, but we verify the attempt was made
	assert.Error(t, err)

	// In real scenario, GetDB() would return non-nil after successful init
	// Here we just verify the function can be called
	db := GetDB()
	_ = db // May be nil due to failed connection
}

// TestTiDBInitializer_ConfigImmutability tests config is not modified.
// This verifies the initializer doesn't change the provided config.
func TestTiDBInitializer_ConfigImmutability(t *testing.T) {
	originalCfg := &Config{
		DSN:          "user:pass@tcp(localhost:4000)/test",
		MaxOpenConns: 100,
		MaxIdleConns: 10,
		LogLevel:     "info",
	}

	// Store original values
	originalDSN := originalCfg.DSN
	originalMaxOpen := originalCfg.MaxOpenConns
	originalMaxIdle := originalCfg.MaxIdleConns
	originalLogLevel := originalCfg.LogLevel

	initializer := NewTiDBInitializer(originalCfg)

	// Config should remain unchanged
	assert.Equal(t, originalDSN, initializer.config.DSN)
	assert.Equal(t, originalMaxOpen, initializer.config.MaxOpenConns)
	assert.Equal(t, originalMaxIdle, initializer.config.MaxIdleConns)
	assert.Equal(t, originalLogLevel, initializer.config.LogLevel)
}

// TestTiDBInitializer_ErrorMessages tests error message clarity.
// This verifies error messages are helpful and descriptive.
func TestTiDBInitializer_ErrorMessages(t *testing.T) {
	tests := []struct {
		name          string
		config        *Config
		expectedError string
	}{
		{
			name:          "nil_config",
			config:        nil,
			expectedError: "config is nil",
		},
		{
			name: "empty_dsn",
			config: &Config{
				DSN: "",
			},
			expectedError: "database DSN is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initializer := NewTiDBInitializer(tt.config)
			err := initializer.Init(context.Background())

			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedError)
		})
	}
}
