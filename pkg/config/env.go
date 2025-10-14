// Package config provides unified configuration management for EggyByte services.
package config

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

// ReadFromEnv loads configuration from environment variables into the provided struct.
// This function uses struct tags to map environment variable names to struct fields.
//
// The struct must use `envconfig` tags to specify variable names and options:
//   - `envconfig:"VAR_NAME"` - Maps to environment variable VAR_NAME
//   - `required:"true"` - Makes the field mandatory
//   - `default:"value"` - Sets default value if env var not set
//
// Parameters:
//   - cfg: Pointer to a struct that will receive configuration values.
//     Must have proper envconfig tags on fields.
//
// Returns:
//   - error: Returns error if required fields are missing or type conversion fails.
//
// Example:
//
//	type MyConfig struct {
//	    ServiceName string `envconfig:"SERVICE_NAME" required:"true"`
//	    Port        int    `envconfig:"PORT" default:"8080"`
//	}
//
//	var cfg MyConfig
//	if err := config.ReadFromEnv(&cfg); err != nil {
//	    log.Fatal("Failed to load config:", err)
//	}
func ReadFromEnv(cfg interface{}) error {
	if err := envconfig.Process("", cfg); err != nil {
		return fmt.Errorf("failed to process environment configuration: %w", err)
	}
	return nil
}

// MustReadFromEnv loads configuration from environment variables and panics on error.
// This is a convenience wrapper around ReadFromEnv for use during service initialization
// where configuration errors should halt startup.
//
// Parameters:
//   - cfg: Pointer to a struct that will receive configuration values.
//
// Panics:
//   - If any required environment variables are missing
//   - If type conversion fails for any field
//
// Example:
//
//	var cfg config.Config
//	config.MustReadFromEnv(&cfg)
//	// Service continues with valid configuration or panics
func MustReadFromEnv(cfg interface{}) {
	if err := ReadFromEnv(cfg); err != nil {
		panic(fmt.Sprintf("fatal configuration error: %v", err))
	}
}

// ValidateConfig performs additional validation on loaded configuration.
// This checks for logical consistency beyond what struct tags can enforce.
//
// Parameters:
//   - cfg: Configuration instance to validate
//
// Returns:
//   - error: Returns error if validation fails with descriptive message
//
// Validation rules:
//   - ServiceName must not be empty
//   - Port must be in valid range (1-65535)
//   - MetricsPort must be different from Port
//   - LogLevel must be one of: debug, info, warn, error, fatal
//   - If K8s watching enabled, namespace and configmap name required
func ValidateConfig(cfg *Config) error {
	if err := validateServiceName(cfg.ServiceName); err != nil {
		return err
	}

	if err := validatePorts(cfg); err != nil {
		return err
	}

	if err := validateLogLevel(cfg.LogLevel); err != nil {
		return err
	}

	if err := validateK8sConfig(cfg); err != nil {
		return err
	}

	return nil
}

// validateServiceName validates the service name
func validateServiceName(serviceName string) error {
	if serviceName == "" {
		return fmt.Errorf("service name cannot be empty")
	}
	return nil
}

// validatePorts validates all port configurations
func validatePorts(cfg *Config) error {
	ports := map[string]int{
		"business HTTP": cfg.BusinessHTTPPort,
		"business gRPC": cfg.BusinessGRPCPort,
		"health check":  cfg.HealthCheckPort,
		"metrics":       cfg.MetricsPort,
	}

	// Validate port ranges
	for name, port := range ports {
		if port < 1 || port > 65535 {
			return fmt.Errorf("%s port must be between 1 and 65535, got: %d", name, port)
		}
	}

	// Validate port uniqueness
	portMap := make(map[int]string)
	for name, port := range ports {
		if existing, exists := portMap[port]; exists {
			return fmt.Errorf("%s and %s ports cannot be the same: %d", name, existing, port)
		}
		portMap[port] = name
	}

	return nil
}

// validateLogLevel validates the log level
func validateLogLevel(logLevel string) error {
	validLogLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
		"fatal": true,
	}
	if !validLogLevels[logLevel] {
		return fmt.Errorf("invalid log level: %s (must be debug, info, warn, error, or fatal)", logLevel)
	}
	return nil
}

// validateK8sConfig validates Kubernetes configuration
func validateK8sConfig(cfg *Config) error {
	if !cfg.EnableK8sConfigWatch {
		return nil
	}

	if cfg.K8sNamespace == "" {
		return fmt.Errorf("kubernetes namespace required when config watch enabled")
	}

	if cfg.K8sConfigMapName == "" {
		return fmt.Errorf("kubernetes configmap name required when config watch enabled")
	}

	return nil
}
