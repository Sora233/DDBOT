package example_concern

import (
	"github.com/Sora233/DDBOT/lsp/concern"
	"math/rand"
)

type GroupConcernConfig struct {
	concern.IConfig
}

func (g *GroupConcernConfig) FilterHook(concern.Notify) *concern.HookResult {
	hook := new(concern.HookResult)
	hook.PassOrReason(rand.Int()%2 == 0, "random FilterHook failed")
	return hook
}

func NewGroupConcernConfig(g concern.IConfig) *GroupConcernConfig {
	return &GroupConcernConfig{g}
}
