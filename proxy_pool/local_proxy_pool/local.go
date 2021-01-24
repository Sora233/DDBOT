package local_proxy_pool

import (
	"errors"
	"github.com/Sora233/Sora233-MiraiGo/proxy_pool"
	"math/rand"
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
	proxies map[proxy_pool.Prefer][]*Proxy
	total   int
}

func (p *Pool) Get(prefer proxy_pool.Prefer) (proxy_pool.IProxy, error) {
	if prefer == proxy_pool.PreferNone {
		idx := rand.Intn(p.total)
		for _, v := range p.proxies {
			if len(v) > idx {
				return v[idx], nil
			} else {
				idx -= len(v)
			}
		}
		return nil, errors.New("out of range")
	} else {
		if s, found := p.proxies[prefer]; !found {
			return nil, errors.New("no proxy found")
		} else {
			return s[rand.Intn(len(s))], nil
		}
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
		total:   len(proxies),
	}
	for _, proxy := range proxies {
		if _, found := pool.proxies[proxy.Type]; !found {
			pool.proxies[proxy.Type] = make([]*Proxy, 0)
		}
		pool.proxies[proxy.Type] = append(pool.proxies[proxy.Type], proxy)
	}
	return pool
}
