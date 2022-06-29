package buntdb

import (
	"errors"
	"github.com/tidwall/buntdb"
)

var (
	ErrKeyExist       = errors.New("key exist")
	ErrNotInitialized = errors.New("not initialized")
	ErrRollback       = errors.New("rollback")
	ErrLockNotHold    = errors.New("lock not hold")
)

func IsRollback(e error) bool {
	return e != nil && e.Error() == ErrRollback.Error()
}

func IsNotFound(e error) bool {
	return e == buntdb.ErrNotFound
}
