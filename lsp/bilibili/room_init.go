package bilibili

import (
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
	resp, err := requests.Get(url, params, 3)
	if err != nil {
		return nil, err
	}
	rir := new(RoomInitResponse)
	err = resp.Json(rir)
	if err != nil {
		return nil, err
	}
	if rir.Code == -412 {
		proxy_pool.Delete(resp.Proxy)
	}
	return rir, nil
}
