package ttlcache

import (
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTTLWeakInMemoryCache(t *testing.T) {
	testCache(t, func() Cache[*int] {
		return NewWeakInMemory[int]()
	})

	t.Run("ReleasesMemory", func(t *testing.T) {
		getItem := func() *string {
			i := "itemValue"
			return &i
		}

		cache := NewWeakInMemory[string]()

		itemVar := getItem()
		cache.Put(t.Context(), "id", itemVar, time.Now().Add(time.Hour))

		cachedItem, found := cache.Get(t.Context(), "id", time.Minute)
		require.True(t, found)
		require.NotNil(t, cachedItem)
		assert.Equal(t, itemVar, cachedItem)

		// There are still references to 'itemVar' in the test, so it should not be garbage collected.
		runtime.GC()
		cachedItem, found = cache.Get(t.Context(), "id", time.Minute)
		require.True(t, found)
		require.NotNil(t, cachedItem)
		assert.Equal(t, itemVar, cachedItem)

		// There are no references to 'itemVar' in the test, so it should be garbage collected.
		runtime.GC()
		cachedItem, found = cache.Get(t.Context(), "id", time.Minute)
		assert.False(t, found)
		assert.Nil(t, cachedItem)
	})
}
