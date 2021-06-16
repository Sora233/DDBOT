package bilibili

import (
	"github.com/Sora233/DDBOT/concern"
	"github.com/Sora233/DDBOT/lsp/concern_manager"
)

type GroupConcernConfig struct {
	concern_manager.GroupConcernConfig
}

func (g *GroupConcernConfig) AtAllBeforeHook(notify concern.Notify) bool {
	switch notify.Type() {
	case concern.BibiliLive:
		return notify.(*ConcernLiveNotify).LiveStatusChanged
	case concern.BilibiliNews:
		return true
	default:
		return false
	}
}

func (g *GroupConcernConfig) ShouldSendHook(notify concern.Notify) bool {
	switch e := notify.(type) {
	case *ConcernLiveNotify:
		if e.Status != LiveStatus_Living {
			return false
		}
		if e.LiveStatusChanged {
			return true
		}
		if e.LiveTitleChanged {
			return g.GroupConcernNotify.CheckTitleChangeNotify(notify.Type())
		}
		return true
	case *ConcernNewsNotify:
		return true
	default:
		return false
	}
}

func NewGroupConcernConfig(g *concern_manager.GroupConcernConfig) *GroupConcernConfig {
	return &GroupConcernConfig{*g}
}
