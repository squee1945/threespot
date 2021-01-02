package storage

import "errors"

const retries = 3

var (
	ErrNotFound  = errors.New("Not found.")
	ErrNotUnique = errors.New("ID is not unique.")
)
