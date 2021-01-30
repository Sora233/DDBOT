package bilibili

import (
	"context"
	"github.com/Sora233/Sora233-MiraiGo/proxy_pool"
	"github.com/Sora233/Sora233-MiraiGo/proxy_pool/requests"
	"github.com/Sora233/Sora233-MiraiGo/utils"
	"math/rand"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"
)

const (
	PathGetRoomInfoOld = "/room/v1/Room/getRoomInfoOld"
)

var buvid int64 = 0

func init() {
	go func() {
		for {
			updateBuvid()
			time.Sleep(time.Second * 30)
		}
	}()
}

func updateBuvid() {
	atomic.StoreInt64(&buvid, rand.Int63n(9000000000000000)+1000000000000000)
}

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
	resp, err := requests.Get(ctx, url, params, 3,
		requests.ProxyOption(proxy_pool.PreferNone),
		requests.CookieOption(&http.Cookie{Name: "DedeUserID", Value: "2"}),
		requests.CookieOption(&http.Cookie{Name: "LIVE_BUVID", Value: genBUVID()}))
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

func genBUVID() string {
	return "AUTO" + strconv.FormatInt(atomic.LoadInt64(&buvid), 10)
}
