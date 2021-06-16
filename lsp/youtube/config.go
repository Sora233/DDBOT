package youtube

import (
	"github.com/Sora233/DDBOT/concern"
	"github.com/Sora233/DDBOT/lsp/concern_manager"
)

type GroupConcernConfig struct {
	concern_manager.GroupConcernConfig
}

func (g *GroupConcernConfig) AtAllBeforeHook(notify concern.Notify) bool {
	switch notify.Type() {
	case concern.YoutubeLive:
		return true
	case concern.YoutubeVideo:
		return true
	default:
		return false
	}
}

func (g *GroupConcernConfig) ShouldSendHook(notify concern.Notify) bool {
	switch notify.(type) {
	case *ConcernNotify:
		return true
	default:
		return false
	}
}

func NewGroupConcernConfig(g *concern_manager.GroupConcernConfig) *GroupConcernConfig {
	return &GroupConcernConfig{*g}
}
