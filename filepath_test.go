package utility

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConsistentFilepath(t *testing.T) {
	t.Run("IsEquivalentToFilepathJoinForUnixStylePaths", func(t *testing.T) {
		parts := []string{"foo", "bar", "bat"}
		expected := "foo/bar/bat"
		assert.Equal(t, expected, filepath.Join(parts...))
		assert.Equal(t, expected, ConsistentFilepath(parts...))
	})
	t.Run("IsEquivalentToFilepathJoinForUnixStylePathsWithSpaces", func(t *testing.T) {
		parts := []string{"foo bar", "bat"}
		expected := "foo bar/bat"
		assert.Equal(t, expected, filepath.Join(parts...))
		assert.Equal(t, expected, ConsistentFilepath(parts...))
	})
	t.Run("ConvertsWindowsStylePathsToUnixStylePaths", func(t *testing.T) {
		windowsPath := "foo\\bar\\bat"
		expected := "foo/bar/bat"
		assert.Equal(t, expected, ConsistentFilepath(windowsPath))
	})
}
