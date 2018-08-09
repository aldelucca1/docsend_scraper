package fs

import (
	"io"
	"os"
	"path"
)

// NewStore - Create a new filesystem backed object store
func NewStore(config *Config) *Store {
	m := new(Store)
	m.config = config
	return m
}

// Store is a Objectstore backed by the filesystem
type Store struct {
	config *Config
}

// Exists checks if an object exists at the given path
func (fs *Store) Exists(pathStr string) (bool, error) {
	out := path.Join(fs.config.outputPath, pathStr)
	if _, err := os.Stat(out); os.IsNotExist(err) {
		return false, err
	}
	return true, nil
}

// Write an object to the given path
func (fs *Store) Write(pathStr string, reader io.Reader) error {

	out := path.Join(fs.config.outputPath, pathStr)

	_, err := os.Stat(path.Dir(out))
	if os.IsNotExist(err) {
		err = os.MkdirAll(path.Dir(out), os.ModePerm)
	}
	if err != nil {
		return err
	}

	to, err := os.OpenFile(out, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		return err
	}
	defer to.Close()

	_, err = io.Copy(to, reader)
	return err
}

// Read an object from the given path
func (fs *Store) Read(pathStr string) (io.Reader, error) {
	out := path.Join(fs.config.outputPath, pathStr)
	return os.Open(out)
}
