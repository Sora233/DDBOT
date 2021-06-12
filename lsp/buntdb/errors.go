package buntdb

import "errors"

var (
	ErrKeyExist       = errors.New("key exist")
	ErrNotInitialized = errors.New("not initialized")
	ErrRollback       = errors.New("rollback")
)

func IsRollback(e error) bool {
	return e != nil && e.Error() == ErrRollback.Error()
}
