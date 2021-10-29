package bilibili

import (
	"github.com/Sora233/DDBOT/proxy_pool"
	"github.com/Sora233/DDBOT/requests"
	"github.com/Sora233/DDBOT/utils"
	"time"
)

const (
	PathXSpaceAccInfo = "/x/space/acc/info"
)

type XSpaceAccInfoRequest struct {
	Mid int64 `json:"mid"`
}

func XSpaceAccInfo(mid int64) (*XSpaceAccInfoResponse, error) {
	st := time.Now()
	defer func() {
		ed := time.Now()
		logger.WithField("FuncName", utils.FuncName()).Tracef("cost %v", ed.Sub(st))
	}()
	url := BPath(PathXSpaceAccInfo)
	params, err := utils.ToParams(&XSpaceAccInfoRequest{
		Mid: mid,
	})
	if err != nil {
		return nil, err
	}
	var opts = []requests.Option{
		requests.ProxyOption(proxy_pool.PreferNone),
		requests.TimeoutOption(time.Second * 15),
		AddUAOption(),
		delete412ProxyOption,
	}
	opts = append(opts, GetVerifyOption()...)
	xsai := new(XSpaceAccInfoResponse)
	err = requests.Get(url, params, xsai, opts...)
	if err != nil {
		return nil, err
	}
	return xsai, nil
}
