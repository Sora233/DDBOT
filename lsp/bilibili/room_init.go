package bilibili

import (
	"context"
	"github.com/Sora233/Sora233-MiraiGo/proxy_pool"
	"github.com/Sora233/Sora233-MiraiGo/proxy_pool/requests"
	"github.com/Sora233/Sora233-MiraiGo/utils"
	"time"
)

const (
	PathRoomInit = "/room/v1/Room/room_init"
)

type RoomInitRequest struct {
	Id int64 `json:"id"`
}

func RoomInit(roomId int64) (*RoomInitResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	st := time.Now()
	defer func() {
		ed := time.Now()
		logger.WithField("FuncName", utils.FuncName()).Tracef("cost %v", ed.Sub(st))
	}()
	url := BPath(PathRoomInit)
	params, err := utils.ToParams(&RoomInitRequest{
		Id: roomId,
	})
	if err != nil {
		return nil, err
	}
	resp, err := requests.Get(ctx, url, params, 3, requests.ProxyOption(proxy_pool.PreferNone))
	if err != nil {
		return nil, err
	}
	rir := new(RoomInitResponse)
	err = resp.Json(rir)
	if err != nil {
		return nil, err
	}
	if rir.Code == -412 && resp.Proxy != "" {
		proxy_pool.Delete(resp.Proxy)
	}
	return rir, nil
}
