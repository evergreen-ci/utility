package utility

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConvertInt32PtrToIntPtr(t *testing.T) {
	t.Run("NilInput", func(t *testing.T) {
		assert.Nil(t, ConvertInt32PtrToIntPtr(nil))
	})
	t.Run("NonNilInput", func(t *testing.T) {
		val := int32(42)
		result := ConvertInt32PtrToIntPtr(&val)
		assert.NotNil(t, result)
		assert.Equal(t, 42, *result)
	})
	t.Run("ZeroValue", func(t *testing.T) {
		val := int32(0)
		result := ConvertInt32PtrToIntPtr(&val)
		assert.NotNil(t, result)
		assert.Equal(t, 0, *result)
	})
	t.Run("NegativeValue", func(t *testing.T) {
		val := int32(-10)
		result := ConvertInt32PtrToIntPtr(&val)
		assert.NotNil(t, result)
		assert.Equal(t, -10, *result)
	})
}
