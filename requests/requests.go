package requests

import (
	"github.com/Sora233/DDBOT/proxy_pool"
	"github.com/Sora233/MiraiGo-Template/utils"
	"github.com/guonaihong/gout"
	"github.com/guonaihong/gout/dataflow"
	middleware "github.com/guonaihong/gout/interface"
	"io"
	"net/http"
	"strings"
	"time"
)

var logger = utils.GetModuleLogger("request")

type option struct {
	Timeout            time.Duration
	InsecureSkipVerify bool

	Debug               bool
	Cookies             []*http.Cookie
	Header              gout.H
	Proxy               string
	HttpCode            *int
	Retry               int
	ProxyCallbackOption func(out interface{}, proxy string)
	CookieJar           http.CookieJar
	ResponseMiddleware  []middleware.ResponseMiddler
}

func (o *option) getGout() *gout.Client {
	var goutOpts []gout.Option
	if o.Timeout != 0 {
		goutOpts = append(goutOpts, gout.WithTimeout(o.Timeout))
	} else {
		goutOpts = append(goutOpts, gout.WithTimeout(time.Second*5))
	}
	if o.InsecureSkipVerify {
		goutOpts = append(goutOpts, gout.WithInsecureSkipVerify())
	}
	if o.CookieJar != nil {
		goutOpts = append(goutOpts, gout.WithClient(&http.Client{
			Jar: o.CookieJar,
		}))
	}
	return gout.NewWithOpt(goutOpts...)
}

type Option func(o *option)

func empty(*option) {}

func HttpCookieOption(cookie *http.Cookie) Option {
	return func(o *option) {
		o.Cookies = append(o.Cookies, cookie)
	}
}

func CookieOption(name, value string) Option {
	return HttpCookieOption(&http.Cookie{Name: name, Value: value})
}

func TimeoutOption(d time.Duration) Option {
	return func(o *option) {
		o.Timeout = d
	}
}

func HttpCodeOption(code *int) Option {
	return func(o *option) {
		o.HttpCode = code
	}
}

func HeaderOption(key, value string) Option {
	return func(o *option) {
		if o.Header == nil {
			o.Header = make(gout.H)
		}
		o.Header[key] = value
	}
}

func AddUAOption() Option {
	return HeaderOption("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.212 Safari/537.36")
}

func ProxyOption(prefer proxy_pool.Prefer) Option {
	if prefer == proxy_pool.PreferNone {
		return empty
	}
	proxy, err := proxy_pool.Get(prefer)
	if err != nil {
		if err != proxy_pool.ErrNil {
			logger.Errorf("get proxy failed")
		}
		return empty
	} else {
		return func(o *option) {
			o.Proxy = proxy.ProxyString()
		}
	}
}

func RetryOption(retry int) Option {
	return func(o *option) {
		o.Retry = retry
	}
}

func ProxyCallbackOption(f func(out interface{}, proxy string)) Option {
	return func(o *option) {
		o.ProxyCallbackOption = f
	}
}

func DisableTlsOption() Option {
	return func(o *option) {
		o.InsecureSkipVerify = true
	}
}

func DebugOption() Option {
	return func(o *option) {
		o.Debug = true
	}
}

// WithCookieJar CookieJar可能导致Cookie泄漏，谨慎使用
func WithCookieJar(jar http.CookieJar) Option {
	return func(o *option) {
		o.CookieJar = jar
	}
}

func WithResponseMiddleware(middler middleware.ResponseMiddler) Option {
	return func(o *option) {
		o.ResponseMiddleware = append(o.ResponseMiddleware, middler)
	}
}

func GetResponseCookieOption(cookies *[]*http.Cookie) Option {
	if cookies == nil {
		return empty
	}
	return WithResponseMiddleware(middleware.WithResponseMiddlerFunc(
		func(response *http.Response) error {
			*cookies = response.Cookies()
			return nil
		},
	))
}

func Do(f func(*gout.Client) *dataflow.DataFlow, out interface{}, options ...Option) error {
	var opt = new(option)
	for _, o := range options {
		o(opt)
	}
	if opt.ProxyCallbackOption != nil && len(opt.Proxy) > 0 {
		defer func() {
			opt.ProxyCallbackOption(out, opt.Proxy)
		}()
	}
	var df = f(opt.getGout())
	if opt.Debug {
		df.Debug(true)
	}
	if len(opt.Cookies) > 0 {
		df.SetCookies(opt.Cookies...)
	}
	if len(opt.Header) > 0 {
		df.SetHeader(opt.Header)
	}
	if len(opt.Proxy) > 0 {
		if strings.HasPrefix(opt.Proxy, "socks5:") {
			df.SetSOCKS5(opt.Proxy)
		} else {
			df.SetProxy(opt.Proxy)
		}
	}
	if opt.HttpCode != nil {
		df.Code(opt.HttpCode)
	}
	if opt.ResponseMiddleware != nil {
		df.ResponseUse(opt.ResponseMiddleware...)
	}
	switch out.(type) {
	case io.Writer, []byte, *string:
		df.BindBody(out)
	default:
		df.BindJSON(out)
	}
	if opt.Retry > 0 {
		return df.F().Retry().Attempt(opt.Retry).Do()
	}
	return df.Do()
}

func Get(url string, params gout.H, out interface{}, options ...Option) error {
	return Do(func(gcli *gout.Client) *dataflow.DataFlow {
		return gcli.GET(url).SetQuery(params)
	}, out, options...)
}

func PostForm(url string, params gout.H, out interface{}, options ...Option) error {
	return Do(func(gcli *gout.Client) *dataflow.DataFlow {
		return gcli.POST(url).SetForm(params)
	}, out, options...)
}

func PostJson(url string, params gout.H, out interface{}, options ...Option) error {
	return Do(func(gcli *gout.Client) *dataflow.DataFlow {
		return gcli.POST(url).SetJSON(params)
	}, out, options...)
}

func PostWWWForm(url string, params gout.H, out interface{}, options ...Option) error {
	return Do(func(gcli *gout.Client) *dataflow.DataFlow {
		return gcli.POST(url).SetWWWForm(params)
	}, out, options...)
}

func PostBody(url string, body []byte, out interface{}, options ...Option) error {
	return Do(func(gcli *gout.Client) *dataflow.DataFlow {
		return gcli.POST(url).SetBody(body)
	}, out, options...)
}
