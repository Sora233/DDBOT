package bilibili

import (
	"errors"
	"github.com/asmcos/requests"
)

const (
	PathRoomInit = "/room/v1/Room/room_init"
)

type RoomInitRequest struct {
	Id int64 `json:"id"`
}

func RoomInit(roomId int64) (*RoomInitResponse, error) {
	url := BPath(PathRoomInit)
	params, err := BGetRequestToParams(&RoomInitRequest{
		Id: roomId,
	})
	if err != nil {
		return nil, err
	}
	resp, err := requests.Get(url, params)
	if err != nil {
		return nil, err
	}
	rir := new(RoomInitResponse)
	err = resp.Json(rir)
	if err != nil {
		return nil, err
	}
	if rir.GetCode() != 0 {
		return nil, errors.New(rir.Message)
	}
	return rir, nil
}
