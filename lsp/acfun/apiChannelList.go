package acfun

import (
	"github.com/Sora233/DDBOT/proxy_pool"
	"github.com/Sora233/DDBOT/requests"
	"github.com/Sora233/DDBOT/utils"
	"time"
)

const (
	PathApiChannelList = "/api/channel/list"
)

type ApiChannelListRequest struct {
	Count   int32  `json:"count"`
	PCursor string `json:"pcursor"`
	// and maybe filter ...
}

func ApiChannelList(count int32, pcursor string) (*ApiChannelListResponse, error) {
	st := time.Now()
	defer func() {
		ed := time.Now()
		logger.WithField("FuncName", utils.FuncName()).Tracef("cost %v", ed.Sub(st))
	}()
	url := APath(PathApiChannelList)
	params, err := utils.ToParams(&ApiChannelListRequest{
		Count:   count,
		PCursor: pcursor,
	})
	if err != nil {
		return nil, err
	}
	var opts []requests.Option
	opts = append(opts,
		requests.ProxyOption(proxy_pool.PreferNone),
		requests.AddUAOption(),
		requests.TimeoutOption(time.Second*10),
	)
	apiChannelListResp := new(ApiChannelListResponse)
	err = requests.Get(url, params, &apiChannelListResp, opts...)
	if err != nil {
		return nil, err
	}
	return apiChannelListResp, nil
}
