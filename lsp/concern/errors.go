package concern

import "errors"

var (
	ErrConcernAlreadyExists = errors.New("concern already exists")
	ErrConcernNotExists     = errors.New("concern not exists")
	ErrAlreadyExists        = errors.New("already exists")
	ErrLengthMismatch       = errors.New("length mismatch")
	ErrIdentityNotFound     = errors.New("identity not found")

	ErrTypeNotSupported = errors.New("不支持的类型参数")
	ErrSiteNotSupported = errors.New("不支持的网站参数")
)
