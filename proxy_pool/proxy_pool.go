package proxy_pool

import "errors"

var ErrNil = errors.New("<nil>")

type IProxyPool interface {
	Get() (IProxy, error)
	Delete(IProxy) bool
}

type IProxy interface {
	ProxyString() string
}

var proxyPool IProxyPool

func Init(proxy IProxyPool) {
	proxyPool = proxy
}

func Get() (IProxy, error) {
	if proxyPool == nil {
		return nil, ErrNil
	}
	return proxyPool.Get()
}
func Delete(proxy IProxy) bool {
	return proxyPool.Delete(proxy)
}
