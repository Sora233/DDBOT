package douyu

import "errors"

var (
	ErrRoomNotExist = errors.New("房间不存在")
	ErrRoomBanned   = errors.New("房间已被关闭")
)
