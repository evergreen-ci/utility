package ttlcache

import (
	"testing"
)

func TestTLLInMemoryCache(t *testing.T) {
	testCache(t, func() Cache[int] {
		return NewInMemory[int]()
	})
}
