package ttlcache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testCache(t *testing.T, cacheFunc func() Cache[int]) {
	t.Run("ImplementsTTLCache", func(t *testing.T) {
		require.Implements(t, (*Cache[int])(nil), cacheFunc())
	})

	t.Run("InvalidKey", func(t *testing.T) {
		cache := cacheFunc()

		id, ok := cache.Get(t.Context(), "key", time.Minute)
		assert.False(t, ok)
		assert.Zero(t, id)

		cache.Put(t.Context(), "key", 22, time.Now().Add(time.Second))
	})

	t.Run("ValidKey", func(t *testing.T) {
		cache := cacheFunc()

		cache.Put(t.Context(), "key", 22, time.Now().Add(time.Second))

		t.Run("BeforeExpiration", func(t *testing.T) {
			id, ok := cache.Get(t.Context(), "key", time.Millisecond)
			assert.True(t, ok)
			assert.Equal(t, 22, id)
		})

		t.Run("AfterExpiration", func(t *testing.T) {
			id, ok := cache.Get(t.Context(), "key", time.Minute)
			assert.False(t, ok)
			assert.Zero(t, id)
		})
	})

	t.Run("ReplaceKey", func(t *testing.T) {
		cache := cacheFunc()

		cache.Put(t.Context(), "key", 22, time.Now().Add(time.Second))

		// Overwrite the key with a new value and a longer expiration time.
		cache.Put(t.Context(), "key", 23, time.Now().Add(time.Hour))

		t.Run("BeforeExpiration", func(t *testing.T) {
			id, ok := cache.Get(t.Context(), "key", time.Minute)
			assert.True(t, ok)
			assert.Equal(t, 23, id)
		})

		t.Run("AfterExpiration", func(t *testing.T) {
			id, ok := cache.Get(t.Context(), "key", time.Hour*2)
			assert.False(t, ok)
			assert.Zero(t, id)
		})
	})
}
