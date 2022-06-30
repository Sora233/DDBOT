package template

import (
	"github.com/Sora233/DDBOT/requests"
)

const (
	DDBOT_REQ_DEBUG = "DDBOT_REQ_DEBUG"
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

	var result []requests.Option

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
	}

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
