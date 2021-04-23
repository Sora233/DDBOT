package lolicon_pool

import (
	"container/list"
	"errors"
	"fmt"
	"github.com/Logiase/MiraiGo-Template/utils"
	"github.com/Sora233/DDBOT/image_pool"
	"sync"
	"time"
)

var logger = utils.GetModuleLogger("lolicon_pool")

type Config struct {
	ApiKey   string
	CacheMin int
	CacheMax int
}

type LoliconPool struct {
	config  *Config
	cache   map[R18Type]*list.List
	cond    *sync.Cond
	changed bool
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

func (pool *LoliconPool) Get(options ...image_pool.OptionFunc) ([]image_pool.Image, error) {
	option := make(image_pool.Option)
	for _, optionFunc := range options {
		optionFunc(option)
	}

	var (
		r18     R18Type = R18Off
		keyword string
		num     int = 1
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
	if keyword != "" {
		logger.Debugf("request remote image")
		resp, err := LoliconAppSetu(pool.config.ApiKey, r18, keyword, num)
		if err != nil {
			return nil, err
		}
		logger.WithField("image num", len(resp.Data)).
			WithField("quota", resp.Quota).
			WithField("quota_min_ttl", resp.QuotaMinTTL).
			Debugf("request done")
		switch resp.Code {
		case 404:
			return nil, ErrNotFound
		case 401:
			return nil, ErrAPIKeyError
		case 429:
			return nil, ErrQuotaExceed
		}
		if resp.Code != 0 {
			return nil, fmt.Errorf("response code %v: %v", resp.Code, resp.Msg)
		}
		var result []image_pool.Image
		for _, img := range resp.Data {
			result = append(result, img)
		}
		return result, nil
	}
	return pool.getCache(r18, num)
}

func (pool *LoliconPool) getCache(r18 R18Type, num int) (result []image_pool.Image, err error) {
	pool.cond.L.Lock()
	defer pool.cond.L.Unlock()
	for i := 0; i < num; i++ {
		if pool.cache[r18].Len() == 0 {
			err = pool.fillCacheFromRemote(r18)
			if err != nil {
				logger.WithField("from", "getCache").Errorf("fill cache from remote failed %v", err)
				break
			}
		}
		result = append(result, pool.cache[r18].Remove(pool.cache[r18].Front()).(*Setu))
		pool.changed = true
	}
	pool.cond.Signal()
	return
}

// caller must hold the lock
func (pool *LoliconPool) fillCacheFromRemote(r18 R18Type) error {
	logger.WithField("r18", r18.String()).Debug("fetch from remote")
	resp, err := LoliconAppSetu(pool.config.ApiKey, r18, "", 10)
	if err != nil {
		return err
	}
	logger.WithField("Quota", resp.Quota).
		WithField("QuotaMinTTL", resp.QuotaMinTTL).
		WithField("Msg", resp.Msg).
		WithField("Code", resp.Code).
		Debug("LoliconPool response")
	if resp.Code != 0 {
		return fmt.Errorf("response code %v: %v", resp.Code, resp.Msg)
	}
	for _, s := range resp.Data {
		pool.cache[r18].PushFront(s)
	}
	pool.changed = true
	return nil
}

func (pool *LoliconPool) background() {
	go func() {
		for range time.Tick(time.Second * 30) {
			pool.store()
		}
	}()
	for {
		var result = true
		pool.cond.L.Lock()
		for {
			var checkResult = false
			for _, v := range pool.cache {
				if v.Len() < pool.config.CacheMin {
					checkResult = true
				}
			}
			if checkResult {
				break
			}
			pool.cond.Wait()
		}
		for r18, l := range pool.cache {
			if l.Len() < pool.config.CacheMin {
				for l.Len() < pool.config.CacheMax {
					pool.changed = true
					if err := pool.fillCacheFromRemote(r18); err != nil {
						logger.WithField("from", "background").Errorf("fill cache from remote failed %v", err)
						result = false
						break
					}
				}
			}
		}
		pool.cond.L.Unlock()
		if !result {
			time.Sleep(time.Minute)
		}
	}
}

func NewLoliconPool(config *Config) (*LoliconPool, error) {
	if config.ApiKey == "" {
		return nil, errors.New("empty api key")
	}
	if config.CacheMin == 0 {
		config.CacheMin = 20
	}
	if config.CacheMax == 0 {
		config.CacheMax = 50
	}
	if config.CacheMin > config.CacheMax {
		config.CacheMin = config.CacheMax
	}
	pool := &LoliconPool{
		config:  config,
		cache:   make(map[R18Type]*list.List),
		cond:    sync.NewCond(&sync.Mutex{}),
		changed: false,
	}
	pool.cache[R18Off] = list.New()
	pool.cache[R18On] = list.New()
	pool.load()
	go pool.background()
	return pool, nil
}
