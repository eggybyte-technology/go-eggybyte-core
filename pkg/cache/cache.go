package cache

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/eggybyte-technology/go-eggybyte-core/pkg/log"
)

// Config holds cache configuration parameters.
type Config struct {
	// Servers is a list of Memcached server addresses.
	// Format: ["host1:port1", "host2:port2"]
	Servers []string

	// MaxIdleConns sets the maximum number of idle connections per server.
	MaxIdleConns int

	// Timeout sets the connection timeout.
	Timeout time.Duration

	// ConnectTimeout sets the initial connection timeout.
	ConnectTimeout time.Duration
}

// DefaultConfig returns cache configuration with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		MaxIdleConns:   10,
		Timeout:        5 * time.Second,
		ConnectTimeout: 5 * time.Second,
	}
}

var (
	// globalClient holds the singleton Memcached client.
	globalClient *memcache.Client
	cacheMutex   sync.RWMutex
)

// Connect establishes a Memcached connection using the provided configuration.
func Connect(cfg *Config) (*memcache.Client, error) {
	if len(cfg.Servers) == 0 {
		return nil, fmt.Errorf("at least one Memcached server is required")
	}

	client := memcache.New(cfg.Servers...)
	client.Timeout = cfg.Timeout
	client.MaxIdleConns = cfg.MaxIdleConns

	// Test connection
	if err := client.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping Memcached: %w", err)
	}

	// Set global client
	SetClient(client)

	log.Info("Memcached connection established",
		log.Field{Key: "servers", Value: cfg.Servers},
		log.Field{Key: "max_idle_conns", Value: cfg.MaxIdleConns})

	return client, nil
}

// GetClient returns the global Memcached client.
func GetClient() *memcache.Client {
	cacheMutex.RLock()
	defer cacheMutex.RUnlock()
	return globalClient
}

// SetClient updates the global Memcached client.
func SetClient(client *memcache.Client) {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	globalClient = client
}

// Close closes the Memcached connection.
func Close() error {
	// Memcached client doesn't have explicit close method
	// Connection pooling is handled internally
	log.Info("Memcached connection closed")
	return nil
}

// CacheService provides high-level cache operations.
type CacheService struct {
	client *memcache.Client
}

// NewCacheService creates a new cache service.
func NewCacheService(client *memcache.Client) *CacheService {
	return &CacheService{client: client}
}

// Get retrieves a value from cache.
func (c *CacheService) Get(ctx context.Context, key string) ([]byte, error) {
	item, err := c.client.Get(key)
	if err != nil {
		if err == memcache.ErrCacheMiss {
			return nil, nil
		}
		return nil, err
	}
	return item.Value, nil
}

// Set stores a value in cache with expiration.
func (c *CacheService) Set(ctx context.Context, key string, value []byte, expiration time.Duration) error {
	item := &memcache.Item{
		Key:        key,
		Value:      value,
		Expiration: int32(expiration.Seconds()),
	}
	return c.client.Set(item)
}

// Delete removes a value from cache.
func (c *CacheService) Delete(ctx context.Context, key string) error {
	return c.client.Delete(key)
}

// Exists checks if a key exists in cache.
func (c *CacheService) Exists(ctx context.Context, key string) (bool, error) {
	_, err := c.client.Get(key)
	if err == memcache.ErrCacheMiss {
		return false, nil
	}
	return err == nil, err
}
