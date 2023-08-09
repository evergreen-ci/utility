package utility

import (
	"io/fs"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMatchesError(t *testing.T) {
	assert.True(t, MatchesError[*fs.PathError](&os.PathError{}))
	assert.False(t, MatchesError[*fs.PathError](os.ErrNotExist))
}
