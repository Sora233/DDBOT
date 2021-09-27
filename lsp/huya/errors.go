package huya

import "errors"

var (
	ErrRoomNotExist = errors.New("房间不存在")
	ErrRoomBanned   = errors.New("涉嫌违规，正在整改中")
)
