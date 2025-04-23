package ttlcache

import (
	"testing"
)

func TestTTLOtelCache(t *testing.T) {
	testCache(t, func() Cache[int] {
		return WithOtel(NewInMemory[int](), "test")
	})
}

func TestTTLPointerOtelCache(t *testing.T) {
	testCache(t, func() Cache[int] {
		return convertPointerCacheToCache(WithPointerOtel(NewWeakInMemory[int](), "test"))
	})
}
