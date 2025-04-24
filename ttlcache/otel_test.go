package ttlcache

import (
	"testing"
)

func TestTTLOtelCache(t *testing.T) {
	testCache(t, func() Cache[*int] {
		return WithOtel(NewInMemory[*int](), "test")
	})
}
