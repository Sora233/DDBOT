package py

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/Sora233/DDBOT/v2/proxy_pool"
	"github.com/Sora233/DDBOT/v2/requests"
)

// ProxyPool is https://github.com/jhao104/proxy_pool
type ProxyPool struct {
	Host string
}

type GetResponse struct {
	CheckCount int    `json:"check_count"`
	FailCount  int    `json:"fail_count"`
	LastStatus int    `json:"last_status"`
	LastTime   string `json:"last_time"`
	Proxy      string `json:"proxy"`
	Region     string `json:"region"`
	Source     string `json:"source"`
	Type       string `json:"type"`
}

type PyProxy struct {
	proxy string
}

func (p *PyProxy) ProxyString() string {
	return p.proxy
}

func (pool *ProxyPool) Get(prefer proxy_pool.Prefer) (proxy_pool.IProxy, error) {
	if pool == nil {
		return nil, errors.New("<nil>")
	}
	getResp := new(GetResponse)
	err := requests.Get(fmt.Sprintf("%v/get", pool.Host), nil, getResp)
	if err != nil {
		return nil, err
	}
	if getResp.Proxy == "" {
		return nil, errors.New("empty proxy")
	}
	proxy := getResp.Proxy
	return &PyProxy{proxy}, nil
}

type DeleteResponse struct {
	Code int `json:"code"`
	Src  int `json:"src"`
}

func (pool *ProxyPool) Delete(proxy string) bool {
	if pool == nil {
		return false
	}
	deleteResp := new(DeleteResponse)
	err := requests.Get(fmt.Sprintf("%v/delete?proxy=%v", pool.Host, proxy), nil, deleteResp)
	if err != nil {
		return false
	}
	return deleteResp.Src == 1
}

func (pool *ProxyPool) Stop() error {
	return nil
}

func NewPYProxyPool(host string) (*ProxyPool, error) {
	pool := &ProxyPool{Host: host}
	var code int
	err := requests.Get(pool.Host, nil, requests.HttpCodeOption(&code))
	if err != nil {
		return nil, err
	}
	if code != 200 {
		return nil, errors.New(http.StatusText(code))
	}
	return pool, nil
}
