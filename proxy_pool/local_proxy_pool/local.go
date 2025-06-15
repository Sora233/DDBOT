package local_proxy_pool

import (
	"errors"
	"net/url"

	"go.uber.org/atomic"

	"github.com/Sora233/DDBOT/v2/proxy_pool"
)

type Proxy struct {
	Proxy string
	Type  proxy_pool.Prefer
}

func (p *Proxy) ProxyString() string {
	uri, err := url.Parse(p.Proxy)
	if err == nil {
		if len(uri.Scheme) == 0 {
			return "http://" + p.Proxy
		} else {
			return p.Proxy
		}
	}
	return "http://" + p.Proxy
}
func (p *Proxy) Prefer() proxy_pool.Prefer {
	return p.Type
}

type Pool struct {
	proxies   map[proxy_pool.Prefer][]*Proxy
	cnt       map[proxy_pool.Prefer]*atomic.Uint32
	total     int
	preferCnt atomic.Uint32
}

func (p *Pool) Get(prefer proxy_pool.Prefer) (proxy_pool.IProxy, error) {
	if prefer == proxy_pool.PreferAny {
		cnt := p.preferCnt.Add(1) % uint32(len(p.proxies))
		var index uint32 = 0
		for k := range p.proxies {
			if index == cnt {
				prefer = k
			} else {
				index++
			}
		}
	}

	if s, found := p.proxies[prefer]; !found {
		return nil, errors.New("no proxy found")
	} else {
		index := p.cnt[prefer].Add(1)
		return s[index%uint32(len(s))], nil
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
		cnt:     make(map[proxy_pool.Prefer]*atomic.Uint32),
		total:   len(proxies),
	}
	for _, proxy := range proxies {
		if _, found := pool.proxies[proxy.Type]; !found {
			pool.proxies[proxy.Type] = make([]*Proxy, 0)
			pool.cnt[proxy.Type] = atomic.NewUint32(0)
		}
		pool.proxies[proxy.Type] = append(pool.proxies[proxy.Type], proxy)
	}
	return pool
}
