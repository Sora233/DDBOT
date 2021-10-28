package bilibili

import (
	"github.com/Sora233/DDBOT/proxy_pool"
	"github.com/Sora233/DDBOT/requests"
	"github.com/Sora233/DDBOT/utils"
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
	var opts = []requests.Option{
		requests.ProxyOption(proxy_pool.PreferNone),
		AddUAOption(),
		requests.HttpCookieOption(&http.Cookie{Name: "DedeUserID", Value: "2"}),
		requests.HttpCookieOption(&http.Cookie{Name: "LIVE_BUVID", Value: genBUVID()}),
		requests.TimeoutOption(time.Second * 15),
		delete412ProxyOption,
	}
	grioResp := new(GetRoomInfoOldResponse)
	err = requests.Get(url, params, grioResp, opts...)
	if err != nil {
		return nil, err
	}
	return grioResp, nil
}

func genBUVID() string {
	return "AUTO" + strconv.FormatInt(atomic.LoadInt64(&buvid), 10)
}
