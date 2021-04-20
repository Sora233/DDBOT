package requests

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/Logiase/MiraiGo-Template/utils"
	"github.com/Sora233/Sora233-MiraiGo/proxy_pool"
	"github.com/Sora233/requests"
	"net"
	"net/http"
	"time"
)

var logger = utils.GetModuleLogger("request")

type Option func(*requests.Request)

func HttpCookieOption(cookie *http.Cookie) Option {
	return func(request *requests.Request) {
		request.SetCookie(cookie)
	}
}

func CookieOption(name string, value string) Option {
	return HttpCookieOption(&http.Cookie{Name: name, Value: value})
}

func TimeoutOption(d time.Duration) Option {
	return func(request *requests.Request) {
		request.SetTimeout(d)
	}
}

func HeaderOption(key, value string) Option {
	return func(request *requests.Request) {
		request.Header.Set(key, value)
	}
}

func ProxyOption(prefer proxy_pool.Prefer) Option {
	return func(request *requests.Request) {
		if prefer == proxy_pool.PreferNone {
			return
		}
		proxy, err := proxy_pool.Get(prefer)
		if err != nil {
			if err != proxy_pool.ErrNil {
				logger.Errorf("get proxy failed")
			}
		} else {
			request.Proxy(proxy.ProxyString())
		}
	}
}

func DisableTlsOption() Option {
	return func(request *requests.Request) {
		if request.Client.Transport == nil {
			request.Client.Transport = &http.Transport{
				DialContext: (&net.Dialer{
					Timeout:   30 * time.Second,
					KeepAlive: 30 * time.Second,
					DualStack: true,
				}).DialContext,
				ForceAttemptHTTP2:     true,
				MaxIdleConns:          100,
				IdleConnTimeout:       90 * time.Second,
				TLSHandshakeTimeout:   10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			}
		} else {
			if tp, ok := request.Client.Transport.(*http.Transport); ok {
				if tp.TLSClientConfig != nil {
					tp.TLSClientConfig.InsecureSkipVerify = true
				} else {
					tp.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
				}
			}
		}
	}
}

func DebugOption() Option {
	return func(request *requests.Request) {
		request.Debug = 1
	}
}

var DefaultTimeoutOption = TimeoutOption(time.Second * 5)

type ResponseWithProxy struct {
	*requests.Response
	Proxy string
}

func Get(ctx context.Context, url string, params requests.Params, maxRetry int, options ...Option) (*ResponseWithProxy, error) {
	return anyHttp(ctx, maxRetry, func(request *requests.Request) (*requests.Response, error) {
		return request.Get(url, params)
	}, options...)
}

func PostJson(ctx context.Context, url string, params interface{}, maxRetry int, options ...Option) (*ResponseWithProxy, error) {
	return anyHttp(ctx, maxRetry, func(request *requests.Request) (*requests.Response, error) {
		return request.PostJson(url, params)
	}, options...)
}

func Post(ctx context.Context, url string, form requests.Datas, maxRetry int, options ...Option) (*ResponseWithProxy, error) {
	return anyHttp(ctx, maxRetry, func(request *requests.Request) (*requests.Response, error) {
		return request.Post(url, form)
	}, options...)
}

func anyHttp(ctx context.Context, maxRetry int, do func(request *requests.Request) (*requests.Response, error), options ...Option) (*ResponseWithProxy, error) {
	var err error
	req := requests.RequestsWithContext(ctx)
	DefaultTimeoutOption(req)
	for _, opt := range options {
		opt(req)
	}

	var (
		resp  *requests.Response
		retry = 0
	)
LOOP:
	for {
		select {
		case <-ctx.Done():
			err = ctx.Err()
			break LOOP
		default:
		}
		resp, err = do(req)
		if err != nil {
			retry += 1
		} else if resp.R.StatusCode != http.StatusOK {
			err = fmt.Errorf("status code %v", resp.R.StatusCode)
			retry += 1
		} else {
			break
		}
		logger.WithField("proxy", req.GetProxy()).WithField("retry", retry).WithField("maxRetry", maxRetry).Debugf("request failed %v, retry", err)
		if retry == maxRetry {
			break
		}
		time.Sleep(time.Second)
	}
	proxy := req.GetProxy()
	if err != nil && proxy != "" {
		proxy_pool.Delete(proxy)
	}
	return &ResponseWithProxy{
		Response: resp,
		Proxy:    proxy,
	}, err
}
