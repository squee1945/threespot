package storage

import "errors"

const retries = 3

var (
	ErrNotFound             = errors.New("Not found")
	ErrNotUnique            = errors.New("ID is not unique")
	ErrPlayerPositionFilled = errors.New("Player position is already filled")
	ErrPlayerAlreadyAdded   = errors.New("Player is already added")
)
