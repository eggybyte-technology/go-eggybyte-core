package cache

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCacheService(t *testing.T) {
	// Skip if no Memcached server available
	cfg := DefaultConfig()
	cfg.Servers = []string{"localhost:11211"}

	client, err := Connect(cfg)
	if err != nil {
		t.Skip("Memcached not available:", err)
	}

	service := NewCacheService(client)
	ctx := context.Background()

	t.Run("Set and Get", func(t *testing.T) {
		key := "test-key"
		value := []byte("test-value")

		err := service.Set(ctx, key, value, time.Minute)
		require.NoError(t, err)

		retrieved, err := service.Get(ctx, key)
		require.NoError(t, err)
		assert.Equal(t, value, retrieved)
	})

	t.Run("Get non-existent key", func(t *testing.T) {
		value, err := service.Get(ctx, "non-existent")
		require.NoError(t, err)
		assert.Nil(t, value)
	})

	t.Run("Delete", func(t *testing.T) {
		key := "delete-key"
		value := []byte("delete-value")

		err := service.Set(ctx, key, value, time.Minute)
		require.NoError(t, err)

		err = service.Delete(ctx, key)
		require.NoError(t, err)

		retrieved, err := service.Get(ctx, key)
		require.NoError(t, err)
		assert.Nil(t, retrieved)
	})

	t.Run("Exists", func(t *testing.T) {
		key := "exists-key"
		value := []byte("exists-value")

		exists, err := service.Exists(ctx, key)
		require.NoError(t, err)
		assert.False(t, exists)

		err = service.Set(ctx, key, value, time.Minute)
		require.NoError(t, err)

		exists, err = service.Exists(ctx, key)
		require.NoError(t, err)
		assert.True(t, exists)
	})
}
