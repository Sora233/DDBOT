package bilibili

import (
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
	req, err := GetBilibiliRequest()
	if err != nil {
		return nil, err
	}
	resp, err := req.Get(url, params)
	if err != nil {
		return nil, err
	}
	xsai := new(XSpaceAccInfoResponse)
	err = resp.Json(xsai)
	if err != nil {
		return nil, err
	}
	return xsai, nil
}
