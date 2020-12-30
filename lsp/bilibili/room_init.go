package bilibili

import (
	"github.com/Sora233/Sora233-MiraiGo/utils"
)

const (
	PathRoomInit = "/room/v1/Room/room_init"
)

type RoomInitRequest struct {
	Id int64 `json:"id"`
}

func RoomInit(roomId int64) (*RoomInitResponse, error) {
	url := BPath(PathRoomInit)
	params, err := utils.ToParams(&RoomInitRequest{
		Id: roomId,
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
	rir := new(RoomInitResponse)
	err = resp.Json(rir)
	if err != nil {
		return nil, err
	}
	return rir, nil
}
