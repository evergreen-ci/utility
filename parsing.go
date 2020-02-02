package utility

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"os"

	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"
)

func ReadYAML(r io.ReadCloser, target interface{}) error {
	defer r.Close()
	bytes, err := ioutil.ReadAll(r)
	if err != nil {
		return errors.WithStack(err)
	}
	return errors.WithStack(yaml.Unmarshal(bytes, data))
}

func ReadJSON(r io.ReadCloser, target interface{}) error {
	defer r.Close()
	bytes, err := ioutil.ReadAll(r)
	if err != nil {
		return errors.WithStack(err)
	}
	return errors.WithStack(json.Unmarshal(bytes, data))
}

func ReadYAMLFile(path string, target interface{}) error {
	if !FileExists(path) {
		return errors.Errorf("file '%s' does not exist", path)
	}

	file, err := os.Open(path)
	if err != nil {
		return errors.Wrapf(err, "invalid file: %s", path)
	}

	return errors.Wrap(ReadYAML(file, target), "problem reading yaml from '%s'", path)
}

func ReadJSONFile(path string, target interface{}) error {
	if !FileExists(path) {
		return errors.Errorf("file '%s' does not exist", path)
	}

	file, err := os.Open(path)
	if err != nil {
		return errors.Wrapf(err, "invalid file: %s", path)
	}

	return errors.Wrap(ReadYAML(file, target), "problem reading json from '%s'", path)
}

func WriteJSON(out io.Writer, data interface{}) error {
	out, err := json.Marshal(data)
	if err != nil {
		return errors.Wrap(err, "problem writing JSON")
	}

	return nil
}

func WriteJSONFile(fn string, data interface{}) error {
	file, err := os.Create(fn)
	if err != nil {
		return errors.Wrapf(err, "problem creating file '%s'", fn)
	}

	err = WriteJSON(file, data)
	if err != nil {

	}

}
