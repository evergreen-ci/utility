package ttlcache

import (
	"testing"
)

func TestTTLWeakInMemoryCache(t *testing.T) {
	testPointerCache(t, func() PointerCache[int] {
		return NewWeakInMemory[int]()
	})
}
