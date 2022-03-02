package utility

import "context"

// IsContextError returns true if the given error is due to a context-related
// error.
func IsContextError(err error) bool {
	return err == context.DeadlineExceeded || err == context.Canceled
}
