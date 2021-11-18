package concern

import "errors"

var (
	ErrAlreadyExists  = errors.New("already exists")
	ErrLengthMismatch = errors.New("length mismatch")

	ErrTypeNotSupported   = errors.New("不支持的类型参数")
	ErrSiteNotSupported   = errors.New("不支持的网站参数")
	ErrConfigNotSupported = errors.New("不支持的配置")
)
