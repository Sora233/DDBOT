package youtube

import (
	"github.com/Sora233/DDBOT/lsp/concern"
)

type GroupConcernConfig struct {
	concern.GroupConcernConfig
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
	return g.GroupConcernConfig.AtBeforeHook(notify)
}

func (g *GroupConcernConfig) ShouldSendHook(notify concern.Notify) (hook *concern.HookResult) {
	hook = new(concern.HookResult)
	switch notify.(type) {
	case *ConcernNotify:
		hook.Pass = true
		return
	}
	return g.GroupConcernConfig.ShouldSendHook(notify)
}

func NewGroupConcernConfig(g *concern.GroupConcernConfig) *GroupConcernConfig {
	return &GroupConcernConfig{*g}
}
