package bilibili

import (
	"context"
	"github.com/Sora233/DDBOT/proxy_pool"
	"github.com/Sora233/DDBOT/requests"
	"github.com/Sora233/DDBOT/utils"
	"time"
)

const (
	PathXRelationStat = "/x/relation/stat"
)

type XRelationStatRequest struct {
	Mid int64 `json:"vmid"`
}

func XRelationStat(mid int64) (*XRelationStatResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()
	st := time.Now()
	defer func() {
		ed := time.Now()
		logger.WithField("FuncName", utils.FuncName()).Tracef("cost %v", ed.Sub(st))
	}()
	url := BPath(PathXRelationStat)
	params, err := utils.ToParams(&XRelationStatRequest{
		Mid: mid,
	})
	if err != nil {
		return nil, err
	}
	var opts = []requests.Option{
		requests.ProxyOption(proxy_pool.PreferNone),
		requests.TimeoutOption(time.Second * 5),
		AddUAOption(),
	}
	opts = append(opts, GetVerifyOption()...)
	resp, err := requests.Get(ctx, url, params, 1, opts...)
	if err != nil {
		return nil, err
	}
	xrsr := new(XRelationStatResponse)
	err = resp.Json(xrsr)
	if err != nil {
		return nil, err
	}
	if xrsr.Code == -412 && resp.Proxy != "" {
		proxy_pool.Delete(resp.Proxy)
	}
	return xrsr, nil
}
