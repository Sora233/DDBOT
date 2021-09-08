package concern

import "errors"

var (
	ErrConcernAlreadyExists = errors.New("concern already exists")
	ErrConcernNotExists     = errors.New("concern not exists")
	ErrAlreadyExists        = errors.New("already exists")
	ErrLengthMismatch       = errors.New("length mismatch")
	ErrIdentityNotFound     = errors.New("identity not found")
)
