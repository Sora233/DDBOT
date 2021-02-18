package buntdb

import "errors"

var (
	ErrKeyExist       = errors.New("key exist")
	ErrNotInitialized = errors.New("not initialized")
)
