package huya

import (
	"github.com/Sora233/DDBOT/concern"
	"github.com/Sora233/DDBOT/lsp/concern_manager"
)

type GroupConcernConfig struct {
	concern_manager.GroupConcernConfig
}

func (g *GroupConcernConfig) AtAllBeforeHook(notify concern.Notify) bool {
	switch notify.Type() {
	case concern.HuyaLive:
		return notify.(*ConcernLiveNotify).LiveStatusChanged
	default:
		return false
	}
}

func (g *GroupConcernConfig) ShouldSendHook(notify concern.Notify) bool {
	switch e := notify.(type) {
	case *ConcernLiveNotify:
		return e.Living
	default:
		return false
	}
}

func NewGroupConcernConfig(g *concern_manager.GroupConcernConfig) *GroupConcernConfig {
	return &GroupConcernConfig{*g}
}
