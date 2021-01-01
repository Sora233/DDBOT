package bilibili

import (
	"github.com/Sora233/Sora233-MiraiGo/proxy_pool"
	"github.com/Sora233/Sora233-MiraiGo/proxy_pool/requests"
	"github.com/Sora233/Sora233-MiraiGo/utils"
	"net/http"
)

const (
	PathGetRoomInfoOld = "/room/v1/Room/getRoomInfoOld"
)

type GetRoomInfoOldRequest struct {
	Mid int64 `json:"mid"`
}

func GetRoomInfoOld(mid int64) (*GetRoomInfoOldResponse, error) {
	url := BPath(PathGetRoomInfoOld)
	params, err := utils.ToParams(&GetRoomInfoOldRequest{
		Mid: mid,
	})
	if err != nil {
		return nil, err
	}
	resp, err := requests.Get(url, params, 3, requests.CookieOption(&http.Cookie{Name: "DedeUserID", Value: "2"}))
	if err != nil {
		return nil, err
	}
	grioResp := new(GetRoomInfoOldResponse)
	err = resp.Json(grioResp)
	if err != nil {
		return nil, err
	}
	if grioResp.Code != 0 {
		proxy_pool.Delete(resp.Proxy)
	}
	return grioResp, nil
}
