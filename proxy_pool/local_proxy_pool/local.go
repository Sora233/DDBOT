package local_proxy_pool

import (
	"errors"
	"github.com/Sora233/Sora233-MiraiGo/proxy_pool"
	"sync/atomic"
)

type Proxy struct {
	Proxy string
	Type  proxy_pool.Prefer
}

func (p *Proxy) ProxyString() string {
	return "http://" + p.Proxy
}
func (p *Proxy) Prefer() proxy_pool.Prefer {
	return p.Type
}

type Pool struct {
	proxies   map[proxy_pool.Prefer][]*Proxy
	cnt       map[proxy_pool.Prefer]*int64
	total     int
	preferCnt int64
}

func (p *Pool) Get(prefer proxy_pool.Prefer) (proxy_pool.IProxy, error) {
	if prefer == proxy_pool.PreferNone {
		cnt := atomic.AddInt64(&p.preferCnt, 1)
		if cnt%2 == 0 {
			prefer = proxy_pool.PreferOversea
		} else {
			prefer = proxy_pool.PreferMainland
		}
	}

	if s, found := p.proxies[prefer]; !found {
		return nil, errors.New("no proxy found")
	} else {
		index := atomic.AddInt64(p.cnt[prefer], 1)
		return s[index%int64(len(s))], nil
	}
}

func (p *Pool) Delete(proxy string) bool {
	return false
}

func (p *Pool) Stop() error {
	return nil
}

func NewLocalPool(proxies []*Proxy) *Pool {
	pool := &Pool{
		proxies: make(map[proxy_pool.Prefer][]*Proxy),
		cnt:     make(map[proxy_pool.Prefer]*int64),
		total:   len(proxies),
	}
	for _, proxy := range proxies {
		if _, found := pool.proxies[proxy.Type]; !found {
			pool.proxies[proxy.Type] = make([]*Proxy, 0)
			pool.cnt[proxy.Type] = new(int64)
		}
		pool.proxies[proxy.Type] = append(pool.proxies[proxy.Type], proxy)
	}
	return pool
}
