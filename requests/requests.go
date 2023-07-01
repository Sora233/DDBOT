package requests

import (
	"fmt"
	"github.com/Sora233/DDBOT/proxy_pool"
	"github.com/Sora233/MiraiGo-Template/utils"
	"github.com/guonaihong/gout"
	"github.com/guonaihong/gout/dataflow"
	"github.com/guonaihong/gout/middler"
	"io"
	"net/http"
	"strings"
	"time"
)

var logger = utils.GetModuleLogger("request")

var defaultUA = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/105.0.0.0 Safari/537.36"

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
	ResponseMiddleware  []middler.ResponseMiddler
	AutoHeaderHost      bool
	NotIgnoreEmpty      bool
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
	df := gout.NewWithOpt(goutOpts...)
	if o.NotIgnoreEmpty {
		df.NotIgnoreEmpty = true
	}
	return df
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

func AddUAOption(ua ...string) Option {
	if len(ua) > 0 && len(ua[0]) > 0 {
		return HeaderOption("user-agent", ua[0])
	}
	return HeaderOption("user-agent", defaultUA)
}

func AddRandomUAOption(entry FakeUAEntry) Option {
	return HeaderOption("user-agent", RandomUA(entry))
}

func ProxyOption(prefer proxy_pool.Prefer) Option {
	if prefer == proxy_pool.PreferNone {
		return empty
	}
	proxy, err := proxy_pool.Get(prefer)
	if err != nil {
		if err != proxy_pool.ErrNil {
			logger.Errorf("get proxy failed: %v", err)
		}
		return empty
	} else {
		return func(o *option) {
			o.Proxy = proxy.ProxyString()
		}
	}
}

func RawProxyOption(proxy string) Option {
	return func(o *option) {
		o.Proxy = proxy
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

func RequestAutoHostOption() Option {
	return func(o *option) {
		o.AutoHeaderHost = true
	}
}

func NotIgnoreEmptyOption() Option {
	return func(o *option) {
		o.NotIgnoreEmpty = true
	}
}

// WithCookieJar CookieJar可能导致Cookie泄漏，谨慎使用
func WithCookieJar(jar http.CookieJar) Option {
	return func(o *option) {
		o.CookieJar = jar
	}
}

func WithResponseMiddleware(middler middler.ResponseMiddler) Option {
	return func(o *option) {
		o.ResponseMiddleware = append(o.ResponseMiddleware, middler)
	}
}

func GetResponseCookieOption(cookies *[]*http.Cookie) Option {
	if cookies == nil {
		return empty
	}
	return WithResponseMiddleware(middler.WithResponseMiddlerFunc(
		func(response *http.Response) error {
			*cookies = response.Cookies()
			return nil
		},
	))
}

func Do(f func(*gout.Client) *dataflow.DataFlow, out interface{}, options ...Option) error {
	var (
		opt  = new(option)
		code int
		err  error
	)
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
	if opt.AutoHeaderHost {
		if h, err := df.GetHost(); err == nil {
			opt.Header["host"] = h
		}
	}
	if len(opt.Cookies) > 0 {
		df.SetCookies(opt.Cookies...)
	}
	if len(opt.Header) > 0 {
		df.SetHeader(opt.Header)
	}
	if len(opt.Proxy) > 0 {
		if strings.HasPrefix(opt.Proxy, "socks5://") {
			df.SetSOCKS5(strings.TrimPrefix(opt.Proxy, "socks5://"))
		} else {
			df.SetProxy(opt.Proxy)
		}
	}
	df.Code(&code)
	if opt.ResponseMiddleware != nil {
		df.ResponseUse(opt.ResponseMiddleware...)
	}
	switch out.(type) {
	case io.Writer, []byte, *[]byte, *string:
		df.BindBody(out)
	default:
		df.BindJSON(out)
	}
	if opt.Retry > 0 {
		err = df.F().Retry().Attempt(opt.Retry).Do()
	} else {
		err = df.Do()
	}
	if opt.HttpCode != nil {
		*opt.HttpCode = code
	}
	if err != nil {
		return err
	}
	if code >= http.StatusBadRequest {
		return fmt.Errorf("http code error %v", code)
	}
	return nil
}

func Get(url string, params interface{}, out interface{}, options ...Option) error {
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
