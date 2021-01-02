package zhima

import (
	"github.com/Sora233/Sora233-MiraiGo/proxy_pool"
	zhima_proxy_pool "github.com/Sora233/zhima-proxy-pool"
)

type ZhimaWrapper struct {
	pool *zhima_proxy_pool.ZhimaProxyPool
}

func (z *ZhimaWrapper) Get() (proxy_pool.IProxy, error) {
	return z.pool.Get()
}

func (z *ZhimaWrapper) Delete(iproxy proxy_pool.IProxy) bool {
	if proxy, ok := iproxy.(*zhima_proxy_pool.Proxy); ok {
		return z.pool.Delete(proxy)
	} else {
		return false
	}
}

func (z *ZhimaWrapper) Stop() error {
	return z.pool.Stop()
}

func NewZhimaWrapper(pool *zhima_proxy_pool.ZhimaProxyPool) *ZhimaWrapper {
	return &ZhimaWrapper{pool: pool}
}
