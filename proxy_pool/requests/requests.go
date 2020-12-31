package requests

import (
	"github.com/Logiase/MiraiGo-Template/utils"
	"github.com/Sora233/Sora233-MiraiGo/proxy_pool"
	"github.com/asmcos/requests"
	"net/http"
	"time"
)

var logger = utils.GetModuleLogger("request")

type GetOption func(*requests.Request)

func CookieOption(cookie *http.Cookie) GetOption {
	return func(request *requests.Request) {
		request.SetCookie(cookie)
	}
}
func TimeoutOption(d time.Duration) GetOption {
	return func(request *requests.Request) {
		request.SetTimeout(d)
	}
}

var DefaultTimeoutOption = TimeoutOption(time.Second * 5)

type ResponseWithProxy struct {
	*requests.Response
	Proxy proxy_pool.IProxy
}

func Get(url string, params requests.Params, options ...GetOption) (*ResponseWithProxy, error) {
	req := requests.Requests()
	DefaultTimeoutOption(req)
	for _, opt := range options {
		opt(req)
	}
	proxy, err := proxy_pool.Get()
	if err != nil {
		if err != proxy_pool.ErrNil {
			logger.Errorf("get proxy failed")
		}
	} else {
		req.Proxy("http://" + proxy.ProxyString())
	}
	resp, err := req.Get(url, params)
	if err != nil {
		proxy_pool.Delete(proxy)
	}
	return &ResponseWithProxy{
		Response: resp,
		Proxy:    proxy,
	}, err
}
