package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestReadFromEnv_WithValidEnv tests loading configuration from environment variables.
// This is an isolated method test that sets up clean environment state.
func TestReadFromEnv_WithValidEnv(t *testing.T) {
	// Setup environment variables
	os.Setenv("SERVICE_NAME", "test-service")
	os.Setenv("ENVIRONMENT", "testing")
	os.Setenv("PORT", "8080")
	os.Setenv("METRICS_PORT", "9090")
	defer cleanupEnv()

	var cfg Config
	err := ReadFromEnv(&cfg)

	require.NoError(t, err)
	assert.Equal(t, "test-service", cfg.ServiceName)
	assert.Equal(t, "testing", cfg.Environment)
	assert.Equal(t, 8080, cfg.Port)
	assert.Equal(t, 9090, cfg.MetricsPort)
}

// TestReadFromEnv_MissingRequired tests error handling for missing required fields.
// This verifies that required environment variables are enforced.
func TestReadFromEnv_MissingRequired(t *testing.T) {
	// Clear all environment variables
	cleanupEnv()

	var cfg Config
	err := ReadFromEnv(&cfg)

	assert.Error(t, err, "Expected error when required SERVICE_NAME is missing")
	assert.Contains(t, err.Error(), "failed to process environment configuration")
}

// TestReadFromEnv_WithDefaults tests that default values are applied.
// This verifies the struct tag default mechanism works correctly.
func TestReadFromEnv_WithDefaults(t *testing.T) {
	// Only set required field
	os.Setenv("SERVICE_NAME", "test-service")
	defer cleanupEnv()

	var cfg Config
	err := ReadFromEnv(&cfg)

	require.NoError(t, err)
	assert.Equal(t, "development", cfg.Environment, "Expected default environment")
	assert.Equal(t, 8080, cfg.Port, "Expected default port")
	assert.Equal(t, 9090, cfg.MetricsPort, "Expected default metrics port")
	assert.Equal(t, "info", cfg.LogLevel, "Expected default log level")
	assert.Equal(t, "json", cfg.LogFormat, "Expected default log format")
}

// TestReadFromEnv_TypeConversion tests integer and boolean field parsing.
// This verifies that envconfig correctly converts string env vars to typed fields.
func TestReadFromEnv_TypeConversion(t *testing.T) {
	os.Setenv("SERVICE_NAME", "test-service")
	os.Setenv("PORT", "8888")
	os.Setenv("DATABASE_MAX_OPEN_CONNS", "200")
	os.Setenv("ENABLE_K8S_CONFIG_WATCH", "true")
	defer cleanupEnv()

	var cfg Config
	err := ReadFromEnv(&cfg)

	require.NoError(t, err)
	assert.Equal(t, 8888, cfg.Port)
	assert.Equal(t, 200, cfg.DatabaseMaxOpenConns)
	assert.True(t, cfg.EnableK8sConfigWatch)
}

// TestMustReadFromEnv_Success tests the panic-free path.
// This verifies MustReadFromEnv works correctly with valid config.
func TestMustReadFromEnv_Success(t *testing.T) {
	os.Setenv("SERVICE_NAME", "test-service")
	defer cleanupEnv()

	var cfg Config

	assert.NotPanics(t, func() {
		MustReadFromEnv(&cfg)
	}, "MustReadFromEnv should not panic with valid environment")

	assert.Equal(t, "test-service", cfg.ServiceName)
}

// TestMustReadFromEnv_Panic tests that MustReadFromEnv panics on error.
// This verifies the panic behavior for missing required fields.
func TestMustReadFromEnv_Panic(t *testing.T) {
	cleanupEnv() // No SERVICE_NAME set

	var cfg Config

	assert.Panics(t, func() {
		MustReadFromEnv(&cfg)
	}, "MustReadFromEnv should panic when required fields are missing")
}

// TestValidateConfig_ValidConfig tests successful validation.
// This is an isolated method test with no external dependencies.
func TestValidateConfig_ValidConfig(t *testing.T) {
	cfg := &Config{
		ServiceName: "test-service",
		Port:        8080,
		MetricsPort: 9090,
		LogLevel:    "info",
	}

	err := ValidateConfig(cfg)

	assert.NoError(t, err, "Valid configuration should pass validation")
}

// TestValidateConfig_EmptyServiceName tests service name validation.
// This verifies the required field validation logic.
func TestValidateConfig_EmptyServiceName(t *testing.T) {
	cfg := &Config{
		ServiceName: "",
		Port:        8080,
		MetricsPort: 9090,
	}

	err := ValidateConfig(cfg)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "service name cannot be empty")
}

// TestValidateConfig_InvalidPort tests port range validation.
// This verifies port boundary checking (1-65535).
func TestValidateConfig_InvalidPort(t *testing.T) {
	tests := []struct {
		name string
		port int
	}{
		{"zero port", 0},
		{"negative port", -1},
		{"too large port", 65536},
		{"way too large port", 100000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				ServiceName: "test-service",
				Port:        tt.port,
				MetricsPort: 9090,
				LogLevel:    "info",
			}

			err := ValidateConfig(cfg)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), "port must be between 1 and 65535")
		})
	}
}

// TestValidateConfig_InvalidMetricsPort tests metrics port validation.
// This verifies metrics port boundary checking.
func TestValidateConfig_InvalidMetricsPort(t *testing.T) {
	cfg := &Config{
		ServiceName: "test-service",
		Port:        8080,
		MetricsPort: 0,
		LogLevel:    "info",
	}

	err := ValidateConfig(cfg)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "metrics port must be between 1 and 65535")
}

// TestValidateConfig_SamePort tests that business and metrics ports must differ.
// This verifies the port conflict detection logic.
func TestValidateConfig_SamePort(t *testing.T) {
	cfg := &Config{
		ServiceName: "test-service",
		Port:        8080,
		MetricsPort: 8080, // Same as business port
		LogLevel:    "info",
	}

	err := ValidateConfig(cfg)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "business port and metrics port must be different")
}

// TestValidateConfig_InvalidLogLevel tests log level validation.
// This verifies that only allowed log levels are accepted.
func TestValidateConfig_InvalidLogLevel(t *testing.T) {
	cfg := &Config{
		ServiceName: "test-service",
		Port:        8080,
		MetricsPort: 9090,
		LogLevel:    "invalid",
	}

	err := ValidateConfig(cfg)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid log level")
}

// TestValidateConfig_ValidLogLevels tests all valid log levels.
// This is an isolated test verifying the log level validation logic.
func TestValidateConfig_ValidLogLevels(t *testing.T) {
	validLevels := []string{"debug", "info", "warn", "error", "fatal"}

	for _, level := range validLevels {
		t.Run(level, func(t *testing.T) {
			cfg := &Config{
				ServiceName: "test-service",
				Port:        8080,
				MetricsPort: 9090,
				LogLevel:    level,
			}

			err := ValidateConfig(cfg)

			assert.NoError(t, err, "Log level '%s' should be valid", level)
		})
	}
}

// TestValidateConfig_K8sWatch_MissingNamespace tests K8s validation.
// This verifies that K8s config requires namespace when enabled.
func TestValidateConfig_K8sWatch_MissingNamespace(t *testing.T) {
	cfg := &Config{
		ServiceName:          "test-service",
		Port:                 8080,
		MetricsPort:          9090,
		LogLevel:             "info",
		EnableK8sConfigWatch: true,
		K8sNamespace:         "", // Missing
		K8sConfigMapName:     "config",
	}

	err := ValidateConfig(cfg)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "kubernetes namespace required when config watch enabled")
}

// TestValidateConfig_K8sWatch_MissingConfigMapName tests K8s validation.
// This verifies that K8s config requires configmap name when enabled.
func TestValidateConfig_K8sWatch_MissingConfigMapName(t *testing.T) {
	cfg := &Config{
		ServiceName:          "test-service",
		Port:                 8080,
		MetricsPort:          9090,
		LogLevel:             "info",
		EnableK8sConfigWatch: true,
		K8sNamespace:         "default",
		K8sConfigMapName:     "", // Missing
	}

	err := ValidateConfig(cfg)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "kubernetes configmap name required when config watch enabled")
}

// TestValidateConfig_K8sWatch_ValidConfig tests K8s validation with complete config.
// This verifies that K8s config passes validation when all fields are present.
func TestValidateConfig_K8sWatch_ValidConfig(t *testing.T) {
	cfg := &Config{
		ServiceName:          "test-service",
		Port:                 8080,
		MetricsPort:          9090,
		LogLevel:             "info",
		EnableK8sConfigWatch: true,
		K8sNamespace:         "default",
		K8sConfigMapName:     "app-config",
	}

	err := ValidateConfig(cfg)

	assert.NoError(t, err)
}

// cleanupEnv removes all test environment variables.
// Helper function for test isolation.
func cleanupEnv() {
	vars := []string{
		"SERVICE_NAME", "ENVIRONMENT", "PORT", "METRICS_PORT",
		"LOG_LEVEL", "LOG_FORMAT", "DATABASE_DSN",
		"DATABASE_MAX_OPEN_CONNS", "DATABASE_MAX_IDLE_CONNS",
		"ENABLE_K8S_CONFIG_WATCH", "K8S_NAMESPACE", "K8S_CONFIGMAP_NAME",
	}
	for _, v := range vars {
		os.Unsetenv(v)
	}
}
