package local_proxy_pool

import (
	"github.com/Sora233/DDBOT/proxy_pool"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestProxy(t *testing.T) {
	var testCase = []*Proxy{
		{
			Proxy: "127.0.0.1:9999",
		},
		{
			Proxy: "http://localhost:10000",
		},
		{
			Proxy: "https://localhost:8888",
		},
		{
			Proxy: "xxx",
		},
	}
	var expected = []string{
		"http://127.0.0.1:9999",
		"http://localhost:10000",
		"https://localhost:8888",
		"http://xxx",
	}
	assert.EqualValues(t, len(expected), len(testCase))
	for idx := 0; idx < len(expected); idx++ {
		assert.Equal(t, expected[idx], testCase[idx].ProxyString())
	}
}

func TestProxyPool(t *testing.T) {
	pool := NewLocalPool([]*Proxy{
		{
			Proxy: "mainland",
			Type:  proxy_pool.PreferMainland,
		},
		{
			Proxy: "oversea",
			Type:  proxy_pool.PreferOversea,
		},
	})

	proxy, err := pool.Get(proxy_pool.PreferAny)
	assert.Nil(t, err)
	assert.NotNil(t, proxy)

	proxy, err = pool.Get(proxy_pool.PreferOversea)
	assert.Nil(t, err)
	assert.Equal(t, "http://oversea", proxy.ProxyString())

	proxy, err = pool.Get(proxy_pool.PreferMainland)
	assert.Nil(t, err)
	assert.Equal(t, "http://mainland", proxy.ProxyString())
	assert.Equal(t, proxy_pool.PreferMainland, proxy.(*Proxy).Prefer())

	_, err = pool.Get(proxy_pool.Prefer(9999))
	assert.NotNil(t, err)

	pool.Delete(proxy.ProxyString())
	pool.Stop()
}
