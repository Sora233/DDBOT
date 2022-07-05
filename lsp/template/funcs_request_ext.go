package template

import (
	"github.com/Sora233/DDBOT/proxy_pool"
	"github.com/Sora233/DDBOT/requests"
	"github.com/spf13/cast"
	"reflect"
	"strings"
)

const (
	DDBOT_REQ_DEBUG      = "DDBOT_REQ_DEBUG"
	DDBOT_REQ_HEADER     = "DDBOT_REQ_HEADER"
	DDBOT_REQ_COOKIE     = "DDBOT_REQ_COOKIE"
	DDBOT_REQ_PROXY      = "DDBOT_REQ_PROXY"
	DDBOT_REQ_USER_AGENT = "DDBOT_REQ_USER_AGENT"
)

func preProcess(oParams []map[string]interface{}) (map[string]interface{}, []requests.Option) {
	var params map[string]interface{}
	if len(oParams) == 0 {
		return nil, nil
	} else if len(oParams) == 1 {
		params = oParams[0]
	} else {
		panic("given more than one params")
	}
	fn := func(key string, f func() []requests.Option) []requests.Option {
		var r []requests.Option

		if _, found := params[key]; found {
			r = f()
			delete(params, key)
		}
		return r
	}

	collectStringSlice := func(i interface{}) []string {
		v := reflect.ValueOf(i)
		if v.Kind() == reflect.String {
			return []string{v.String()}
		}
		return cast.ToStringSlice(i)
	}

	var item = []struct {
		key string
		f   func() []requests.Option
	}{
		{
			DDBOT_REQ_DEBUG,
			func() []requests.Option {
				return []requests.Option{requests.DebugOption()}
			},
		},
		{
			DDBOT_REQ_HEADER,
			func() []requests.Option {
				var result []requests.Option
				var header = collectStringSlice(params[DDBOT_REQ_HEADER])
				for _, h := range header {
					spt := strings.SplitN(h, "=", 2)
					if len(spt) >= 2 {
						result = append(result, requests.HeaderOption(spt[0], spt[1]))
					} else {
						logger.WithField("DDBOT_REQ_HEADER", h).Errorf("invalid header format")
					}
				}
				return result
			},
		},
		{
			DDBOT_REQ_COOKIE,
			func() []requests.Option {
				var result []requests.Option
				var cookie = collectStringSlice(params[DDBOT_REQ_COOKIE])
				for _, c := range cookie {
					spt := strings.SplitN(c, "=", 2)
					if len(spt) >= 2 {
						result = append(result, requests.CookieOption(spt[0], spt[1]))
					} else {
						logger.WithField("DDBOT_REQ_COOKIE", c).Errorf("invalid cookie format")
					}
				}
				return result
			},
		},
		{
			DDBOT_REQ_PROXY,
			func() []requests.Option {
				iproxy := params[DDBOT_REQ_PROXY]
				proxy, ok := iproxy.(string)
				if !ok {
					logger.WithField("DDBOT_REQ_PROXY", iproxy).Errorf("invalid proxy format")
					return nil
				}
				if proxy == "prefer_mainland" {
					return []requests.Option{requests.ProxyOption(proxy_pool.PreferMainland)}
				} else if proxy == "prefer_oversea" {
					return []requests.Option{requests.ProxyOption(proxy_pool.PreferOversea)}
				} else if proxy == "prefer_none" {
					return nil
				} else if proxy == "prefer_any" {
					return []requests.Option{requests.ProxyOption(proxy_pool.PreferAny)}
				} else {
					return []requests.Option{requests.RawProxyOption(proxy)}
				}
			},
		},
		{
			DDBOT_REQ_USER_AGENT,
			func() []requests.Option {
				iua := params[DDBOT_REQ_USER_AGENT]
				ua, ok := iua.(string)
				if !ok {
					logger.WithField("DDBOT_REQ_USER_AGENT", iua).Errorf("invalid ua format")
					return nil
				}
				return []requests.Option{requests.AddUAOption(ua)}
			},
		},
	}

	var result = []requests.Option{requests.AddUAOption()}
	for _, i := range item {
		result = append(result, fn(i.key, i.f)...)
	}
	return params, result
}

func httpGet(url string, oParams ...map[string]interface{}) (body []byte) {
	params, opts := preProcess(oParams)
	err := requests.Get(url, params, &body, opts...)
	if err != nil {
		logger.Errorf("template: httpGet error %v", err)
	}
	return
}

func httpPostJson(url string, oParams ...map[string]interface{}) (body []byte) {
	params, opts := preProcess(oParams)
	err := requests.PostJson(url, params, &body, opts...)
	if err != nil {
		logger.Errorf("template: httpGet error %v", err)
	}
	return
}

func httpPostForm(url string, oParams ...map[string]interface{}) (body []byte) {
	params, opts := preProcess(oParams)
	err := requests.PostForm(url, params, &body, opts...)
	if err != nil {
		logger.Errorf("template: httpGet error %v", err)
	}
	return
}
