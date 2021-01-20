package permission

import "errors"

var (
	ErrPermissionExist    = errors.New("already exist")
	ErrPermissionNotExist = errors.New("not exist")
)
