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
		if e.LiveStatusChanged {
			if !e.Living {
				// 下播了，检查下播推送配置
				return g.GroupConcernNotify.CheckOfflineNotify(notify.Type())
			} else {
				// 上播了，推
				return true
			}
		}
		if e.LiveTitleChanged {
			// 直播间标题改了，检查改标题推送配置
			return g.GroupConcernNotify.CheckTitleChangeNotify(notify.Type())
		}
		return g.GroupConcernConfig.ShouldSendHook(notify)
	default:
		return false
	}
}

func NewGroupConcernConfig(g *concern_manager.GroupConcernConfig) *GroupConcernConfig {
	return &GroupConcernConfig{*g}
}
