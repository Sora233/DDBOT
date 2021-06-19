package douyu

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLiveInfo(t *testing.T) {
	l := &LiveInfo{
		ShowStatus: ShowStatus_Living,
		VideoLoop:  VideoLoopStatus_On,
	}
	assert.False(t, l.Living())
	l = &LiveInfo{
		ShowStatus: ShowStatus_Living,
		VideoLoop:  VideoLoopStatus_Off,
	}
	assert.True(t, l.Living())
	l = &LiveInfo{
		ShowStatus: ShowStatus_NoLiving,
		VideoLoop:  VideoLoopStatus_Off,
	}
	assert.False(t, l.Living())

	l = &LiveInfo{
		Nickname: "nickname",
		RoomId:   123,
		RoomName: "roomname",
		RoomUrl:  "url",
		Avatar: &Avatar{
			Big:    "big",
			Middle: "middle",
			Small:  "small",
		},
	}
	assert.Equal(t, "nickname", l.GetNickname())
	assert.Equal(t, int64(123), l.GetRoomId())
	assert.Equal(t, "roomname", l.GetRoomName())
	assert.Equal(t, "url", l.GetRoomUrl())
	assert.Equal(t, "big", l.GetAvatar().GetBig())
	assert.Equal(t, "middle", l.GetAvatar().GetMiddle())
	assert.Equal(t, "small", l.GetAvatar().GetSmall())
}
