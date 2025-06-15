package concern

import (
	"strings"

	"github.com/Sora233/DDBOT/v2/utils/msgstringer"
)

// IConfig 定义了Config的通用接口
// TODO 需要一种支持自定义配置的方法，要怎么样做呢
type IConfig interface {
	Validate() error
	GetGroupConcernAt() *GroupConcernAtConfig[uint32]
	GetGroupConcernNotify() *GroupConcernNotifyConfig
	GetGroupConcernFilter() *GroupConcernFilterConfig
	ICallback
	Hook
}

// GroupConcernConfig 实现了 IConfig，并附带一些默认逻辑
// 如果 Notify 有实现 NotifyLiveExt，则会使用默认逻辑
type GroupConcernConfig struct {
	DefaultCallback
	GroupConcernAt     GroupConcernAtConfig[uint32] `json:"group_concern_at"`
	GroupConcernNotify GroupConcernNotifyConfig     `json:"group_concern_notify"`
	GroupConcernFilter GroupConcernFilterConfig     `json:"group_concern_filter"`
}

// Validate 可以在此自定义config校验，每次对config修改后会在同一个事务中调用，如果返回non-nil，则改动会回滚，此次操作失败
// 默认支持 GroupConcernNotifyConfig GroupConcernAtConfig
// GroupConcernFilterConfig 默认只支持 text
func (g *GroupConcernConfig) Validate() error {
	if !g.GetGroupConcernFilter().Empty() && g.GetGroupConcernFilter().Type != FilterTypeText {
		return ErrConfigNotSupported
	}
	return nil
}

// FilterHook 默认支持filter text配置，其他为Pass，可以重写这个函数实现自定义的过滤
// b站推送使用这个Hook来支持配置动态类型的过滤（过滤转发动态等）
func (g *GroupConcernConfig) FilterHook(notify Notify) *HookResult {
	if g.GetGroupConcernFilter().Empty() {
		return HookResultPass
	}
	logger := notify.Logger().WithField("FilterType", g.GetGroupConcernFilter().Type)
	switch g.GetGroupConcernFilter().Type {
	case FilterTypeText:
		textFilter, err := g.GetGroupConcernFilter().GetFilterByText()
		if err != nil {
			logger.WithField("Content", g.GetGroupConcernFilter().Config).
				Errorf("GetFilterByText() error %v", err)
		} else {
			var hook = new(HookResult)
			msgString := msgstringer.MsgToString(notify.ToMessage().Elements())
			for _, text := range textFilter.Text {
				if strings.Contains(msgString, text) {
					hook.Pass = true
					break
				}
			}
			if !hook.Pass {
				logger.WithField("TextFilter", textFilter.Text).
					Debug("news notify filtered by textFilter")
				hook.Reason = "TextFilter All pattern match failed"
			} else {
				logger.Debugf("news notify FilterHook pass")
			}
			return hook
		}
	}
	return HookResultPass
}

// AtBeforeHook 默认为Pass
// 当 Notify 实现了 NotifyLiveExt，则仅有上播(Living && LiveStatusChanged 均为true)的时候会Pass
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

// ShouldSendHook 默认为Pass
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

// GetGroupConcernAt 返回 GroupConcernAtConfig，总是返回 non-nil
func (g *GroupConcernConfig) GetGroupConcernAt() *GroupConcernAtConfig[uint32] {
	return &g.GroupConcernAt
}

// GetGroupConcernNotify 返回 GroupConcernNotifyConfig，总是返回 non-nil
func (g *GroupConcernConfig) GetGroupConcernNotify() *GroupConcernNotifyConfig {
	return &g.GroupConcernNotify
}

// GetGroupConcernFilter 返回 GroupConcernFilterConfig，总是返回 non-nil
func (g *GroupConcernConfig) GetGroupConcernFilter() *GroupConcernFilterConfig {
	return &g.GroupConcernFilter
}

// ToString 将 GroupConcernConfig 通过json序列化成string
func (g *GroupConcernConfig) ToString() string {
	b, e := json.Marshal(g)
	if e != nil {
		panic(e)
	}
	return string(b)
}

// NewGroupConcernConfigFromString 从json格式反序列化 GroupConcernConfig
func NewGroupConcernConfigFromString(s string) (*GroupConcernConfig, error) {
	var concernConfig *GroupConcernConfig
	decoder := json.NewDecoder(strings.NewReader(s))
	decoder.UseNumber()
	err := decoder.Decode(&concernConfig)
	return concernConfig, err
}
