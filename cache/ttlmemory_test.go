package cache_test

import (
	"testing"
	"time"

	"github.com/evergreen-ci/utility/cache"
	"github.com/stretchr/testify/assert"
)

func TestTLLInMemoryCache(t *testing.T) {
	t.Run("InvalidKey", func(t *testing.T) {
		cache := cache.NewTTLInMemory[int]()

		id, ok := cache.Get(t.Context(), "key", time.Minute)
		assert.False(t, ok)
		assert.Zero(t, id)

		cache.Put(t.Context(), "key", 22, time.Now().Add(time.Second))
	})

	t.Run("ValidKey", func(t *testing.T) {
		cache := cache.NewTTLInMemory[int]()

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
		cache := cache.NewTTLInMemory[int]()

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
