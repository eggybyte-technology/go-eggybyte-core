package cache

import (
	"context"
	"fmt"

	"github.com/eggybyte-technology/go-eggybyte-core/pkg/log"
	"github.com/eggybyte-technology/go-eggybyte-core/pkg/service"
)

// CacheInitializer implements service.Initializer interface for cache setup.
type CacheInitializer struct {
	config *Config
}

// NewCacheInitializer creates a new cache initializer.
func NewCacheInitializer(cfg *Config) *CacheInitializer {
	return &CacheInitializer{
		config: cfg,
	}
}

// Init establishes cache connection and verifies connectivity.
func (c *CacheInitializer) Init(ctx context.Context) error {
	log.Info("Initializing Memcached connection")

	if c.config == nil {
		return fmt.Errorf("cache config is nil")
	}

	if len(c.config.Servers) == 0 {
		return fmt.Errorf("at least one Memcached server is required")
	}

	// Establish cache connection
	_, err := Connect(c.config)
	if err != nil {
		return fmt.Errorf("failed to connect to cache: %w", err)
	}

	log.Info("Cache initialization completed successfully")
	return nil
}

// Verify that CacheInitializer implements service.Initializer interface.
var _ service.Initializer = (*CacheInitializer)(nil)
