package concern_manager

import "errors"

var (
	ErrAlreadyExists  = errors.New("already exists")
	ErrLengthMismatch = errors.New("length mismatch")
)
