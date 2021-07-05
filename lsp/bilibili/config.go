package bilibili

import (
	"github.com/Sora233/DDBOT/concern"
	"github.com/Sora233/DDBOT/lsp/concern_manager"
)

type GroupConcernConfig struct {
	concern_manager.GroupConcernConfig
}

func (g *GroupConcernConfig) AtBeforeHook(notify concern.Notify) (hook *concern_manager.HookResult) {
	hook = new(concern_manager.HookResult)
	switch e := notify.(type) {
	case *ConcernLiveNotify:
		if !e.Living() {
			hook.Reason = "Living() is false"
		} else if !notify.(*ConcernLiveNotify).LiveStatusChanged {
			hook.Reason = "Living() ok but LiveStatusChanged is false"
		} else {
			hook.Pass = true
		}
	case *ConcernNewsNotify:
		hook.Pass = true
	default:
		hook.Reason = "default"
	}
	return
}

func (g *GroupConcernConfig) ShouldSendHook(notify concern.Notify) (hook *concern_manager.HookResult) {
	hook = new(concern_manager.HookResult)
	switch e := notify.(type) {
	case *ConcernLiveNotify:
		if e.Living() {
			if e.LiveStatusChanged {
				// 上播了
				hook.Pass = true
				return
			}
			if e.LiveTitleChanged {
				// 直播间标题改了，检查改标题推送配置
				hook.PassOrReason(g.GroupConcernNotify.CheckTitleChangeNotify(notify.Type()), "CheckTitleChangeNotify is false")
				return
			}
		} else {
			if e.LiveStatusChanged {
				// 下播了，检查下播推送配置
				hook.PassOrReason(g.GroupConcernNotify.CheckOfflineNotify(notify.Type()), "CheckOfflineNotify is false")
				return
			}
		}
		return g.GroupConcernConfig.ShouldSendHook(notify)
	case *ConcernNewsNotify:
		hook.Pass = true
		return
	}
	return g.GroupConcernConfig.ShouldSendHook(notify)
}

func (g *GroupConcernConfig) NewsFilterHook(notify concern.Notify) (hook *concern_manager.HookResult) {
	hook = new(concern_manager.HookResult)
	switch notify.(type) {
	case *ConcernLiveNotify:
		hook.Pass = true
		return
	case *ConcernNewsNotify:
		switch g.GroupConcernFilter.Type {
		case concern_manager.FilterTypeText:
			hook.Pass = true
		case concern_manager.FilterTypeType, concern_manager.FilterTypeNotType:
			hook.Pass = true
		default:
			hook.Reason = "unknown filter type"
		}
		return
	default:
		hook.Reason = "unknown notify type"
		return
	}
}

func NewGroupConcernConfig(g *concern_manager.GroupConcernConfig) *GroupConcernConfig {
	return &GroupConcernConfig{*g}
}
