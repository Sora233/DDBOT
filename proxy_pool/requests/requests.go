package requests

import (
	"context"
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

func Get(ctx context.Context, url string, params requests.Params, maxRetry int, options ...GetOption) (*ResponseWithProxy, error) {
	req := requests.RequestsWithContext(ctx)
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

	var (
		resp  *requests.Response
		retry = 0
	)
	for {
		select {
		case <-ctx.Done():
			err = ctx.Err()
			break
		default:
		}
		resp, err = req.Get(url, params)
		if err != nil {
			retry += 1
		} else {
			break
		}
		logger.WithField("retry", retry).WithField("maxRetry", maxRetry).Debugf("request failed %v, retry", err)
		time.Sleep(time.Second)
		if retry == maxRetry {
			break
		}
	}
	if err != nil {
		proxy_pool.Delete(proxy)
	}
	return &ResponseWithProxy{
		Response: resp,
		Proxy:    proxy,
	}, err
}
