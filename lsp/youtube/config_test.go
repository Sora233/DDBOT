package youtube

import (
	"github.com/Sora233/DDBOT/lsp/concern_manager"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGroupConcernConfig_AtBeforeHook(t *testing.T) {
	g := NewGroupConcernConfig(new(concern_manager.GroupConcernConfig))
	notify := &ConcernNotify{
		VideoInfo: VideoInfo{
			VideoStatus:       VideoStatus_Living,
			LiveStatusChanged: false,
		},
	}
	hook := g.AtBeforeHook(notify)
	assert.False(t, hook.Pass)

	notify.LiveStatusChanged = true
	hook = g.AtBeforeHook(notify)
	assert.True(t, hook.Pass)

	notify.VideoStatus = VideoStatus_Waiting
	hook = g.AtBeforeHook(notify)
	assert.False(t, hook.Pass)

	notify.VideoType = VideoType_Video
	notify.VideoStatus = VideoStatus_Upload
	hook = g.AtBeforeHook(notify)
	assert.True(t, hook.Pass)
}

func TestGroupConcernConfig_ShouldSendHook(t *testing.T) {
	g := NewGroupConcernConfig(new(concern_manager.GroupConcernConfig))
	notify := &ConcernNotify{
		VideoInfo: VideoInfo{
			VideoStatus:       VideoStatus_Living,
			LiveStatusChanged: false,
		},
	}
	hook := g.ShouldSendHook(notify)
	assert.True(t, hook.Pass)
}
