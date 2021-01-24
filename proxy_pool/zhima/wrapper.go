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

func (z *Wrapper) Get(prefer proxy_pool.Prefer) (proxy_pool.IProxy, error) {
	z.mutex.RLock()
	defer z.mutex.RUnlock()
	return z.pool.Get()
}

func (z *Wrapper) Delete(proxy string) bool {
	z.mutex.Lock()
	defer z.mutex.Unlock()

	var result = false

	if _, found := z.deleteCount[proxy]; !found {
		z.deleteCount[proxy] = 1
	} else {
		z.deleteCount[proxy] += 1
	}

	if z.deleteCount[proxy] == z.deleteLimit {
		result = z.pool.Delete(proxy)
		delete(z.deleteCount, proxy)
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
