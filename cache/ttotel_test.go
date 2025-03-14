package cache

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTTLOtel(t *testing.T) {
	t.Run("ImplementsTTLCache", func(t *testing.T) {
		require.Implements(t, (*TTLCache[int])(nil), WithOtel[int](nil, "Test"))
	})
}
