package store

import "errors"

var (
	ErrDuplicateKey = errors.New("duplicate key")
	ErrNotFound     = errors.New("not found")
	ErrInternal     = errors.New("internal error")
)
