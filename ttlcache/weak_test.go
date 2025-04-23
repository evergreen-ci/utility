package ttlcache

import (
	"testing"
)

func TestTTLWeakInMemoryCache(t *testing.T) {
	testCache(t, func() Cache[int] {
		return convertPointerCacheToCache(NewWeakInMemory[int]())
	})
}
