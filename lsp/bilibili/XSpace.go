package bilibili

import (
	"context"
	"github.com/Sora233/DDBOT/proxy_pool"
	"github.com/Sora233/DDBOT/proxy_pool/requests"
	"github.com/Sora233/DDBOT/utils"
	"time"
)

const (
	PathSpaceAccInfo = "/x/space/acc/info"
)

type SpaceAccInfoRequest struct {
	Mid int64 `json:"mid"`
}

func XSpaceAccInfo(mid int64) (*XSpaceAccInfoResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()
	st := time.Now()
	defer func() {
		ed := time.Now()
		logger.WithField("FuncName", utils.FuncName()).Tracef("cost %v", ed.Sub(st))
	}()
	url := BPath(PathSpaceAccInfo)
	params, err := utils.ToParams(&SpaceAccInfoRequest{
		Mid: mid,
	})
	if err != nil {
		return nil, err
	}
	resp, err := requests.Get(ctx, url, params, 1,
		requests.ProxyOption(proxy_pool.PreferNone),
		requests.TimeoutOption(time.Second*5),
		AddUAOption(),
	)
	if err != nil {
		return nil, err
	}
	xsai := new(XSpaceAccInfoResponse)
	err = resp.Json(xsai)
	if err != nil {
		return nil, err
	}
	if xsai.Code == -412 && resp.Proxy != "" {
		proxy_pool.Delete(resp.Proxy)
	}
	return xsai, nil
}
