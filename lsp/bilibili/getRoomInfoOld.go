package bilibili

import (
	"context"
	"github.com/Sora233/Sora233-MiraiGo/proxy_pool"
	"github.com/Sora233/Sora233-MiraiGo/proxy_pool/requests"
	"github.com/Sora233/Sora233-MiraiGo/utils"
	"net/http"
	"time"
)

const (
	PathGetRoomInfoOld = "/room/v1/Room/getRoomInfoOld"
)

type GetRoomInfoOldRequest struct {
	Mid int64 `json:"mid"`
}

func GetRoomInfoOld(mid int64) (*GetRoomInfoOldResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	st := time.Now()
	defer func() {
		ed := time.Now()
		logger.WithField("FuncName", utils.FuncName()).Tracef("cost %v", ed.Sub(st))
	}()
	url := BPath(PathGetRoomInfoOld)
	params, err := utils.ToParams(&GetRoomInfoOldRequest{
		Mid: mid,
	})
	if err != nil {
		return nil, err
	}
	resp, err := requests.Get(ctx, url, params, 3, requests.CookieOption(&http.Cookie{Name: "DedeUserID", Value: "2"}), requests.ProxyOption(proxy_pool.PreferNone))
	if err != nil {
		return nil, err
	}
	grioResp := new(GetRoomInfoOldResponse)
	err = resp.Json(grioResp)
	if err != nil {
		return nil, err
	}
	if grioResp.Code == -412 && resp.Proxy != "" {
		proxy_pool.Delete(resp.Proxy)
	}
	return grioResp, nil
}
