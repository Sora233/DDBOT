package douyu

import (
	"github.com/Sora233/DDBOT/internal/test"
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
		Nickname:   "nickname",
		RoomId:     test.UID1,
		RoomName:   "roomname",
		RoomUrl:    "url",
		ShowStatus: ShowStatus_Living,
		VideoLoop:  VideoLoopStatus_Off,
		Avatar: &Avatar{
			Big:    "big",
			Middle: "middle",
			Small:  "small",
		},
	}
	assert.Equal(t, "nickname", l.GetNickname())
	assert.Equal(t, test.UID1, l.GetRoomId())
	assert.Equal(t, "roomname", l.GetRoomName())
	assert.Equal(t, "url", l.GetRoomUrl())
	assert.Equal(t, "big", l.GetAvatar().GetBig())
	assert.Equal(t, "middle", l.GetAvatar().GetMiddle())
	assert.Equal(t, "small", l.GetAvatar().GetSmall())
	assert.Equal(t, Live, l.Type())
	assert.Equal(t, ShowStatus_Living, l.GetShowStatus())
	assert.Equal(t, VideoLoopStatus_Off, l.GetVideoLoop())
	assert.False(t, l.GetLiveStatusChanged())

	notify := NewConcernLiveNotify(test.G1, l)
	assert.NotNil(t, notify)
	assert.Equal(t, Live, notify.Type())
	assert.NotNil(t, notify.Logger())
	assert.Equal(t, test.G1, notify.GetGroupCode())
	assert.Equal(t, test.UID1, notify.GetUid())

	m := notify.ToMessage()
	assert.NotNil(t, m)

	notify.ShowStatus = ShowStatus_NoLiving
	notify.VideoLoop = VideoLoopStatus_Off
	m = notify.ToMessage()
	assert.NotNil(t, m)

}
