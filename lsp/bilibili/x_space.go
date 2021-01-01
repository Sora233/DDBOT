package bilibili

import (
	"github.com/Sora233/Sora233-MiraiGo/proxy_pool"
	"github.com/Sora233/Sora233-MiraiGo/proxy_pool/requests"
	"github.com/Sora233/Sora233-MiraiGo/utils"
)

const (
	PathSpaceAccInfo = "/x/space/acc/info"
)

type SpaceAccInfoRequest struct {
	Mid int64 `json:"mid"`
}

func XSpaceAccInfo(mid int64) (*XSpaceAccInfoResponse, error) {
	url := BPath(PathSpaceAccInfo)
	params, err := utils.ToParams(&SpaceAccInfoRequest{
		Mid: mid,
	})
	if err != nil {
		return nil, err
	}
	resp, err := requests.Get(url, params, 3)
	if err != nil {
		return nil, err
	}
	xsai := new(XSpaceAccInfoResponse)
	err = resp.Json(xsai)
	if err != nil {
		return nil, err
	}
	if xsai.Code != 0 {
		proxy_pool.Delete(resp.Proxy)
	}
	return xsai, nil
}
