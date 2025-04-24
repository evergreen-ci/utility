package ttlcache

import (
	"testing"
)

func TestTTLInMemoryCache(t *testing.T) {
	testCache(t, func() Cache[int] {
		return NewInMemory[int]()
	})
}
