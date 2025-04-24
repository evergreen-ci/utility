package ttlcache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testCache(t *testing.T, cacheFunc func() Cache[*int]) {
	t.Run("InvalidKey", func(t *testing.T) {
		cache := cacheFunc()

		id, ok := cache.Get(t.Context(), "key", time.Minute)
		assert.False(t, ok)
		assert.Zero(t, id)

		val := 22
		cache.Put(t.Context(), "key", &val, time.Now().Add(time.Second))
	})

	t.Run("ValidKey", func(t *testing.T) {
		cache := cacheFunc()

		val := 22
		cache.Put(t.Context(), "key", &val, time.Now().Add(time.Second))

		t.Run("BeforeExpiration", func(t *testing.T) {
			cachedVal, ok := cache.Get(t.Context(), "key", time.Millisecond)
			require.True(t, ok)
			require.NotNil(t, cachedVal)
			assert.Equal(t, val, *cachedVal)
		})

		t.Run("AfterExpiration", func(t *testing.T) {
			id, ok := cache.Get(t.Context(), "key", time.Minute)
			assert.False(t, ok)
			assert.Zero(t, id)
		})
	})

	t.Run("ReplaceKey", func(t *testing.T) {
		cache := cacheFunc()

		val := 22
		cache.Put(t.Context(), "key", &val, time.Now().Add(time.Second))

		// Overwrite the key with a new value and a longer expiration time.
		secondVal := 23
		cache.Put(t.Context(), "key", &secondVal, time.Now().Add(time.Hour))

		t.Run("BeforeExpiration", func(t *testing.T) {
			cachedVal, ok := cache.Get(t.Context(), "key", time.Minute)
			require.True(t, ok)
			require.NotNil(t, cachedVal)
			assert.Equal(t, secondVal, *cachedVal)
		})

		t.Run("AfterExpiration", func(t *testing.T) {
			cachedVal, ok := cache.Get(t.Context(), "key", time.Hour*2)
			assert.False(t, ok)
			assert.Zero(t, cachedVal)
		})
	})
}
