package utility

import (
	"context"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestIsContextError(t *testing.T) {
	t.Run("ContextCanceledReturnsTrue", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		assert.True(t, IsContextError(ctx.Err()))
	})
	t.Run("ContextDeadlineExceededReturnsTrue", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
		defer cancel()
		time.Sleep(10 * time.Millisecond)
		assert.True(t, IsContextError(ctx.Err()))
	})
	t.Run("ContextWithoutErrorReturnsFalse", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		assert.False(t, IsContextError(ctx.Err()))
	})
	t.Run("NonContextErrorReturnsFalse", func(t *testing.T) {
		assert.False(t, IsContextError(errors.New("custom error")))
	})
	t.Run("WrappedContextErrorReturnsFalse", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		wrappedErr := errors.Wrap(ctx.Err(), "wrapped error")
		assert.False(t, IsContextError(wrappedErr))
	})
	t.Run("UnwrappedContextErrorReturnsTrue", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		wrappedErr := errors.Wrap(ctx.Err(), "wrapped error")
		assert.True(t, IsContextError(errors.Cause(wrappedErr)))
	})
}
