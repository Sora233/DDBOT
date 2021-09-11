package registry

import "errors"

var (
	ErrTypeNotSupported = errors.New("不支持的类型参数")
	ErrSiteNotSupported = errors.New("不支持的网站参数")
)
