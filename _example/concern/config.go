package example_concern

import (
	"github.com/Sora233/DDBOT/lsp/concern"
	"math/rand"
)

// GroupConcernConfig 创建一个新结构，准备重写 FilterHook
type GroupConcernConfig struct {
	concern.IConfig
}

// FilterHook 自定义推送过滤逻辑
func (g *GroupConcernConfig) FilterHook(concern.Notify) *concern.HookResult {
	hook := new(concern.HookResult)
	hook.PassOrReason(rand.Int()%2 == 0, "random FilterHook failed")
	return hook
}

// NewGroupConcernConfig 创建一个新的 GroupConcernConfig
func NewGroupConcernConfig(g concern.IConfig) *GroupConcernConfig {
	return &GroupConcernConfig{g}
}
