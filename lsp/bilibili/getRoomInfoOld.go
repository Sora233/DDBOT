package bilibili

import (
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
	req, err := GetBilibiliRequest()
	if err != nil {
		return nil, err
	}
	req.SetCookie(&http.Cookie{
		Name:  "DedeUserID",
		Value: "2",
	})
	resp, err := req.Get(url, params)
	if err != nil {
		return nil, err
	}
	grioResp := new(GetRoomInfoOldResponse)
	err = resp.Json(grioResp)
	if err != nil {
		return nil, err
	}
	return grioResp, nil
}
