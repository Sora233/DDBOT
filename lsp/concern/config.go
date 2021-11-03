package concern

import (
	"strings"
)

type IConfig interface {
	GetGroupConcernAt() *GroupConcernAtConfig
	GetGroupConcernNotify() *GroupConcernNotifyConfig
	GetGroupConcernFilter() *GroupConcernFilterConfig
	ICallback
	Hook
}

// GroupConcernConfig 默认实现了一些逻辑
// 如果 Notify 有实现 NotifyLive , NotifyLiveStatusChanged , NotifyTitleChanged 等，则会使用默认逻辑
type GroupConcernConfig struct {
	DefaultCallback
	GroupConcernAt     GroupConcernAtConfig     `json:"group_concern_at"`
	GroupConcernNotify GroupConcernNotifyConfig `json:"group_concern_notify"`
	GroupConcernFilter GroupConcernFilterConfig `json:"group_concern_filter"`
}

func (g *GroupConcernConfig) FilterHook(notify Notify) *HookResult {
	return HookResultPass
}

func (g *GroupConcernConfig) AtBeforeHook(notify Notify) *HookResult {
	var result = new(HookResult)
	if liveExe, ok := notify.(NotifyLiveExt); ok && liveExe.IsLive() {
		if !liveExe.Living() {
			result.Reason = "Living() is false"
		} else if !liveExe.LiveStatusChanged() {
			result.Reason = "Living() ok but LiveStatusChanged() is false"
		} else {
			result.Pass = true
		}
		return result
	}
	return HookResultPass
}

func (g *GroupConcernConfig) ShouldSendHook(notify Notify) *HookResult {
	var result = new(HookResult)
	if liveExt, ok := notify.(NotifyLiveExt); ok && liveExt.IsLive() {
		if liveExt.Living() {
			if liveExt.LiveStatusChanged() {
				// 上播了
				return HookResultPass
			}
			if liveExt.TitleChanged() {
				// 直播间标题改了，检查改标题推送配置
				result.PassOrReason(
					g.GetGroupConcernNotify().CheckTitleChangeNotify(notify.Type()),
					"CheckTitleChangeNotify is false",
				)
				return result
			}
		} else if liveExt.LiveStatusChanged() {
			// 下播了，检查下播推送配置
			result.PassOrReason(
				g.GetGroupConcernNotify().CheckOfflineNotify(notify.Type()),
				"CheckOfflineNotify is false",
			)
			return result
		}
		return defaultHookResult
	}
	return HookResultPass
}

func (g *GroupConcernConfig) GetGroupConcernAt() *GroupConcernAtConfig {
	return &g.GroupConcernAt
}

func (g *GroupConcernConfig) GetGroupConcernNotify() *GroupConcernNotifyConfig {
	return &g.GroupConcernNotify
}

func (g *GroupConcernConfig) GetGroupConcernFilter() *GroupConcernFilterConfig {
	return &g.GroupConcernFilter
}

func (g *GroupConcernConfig) ToString() string {
	b, e := json.Marshal(g)
	if e != nil {
		panic(e)
	}
	return string(b)
}

func NewGroupConcernConfigFromString(s string) (*GroupConcernConfig, error) {
	var concernConfig *GroupConcernConfig
	decoder := json.NewDecoder(strings.NewReader(s))
	decoder.UseNumber()
	err := decoder.Decode(&concernConfig)
	return concernConfig, err
}
