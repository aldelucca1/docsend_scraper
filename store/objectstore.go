package store

import "io"

// ObjectStore is an interface to an underlying object storage
type ObjectStore interface {

	// Check if an object exists at the given path
	Exists(path string) (bool, error)

	// Write an object to the given path
	Write(path string, reader io.Reader) error

	// Read an object from the given path
	Read(path string) (io.Reader, error)
}
