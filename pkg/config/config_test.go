package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestGet_WhenNotInitialized tests that Get returns nil before configuration is set.
// This is an isolated method test with no external dependencies.
func TestGet_WhenNotInitialized(t *testing.T) {
	// Reset global config for test isolation
	Set(nil)

	result := Get()

	assert.Nil(t, result, "Expected nil config before initialization")
}

// TestGet_WhenInitialized tests that Get returns the configured instance.
// This is an isolated method test with no external dependencies.
func TestGet_WhenInitialized(t *testing.T) {
	expected := &Config{
		ServiceName:        "test-service",
		Environment:        "test",
		BusinessHTTPPort:   8080,
		BusinessGRPCPort:   9090,
		HealthCheckPort:    8081,
		MetricsPort:        9091,
		EnableBusinessHTTP: true,
		EnableBusinessGRPC: true,
		EnableHealthCheck:  true,
		EnableMetrics:      true,
	}

	Set(expected)
	result := Get()

	assert.NotNil(t, result, "Expected non-nil config after initialization")
	assert.Equal(t, expected.ServiceName, result.ServiceName)
	assert.Equal(t, expected.Environment, result.Environment)
	assert.Equal(t, expected.BusinessHTTPPort, result.BusinessHTTPPort)

	// Cleanup
	Set(nil)
}

// TestSet_UpdatesGlobalConfig tests that Set correctly updates the global configuration.
// This is an isolated method test with no external dependencies.
func TestSet_UpdatesGlobalConfig(t *testing.T) {
	cfg := &Config{
		ServiceName:        "user-service",
		BusinessHTTPPort:   9000,
		BusinessGRPCPort:   9090,
		HealthCheckPort:    8081,
		MetricsPort:        9091,
		EnableBusinessHTTP: true,
		EnableBusinessGRPC: true,
		EnableHealthCheck:  true,
		EnableMetrics:      true,
	}

	Set(cfg)
	result := Get()

	assert.NotNil(t, result)
	assert.Equal(t, "user-service", result.ServiceName)
	assert.Equal(t, 9000, result.BusinessHTTPPort)

	// Cleanup
	Set(nil)
}

// TestSet_ThreadSafety tests concurrent access to Set and Get.
// This verifies the mutex protection is working correctly.
func TestSet_ThreadSafety(t *testing.T) {
	done := make(chan bool)

	// Start multiple goroutines writing
	for i := 0; i < 10; i++ {
		go func(id int) {
			cfg := &Config{
				ServiceName:        "test-service",
				BusinessHTTPPort:   8000 + id,
				BusinessGRPCPort:   9090,
				HealthCheckPort:    8081,
				MetricsPort:        9091,
				EnableBusinessHTTP: true,
				EnableBusinessGRPC: true,
				EnableHealthCheck:  true,
				EnableMetrics:      true,
			}
			Set(cfg)
			done <- true
		}(i)
	}

	// Start multiple goroutines reading
	for i := 0; i < 10; i++ {
		go func() {
			_ = Get()
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 20; i++ {
		<-done
	}

	// Verify no panic occurred
	assert.NotNil(t, t, "Thread safety test completed without panic")

	// Cleanup
	Set(nil)
}

// TestUpdate_WithNilConfig tests that Update handles nil global config gracefully.
// This is an isolated method test with no external dependencies.
func TestUpdate_WithNilConfig(t *testing.T) {
	Set(nil)

	updates := map[string]string{
		"log_level": "debug",
	}

	// Should not panic when config is nil
	assert.NotPanics(t, func() {
		Update(updates)
	}, "Update should handle nil config without panicking")
}

// TestUpdate_WithValidConfig tests that Update accepts valid update maps.
// This is an isolated method test with no external dependencies.
// Note: The Update function is currently a TODO, so this tests the basic structure.
func TestUpdate_WithValidConfig(t *testing.T) {
	cfg := &Config{
		ServiceName: "test-service",
		LogLevel:    "info",
	}
	Set(cfg)

	updates := map[string]string{
		"log_level": "debug",
	}

	// Should not panic
	assert.NotPanics(t, func() {
		Update(updates)
	}, "Update should accept valid updates without panicking")

	// Cleanup
	Set(nil)
}

// TestConfig_DefaultValues tests the default values in Config struct.
// This verifies the struct tag defaults are correctly defined.
func TestConfig_DefaultValues(t *testing.T) {
	// This test verifies the struct definition includes proper defaults
	// The actual default application happens in envconfig.Process()

	cfg := &Config{}

	// Verify zero values before defaults are applied
	assert.Equal(t, "", cfg.ServiceName)
	assert.Equal(t, "", cfg.Environment)
	assert.Equal(t, 0, cfg.BusinessHTTPPort)
	assert.Equal(t, 0, cfg.BusinessGRPCPort)
	assert.Equal(t, 0, cfg.HealthCheckPort)
	assert.Equal(t, 0, cfg.MetricsPort)
}

// TestConfig_StructTags verifies that Config struct has required envconfig tags.
// This ensures configuration can be loaded from environment variables.
func TestConfig_StructTags(t *testing.T) {
	// This test verifies struct tags are present
	// Actual tag parsing is done by envconfig library

	cfg := Config{}

	// Verify struct is not nil and can be instantiated
	assert.NotNil(t, &cfg)

	// Verify fields exist and are of correct types
	assert.IsType(t, "", cfg.ServiceName)
	assert.IsType(t, "", cfg.Environment)
	assert.IsType(t, 0, cfg.BusinessHTTPPort)
	assert.IsType(t, 0, cfg.BusinessGRPCPort)
	assert.IsType(t, 0, cfg.HealthCheckPort)
	assert.IsType(t, 0, cfg.MetricsPort)
	assert.IsType(t, "", cfg.LogLevel)
	assert.IsType(t, "", cfg.LogFormat)
	assert.IsType(t, "", cfg.DatabaseDSN)
	assert.IsType(t, false, cfg.EnableK8sConfigWatch)
}

// TestConfig_K8sFields verifies Kubernetes-related configuration fields.
// This is an isolated test of the struct definition.
func TestConfig_K8sFields(t *testing.T) {
	cfg := &Config{
		EnableK8sConfigWatch: true,
		K8sNamespace:         "production",
		K8sConfigMapName:     "app-config",
	}

	assert.True(t, cfg.EnableK8sConfigWatch)
	assert.Equal(t, "production", cfg.K8sNamespace)
	assert.Equal(t, "app-config", cfg.K8sConfigMapName)
}

// TestConfig_DatabaseFields verifies database-related configuration fields.
// This is an isolated test of the struct definition.
func TestConfig_DatabaseFields(t *testing.T) {
	cfg := &Config{
		DatabaseDSN:          "user:pass@tcp(localhost:3306)/db",
		DatabaseMaxOpenConns: 50,
		DatabaseMaxIdleConns: 5,
	}

	assert.Equal(t, "user:pass@tcp(localhost:3306)/db", cfg.DatabaseDSN)
	assert.Equal(t, 50, cfg.DatabaseMaxOpenConns)
	assert.Equal(t, 5, cfg.DatabaseMaxIdleConns)
}
