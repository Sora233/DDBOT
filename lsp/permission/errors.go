package permission

import (
	"errors"
)

var (
	ErrPermissionExist    = errors.New("already exist")
	ErrPermissionNotExist = errors.New("not exist")
	ErrPermissionDenied   = errors.New("permission denied")
	ErrDisabled           = errors.New("disabled")
	ErrGlobalDisabled     = errors.New("global disabled")
	ErrGlobalSilenced     = errors.New("global silenced")
)

func IsPermissionError(err error) bool {
	if errors.Is(err, ErrDisabled) || errors.Is(err, ErrPermissionDenied) ||
		errors.Is(err, ErrPermissionExist) || errors.Is(err, ErrPermissionNotExist) {
		return true
	}
	return false
}
