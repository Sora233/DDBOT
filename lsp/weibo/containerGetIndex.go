package weibo

import (
	"github.com/Sora233/DDBOT/proxy_pool"
	"github.com/Sora233/DDBOT/requests"
	"github.com/Sora233/DDBOT/utils"
	"strconv"
	"time"
)

const (
	PathConcainerGetIndex_Profile = "https://m.weibo.cn/api/container/getIndex?containerid=100505"
	PathContainerGetIndex_Cards   = "https://m.weibo.cn/api/container/getIndex?containerid=107603"
)

func ApiContainerGetIndexProfile(uid int64) (*ApiContainerGetIndexProfileResponse, error) {
	st := time.Now()
	defer func() {
		ed := time.Now()
		logger.WithField("FuncName", utils.FuncName()).Tracef("cost %v", ed.Sub(st))
	}()
	path := PathConcainerGetIndex_Profile + strconv.FormatInt(uid, 10)
	var opts []requests.Option
	opts = append(opts,
		requests.ProxyOption(proxy_pool.PreferNone),
		requests.AddUAOption(),
		requests.TimeoutOption(time.Second*10),
	)
	opts = append(opts, CookieOption()...)
	profileResp := new(ApiContainerGetIndexProfileResponse)
	err := requests.Get(path, nil, &profileResp, opts...)
	if err != nil {
		return nil, err
	}
	return profileResp, nil
}

func ApiContainerGetIndexCards(uid int64) (*ApiContainerGetIndexCardsResponse, error) {
	st := time.Now()
	defer func() {
		ed := time.Now()
		logger.WithField("FuncName", utils.FuncName()).Tracef("cost %v", ed.Sub(st))
	}()
	path := PathContainerGetIndex_Cards + strconv.FormatInt(uid, 10)
	var opts []requests.Option
	opts = append(opts,
		requests.ProxyOption(proxy_pool.PreferNone),
		requests.AddUAOption(),
		requests.TimeoutOption(time.Second*10),
	)
	opts = append(opts, CookieOption()...)
	profileResp := new(ApiContainerGetIndexCardsResponse)
	err := requests.Get(path, nil, &profileResp, opts...)
	if err != nil {
		return nil, err
	}
	return profileResp, nil
}
