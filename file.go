package utility

import (
	"os"

	"github.com/mongodb/grip"
	"github.com/pkg/errors"
)

// FileExists provides a clearer interface for checking if a file
// exists.
func FileExists(path string) bool {
	if path == "" {
		return false
	}

	_, err := os.Stat(path)

	return !os.IsNotExist(err)
}

// WriteFile provides a clearer interface for writing string data to a
// file.
func WriteFile(path string, data string) error {
	file, err := os.Create(path)
	if err != nil {
		return errors.Wrapf(err, "problem creating file '%s'", path)
	}

	_, err = file.WriteString(data)
	if err != nil {
		grip.Warning(errors.Wrapf(file.Close(), "problem closing file '%s' after error", path))
		return errors.Wrapf(err, "problem writing data to file '%s'", path)
	}

	if err = file.Close(); err != nil {
		return errors.Wrapf(err, "problem closing file '%s' after successfully writing data", path)
	}

	return nil
}
