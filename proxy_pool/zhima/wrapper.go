package zhima

import (
	"github.com/Sora233/Sora233-MiraiGo/proxy_pool"
	zhima_proxy_pool "github.com/Sora233/zhima-proxy-pool"
	"sync"
)

type Wrapper struct {
	pool        *zhima_proxy_pool.ZhimaProxyPool
	deleteLimit int
	deleteCount map[string]int
	mutex       *sync.RWMutex
}

func (z *Wrapper) Get() (proxy_pool.IProxy, error) {
	z.mutex.RLock()
	defer z.mutex.RUnlock()
	return z.pool.Get()
}

func (z *Wrapper) Delete(iproxy proxy_pool.IProxy) bool {
	z.mutex.Lock()
	defer z.mutex.Unlock()

	var result = false

	if _, found := z.deleteCount[iproxy.ProxyString()]; !found {
		z.deleteCount[iproxy.ProxyString()] = 1
	} else {
		z.deleteCount[iproxy.ProxyString()] += 1
	}

	if z.deleteCount[iproxy.ProxyString()] == z.deleteLimit {
		if proxy, ok := iproxy.(*zhima_proxy_pool.Proxy); ok {
			result = z.pool.Delete(proxy)
		}
		delete(z.deleteCount, iproxy.ProxyString())
	}
	return result
}

func (z *Wrapper) Stop() error {
	z.mutex.Lock()
	defer z.mutex.Unlock()
	return z.pool.Stop()
}

func NewZhimaWrapper(pool *zhima_proxy_pool.ZhimaProxyPool, deleteLimit int) *Wrapper {
	return &Wrapper{
		pool:        pool,
		deleteLimit: deleteLimit,
		deleteCount: make(map[string]int),
		mutex:       new(sync.RWMutex),
	}
}
