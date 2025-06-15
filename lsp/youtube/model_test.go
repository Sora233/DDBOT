package youtube

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Sora233/DDBOT/v2/internal/test"
)

func TestVideoInfo(t *testing.T) {
	vi := &VideoInfo{
		UserInfo:  *NewUserInfo(test.NAME1, test.NAME2),
		VideoId:   test.BVID1,
		VideoType: VideoType_Video,
	}
	assert.EqualValues(t, test.NAME2, vi.GetChannelName())
	assert.Equal(t, VideoType_Video, vi.VideoType)
	assert.Equal(t, Video, vi.Type())
	assert.True(t, vi.IsVideo())

	info := NewInfo([]*VideoInfo{vi})
	assert.NotNil(t, info)

	notify := NewConcernNotify(test.G1, vi)
	assert.NotNil(t, notify)
	assert.Equal(t, test.G1, notify.GetGroupCode())
	assert.Equal(t, test.NAME1, notify.GetUid())
	assert.NotNil(t, notify.Logger())
	assert.Equal(t, Video, notify.Type())

	assert.Equal(t, Site, notify.Site())

	m := notify.ToMessage()
	assert.NotNil(t, m)

	notify.VideoType = VideoType_Live
	m = notify.ToMessage()
	assert.NotNil(t, m)

	notify.VideoStatus = VideoStatus_Living
	m = notify.ToMessage()
	assert.NotNil(t, m)

}
