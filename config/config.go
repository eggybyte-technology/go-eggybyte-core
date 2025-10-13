package config

import (
	"sync"
)

// Config holds common configuration for EggyByte services.
// This structure provides standard fields that most microservices need,
// including service identification, networking, database, and observability.
//
// Services can embed this struct in their own config and add custom fields.
// All fields support environment variable loading via struct tags.
//
// Thread Safety: After initialization, Config should be treated as read-only.
// Use the global config accessor methods which provide thread-safe access.
//
// Example usage:
//
//	type MyServiceConfig struct {
//	    config.Config
//	    CustomField string `envconfig:"CUSTOM_FIELD"`
//	}
type Config struct {
	// ServiceName is the unique identifier for this service instance.
	// Used in logging, metrics labels, and service discovery.
	ServiceName string `envconfig:"SERVICE_NAME" required:"true"`

	// Environment specifies the deployment environment (e.g., dev, staging, production).
	// Used for environment-specific configuration and feature flags.
	Environment string `envconfig:"ENVIRONMENT" default:"development"`

	// Port is the HTTP server listening port for business APIs.
	Port int `envconfig:"PORT" default:"8080"`

	// MetricsPort is the HTTP server port for Prometheus metrics and health checks.
	// Separated from business port for security and monitoring isolation.
	MetricsPort int `envconfig:"METRICS_PORT" default:"9090"`

	// LogLevel controls the verbosity of logging output.
	// Valid values: debug, info, warn, error, fatal
	LogLevel string `envconfig:"LOG_LEVEL" default:"info"`

	// LogFormat specifies the log output format.
	// Valid values: json, console
	LogFormat string `envconfig:"LOG_FORMAT" default:"json"`

	// DatabaseDSN is the Data Source Name for database connection.
	// Format: "username:password@tcp(host:port)/dbname?charset=utf8mb4&parseTime=True"
	// Empty value means database is not used by this service.
	DatabaseDSN string `envconfig:"DATABASE_DSN"`

	// DatabaseMaxOpenConns sets the maximum number of open database connections.
	DatabaseMaxOpenConns int `envconfig:"DATABASE_MAX_OPEN_CONNS" default:"100"`

	// DatabaseMaxIdleConns sets the maximum number of idle database connections.
	DatabaseMaxIdleConns int `envconfig:"DATABASE_MAX_IDLE_CONNS" default:"10"`

	// EnableK8sConfigWatch enables Kubernetes ConfigMap watching for dynamic config updates.
	// Requires proper RBAC permissions to watch ConfigMaps in the specified namespace.
	EnableK8sConfigWatch bool `envconfig:"ENABLE_K8S_CONFIG_WATCH" default:"false"`

	// K8sNamespace is the Kubernetes namespace to watch for ConfigMaps.
	K8sNamespace string `envconfig:"K8S_NAMESPACE" default:"default"`

	// K8sConfigMapName is the name of the ConfigMap to watch for config updates.
	K8sConfigMapName string `envconfig:"K8S_CONFIGMAP_NAME"`
}

var (
	// globalConfig holds the singleton configuration instance.
	globalConfig *Config

	// configMutex protects concurrent access to globalConfig.
	configMutex sync.RWMutex
)

// Get returns the current global configuration.
// This method is thread-safe and can be called from multiple goroutines.
//
// Returns:
//   - *Config: The current configuration, or nil if not initialized.
//
// Example:
//
//	cfg := config.Get()
//	if cfg != nil {
//	    log.Printf("Service: %s, Port: %d", cfg.ServiceName, cfg.Port)
//	}
func Get() *Config {
	configMutex.RLock()
	defer configMutex.RUnlock()
	return globalConfig
}

// Set updates the global configuration with a new instance.
// This method is thread-safe and typically called during service initialization.
//
// Parameters:
//   - cfg: The new configuration to set globally.
//
// Example:
//
//	newConfig := &config.Config{ServiceName: "user-service"}
//	config.Set(newConfig)
func Set(cfg *Config) {
	configMutex.Lock()
	defer configMutex.Unlock()
	globalConfig = cfg
}

// Update applies partial configuration updates from a map.
// This method is used by Kubernetes ConfigMap watchers to dynamically
// update configuration without restarting the service.
//
// Parameters:
//   - updates: Map of configuration keys to new values.
//
// Thread Safety: This method is thread-safe for concurrent updates.
//
// Note: Only string-based configuration fields can be updated this way.
// Complex types require service restart.
func Update(updates map[string]string) {
	configMutex.Lock()
	defer configMutex.Unlock()

	if globalConfig == nil {
		return
	}

	// Apply updates to mutable fields
	// TODO: Implement field-by-field update logic
	// This will be expanded when K8s watcher is implemented
}
