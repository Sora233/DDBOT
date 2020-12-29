package command

import "errors"

var (
	ErrInvalidPrimaryArg     = errors.New("invalid primary arg")
	ErrPrimaryArgExist       = errors.New("primary arg already exists")
	ErrCommandPrefixConflict = errors.New("primary arg conflict with command prefix")
)
