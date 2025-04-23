package ttlcache

import (
	"testing"
)

func TestWeakInMemory(t *testing.T) {
	testCache(t, func() Cache[int] {
		return convertPointerCacheToCache(NewWeakInMemory[int]())
	})
}
