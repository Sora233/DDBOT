package huya

import (
	"github.com/Sora233/DDBOT/concern"
	"github.com/Sora233/DDBOT/lsp/concern_manager"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGroupConcernConfig_AtBeforeHook(t *testing.T) {
	g := NewGroupConcernConfig(new(concern_manager.GroupConcernConfig))
	notify := &ConcernLiveNotify{
		LiveInfo: LiveInfo{
			Living:            true,
			LiveStatusChanged: false,
			LiveTitleChanged:  false,
		},
	}
	hook := g.AtBeforeHook(notify)
	assert.False(t, hook.Pass)

	notify.LiveStatusChanged = true
	hook = g.AtBeforeHook(notify)
	assert.True(t, hook.Pass)

	notify.Living = false
	hook = g.AtBeforeHook(notify)
	assert.False(t, hook.Pass)
}

func TestGroupConcernConfig_ShouldSendHook(t *testing.T) {
	g := NewGroupConcernConfig(new(concern_manager.GroupConcernConfig))
	notify := &ConcernLiveNotify{
		LiveInfo: LiveInfo{
			Living:            true,
			LiveStatusChanged: false,
			LiveTitleChanged:  false,
		},
	}
	hook := g.ShouldSendHook(notify)
	assert.False(t, hook.Pass)

	notify.LiveTitleChanged = true
	g.GroupConcernNotify.TitleChangeNotify = concern.HuyaLive

	hook = g.ShouldSendHook(notify)
	assert.True(t, hook.Pass)

	g.GroupConcernNotify.OfflineNotify = concern.HuyaLive
	notify.Living = false
	notify.LiveStatusChanged = true
	hook = g.ShouldSendHook(notify)
	assert.True(t, hook.Pass)

	notify.Living = true
	hook = g.ShouldSendHook(notify)
	assert.True(t, hook.Pass)
}
