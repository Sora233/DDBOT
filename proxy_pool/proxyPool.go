package proxy_pool

import "errors"

var ErrNil = errors.New("<nil>")

type Prefer int64

const (
	PreferAny Prefer = 1 << iota
	PreferMainland
	PreferOversea
	PreferNone
)

type IProxyPool interface {
	Get(Prefer) (IProxy, error)
	Delete(string) bool
	Stop() error
}

type IProxy interface {
	ProxyString() string
}

var proxyPool IProxyPool

func Init(proxy IProxyPool) {
	proxyPool = proxy
}

func Get(prefer Prefer) (IProxy, error) {
	if proxyPool == nil {
		return nil, ErrNil
	}
	return proxyPool.Get(prefer)
}
func Delete(proxy string) bool {
	if proxyPool == nil {
		return false
	}
	return proxyPool.Delete(proxy)
}
func Stop() error {
	if proxyPool == nil {
		return ErrNil
	}
	return proxyPool.Stop()
}
