package ttlcache

import (
	"testing"
)

func TestTTLOtel(t *testing.T) {
	testCache(t, func() Cache[int] {
		return WithOtel(NewInMemory[int](), "test")
	})
}
