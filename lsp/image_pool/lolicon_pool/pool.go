package lolicon_pool

import (
	"errors"
	"github.com/Logiase/MiraiGo-Template/utils"
	"github.com/Sora233/Sora233-MiraiGo/lsp/image_pool"
)

var logger = utils.GetModuleLogger("lolicon_pool")

type LoliconPool struct {
	apiKey string
}

type Image struct {
}

func KeywordOption(keyword string) image_pool.OptionFunc {
	return func(option image_pool.Option) image_pool.Option {
		option["keyword"] = keyword
		return option
	}
}

func NumOption(num int) image_pool.OptionFunc {
	return func(option image_pool.Option) image_pool.Option {
		option["num"] = num
		return option
	}
}

func R18Option(r18Type R18Type) image_pool.OptionFunc {
	return func(option image_pool.Option) image_pool.Option {
		option["r18"] = r18Type
		return option
	}
}

func (pool *LoliconPool) Get(options ...image_pool.OptionFunc) (image_pool.Image, error) {
	option := make(image_pool.Option)
	for _, optionFunc := range options {
		optionFunc(option)
	}

	var (
		r18     R18Type
		keyword string
		num     int
	)
	for k, v := range option {
		switch k {
		case "keyword":
			_v, ok := v.(string)
			if ok {
				keyword = _v
			}
		case "num":
			_v, ok := v.(int)
			if ok {
				num = _v
			}
		case "r18":
			_v, ok := v.(R18Type)
			if ok {
				r18 = _v
			}
		}
	}
	logger.Debugf("request remote image")
	resp, err := LoliconAppSetu(pool.apiKey, r18, keyword, num)
	if err != nil {
		return nil, err
	}
	logger.WithField("image num", len(resp.Data)).
		WithField("quota", resp.Quota).
		WithField("quota_min_ttl", resp.QuotaMinTTL).
		Debugf("request done")
	if len(resp.Data) >= 1 {
		return resp.Data[0], nil
	} else {
		return nil, errors.New("no image")
	}
}

func NewLoliconPool(apikey string) (*LoliconPool, error) {
	pool := &LoliconPool{
		apiKey: apikey,
	}
	return pool, nil
}
