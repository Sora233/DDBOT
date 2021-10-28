package bilibili

import (
	"github.com/Sora233/DDBOT/proxy_pool"
	"github.com/Sora233/DDBOT/requests"
	"github.com/Sora233/DDBOT/utils"
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
	var opts = []requests.Option{
		requests.ProxyOption(proxy_pool.PreferNone),
		AddUAOption(),
		requests.TimeoutOption(time.Second * 10),
		delete412ProxyOption,
	}
	rir := new(RoomInitResponse)
	err = requests.Get(url, params, rir, opts...)
	if err != nil {
		return nil, err
	}
	return rir, nil
}
