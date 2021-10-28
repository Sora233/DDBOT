package youtube

import (
	"github.com/Sora233/DDBOT/lsp/concern"
)

type GroupConcernConfig struct {
	concern.IConfig
}

func (g *GroupConcernConfig) AtBeforeHook(notify concern.Notify) (hook *concern.HookResult) {
	hook = new(concern.HookResult)
	switch notify.Type() {
	case Live:
		e := notify.(*ConcernNotify)
		if !e.IsLiving() {
			hook.Reason = "IsLiving() is false"
			return
		} else {
			hook.PassOrReason(e.LiveStatusChanged, "LiveStatusChanged is false")
			return
		}
	case Video:
		hook.Pass = true
		return
	}
	return g.IConfig.AtBeforeHook(notify)
}

func (g *GroupConcernConfig) ShouldSendHook(notify concern.Notify) (hook *concern.HookResult) {
	hook = new(concern.HookResult)
	switch e := notify.(type) {
	case *ConcernNotify:
		if e.IsLive() {
			if e.IsLiving() {
				if e.LiveStatusChanged {
					// 上播了
					hook.Pass = true
					return
				}
				if e.LiveTitleChanged {
					// 直播间标题改了，检查改标题推送配置
					hook.PassOrReason(g.GetGroupConcernNotify().CheckTitleChangeNotify(notify.Type()), "CheckTitleChangeNotify is false")
					return
				}
			} else {
				if e.LiveStatusChanged {
					// 下播了，检查下播推送配置
					hook.PassOrReason(g.GetGroupConcernNotify().CheckOfflineNotify(notify.Type()), "CheckOfflineNotify is false")
					return
				}
			}
		} else {
			hook.Pass = true
			return
		}
	}
	hook.Reason = "nothing changed"
	return
}

func NewGroupConcernConfig(g concern.IConfig) *GroupConcernConfig {
	return &GroupConcernConfig{g}
}
