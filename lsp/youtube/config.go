package youtube

import (
	"github.com/Sora233/DDBOT/concern"
	concern2 "github.com/Sora233/DDBOT/lsp/concern"
	"github.com/Sora233/DDBOT/lsp/concern_manager"
)

type GroupConcernConfig struct {
	concern2.GroupConcernConfig
}

func (g *GroupConcernConfig) AtBeforeHook(notify concern.Notify) (hook *concern_manager.HookResult) {
	hook = new(concern_manager.HookResult)
	switch notify.Type() {
	case concern.YoutubeLive:
		e := notify.(*ConcernNotify)
		if !e.IsLiving() {
			hook.Reason = "IsLiving() is false"
			return
		} else {
			hook.PassOrReason(e.LiveStatusChanged, "LiveStatusChanged is false")
			return
		}
	case concern.YoutubeVideo:
		hook.Pass = true
		return
	}
	return g.GroupConcernConfig.AtBeforeHook(notify)
}

func (g *GroupConcernConfig) ShouldSendHook(notify concern.Notify) (hook *concern_manager.HookResult) {
	hook = new(concern_manager.HookResult)
	switch notify.(type) {
	case *ConcernNotify:
		hook.Pass = true
		return
	}
	return g.GroupConcernConfig.ShouldSendHook(notify)
}

func NewGroupConcernConfig(g *concern2.GroupConcernConfig) *GroupConcernConfig {
	return &GroupConcernConfig{*g}
}
