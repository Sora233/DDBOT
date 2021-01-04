package local_pool

import (
	"github.com/Sora233/Sora233-MiraiGo/proxy_pool"
	"math/rand"
)

type Proxy struct {
	proxy string
}

func (p *Proxy) ProxyString() string {
	return p.proxy
}

type Pool struct {
	proxies []*Proxy
}

func (p *Pool) Get() (proxy_pool.IProxy, error) {
	return p.proxies[rand.Intn(len(p.proxies))], nil
}

func (p *Pool) Delete(iProxy proxy_pool.IProxy) bool {
	return false
}

func (p *Pool) Stop() error {
	return nil
}

func NewLocalPool(_proxies []string) *Pool {
	var proxies []*Proxy
	for _, proxy := range _proxies {
		proxies = append(proxies, &Proxy{proxy})
	}
	return &Pool{
		proxies: proxies,
	}
}
