package utility

import "errors"

// MatchesError returns whether err matches type T.
func MatchesError[T error](err error) bool {
	var errType T
	return errors.As(err, &errType)
}
