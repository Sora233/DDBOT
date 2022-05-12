package concern

import (
	"github.com/Sora233/DDBOT/lsp/mmsg/mt"
	"github.com/Sora233/DDBOT/utils/msgstringer"
	"strings"
)

// IConfig 定义了Config的通用接口
// TODO 需要一种支持自定义配置的方法，要怎么样做呢
type IConfig interface {
	Validate() error
	GetConcernAt(targetType mt.TargetType) *ConcernAtConfig
	GetConcernNotify(targetType mt.TargetType) *ConcernNotifyConfig
	GetConcernFilter(targetType mt.TargetType) *ConcernFilterConfig
	ICallback
	Hook
}

// ConcernConfig 实现了 IConfig，并附带一些默认逻辑
// 如果 Notify 有实现 NotifyLiveExt，则会使用默认逻辑
type ConcernConfig struct {
	DefaultCallback
	ConcernAtMap     map[mt.TargetType]*ConcernAtConfig     `json:"concern_at_map"`
	ConcernNotifyMap map[mt.TargetType]*ConcernNotifyConfig `json:"concern_notify_map"`
	ConcernFilterMap map[mt.TargetType]*ConcernFilterConfig `json:"concern_filter_map"`
}

// Validate 可以在此自定义config校验，每次对config修改后会在同一个事务中调用，如果返回non-nil，则改动会回滚，此次操作失败
// 默认支持 ConcernNotifyConfig ConcernAtConfig
// ConcernFilterConfig 默认只支持 text
func (g *ConcernConfig) Validate() error {
	for _, tp := range []mt.TargetType{mt.TargetGroup, mt.TargetGuild, mt.TargetPrivate} {
		if !g.GetConcernFilter(tp).Empty() && g.GetConcernFilter(tp).Type != FilterTypeText {
			return ErrConfigNotSupported
		}
	}
	return nil
}

// FilterHook 默认支持filter text配置，其他为Pass，可以重写这个函数实现自定义的过滤
// b站推送使用这个Hook来支持配置动态类型的过滤（过滤转发动态等）
func (g *ConcernConfig) FilterHook(notify Notify) *HookResult {
	target := notify.GetTarget()
	if g.GetConcernFilter(target.GetTargetType()).Empty() {
		return HookResultPass
	}
	logger := notify.Logger().WithField("FilterType", g.GetConcernFilter(target.GetTargetType()).Type)
	switch g.GetConcernFilter(target.GetTargetType()).Type {
	case FilterTypeText:
		textFilter, err := g.GetConcernFilter(target.GetTargetType()).GetFilterByText()
		if err != nil {
			logger.WithField("Content", g.GetConcernFilter(target.GetTargetType()).Config).
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
func (g *ConcernConfig) AtBeforeHook(notify Notify) *HookResult {
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
func (g *ConcernConfig) ShouldSendHook(notify Notify) *HookResult {
	target := notify.GetTarget()
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
					g.GetConcernNotify(target.GetTargetType()).CheckTitleChangeNotify(notify.Type()),
					"CheckTitleChangeNotify is false",
				)
				return result
			}
		} else if liveExt.LiveStatusChanged() {
			// 下播了，检查下播推送配置
			result.PassOrReason(
				g.GetConcernNotify(target.GetTargetType()).CheckOfflineNotify(notify.Type()),
				"CheckOfflineNotify is false",
			)
			return result
		}
		return defaultHookResult
	}
	return HookResultPass
}

// GetConcernAt 返回 ConcernAtConfig，总是返回 non-nil
func (g *ConcernConfig) GetConcernAt(targetType mt.TargetType) *ConcernAtConfig {
	if g.ConcernAtMap == nil {
		g.ConcernAtMap = make(map[mt.TargetType]*ConcernAtConfig)
	}
	var v *ConcernAtConfig
	v = g.ConcernAtMap[targetType]
	if v == nil {
		v = new(ConcernAtConfig)
		g.ConcernAtMap[targetType] = v
	}
	return v
}

// GetConcernNotify 返回 ConcernNotifyConfig，总是返回 non-nil
func (g *ConcernConfig) GetConcernNotify(targetType mt.TargetType) *ConcernNotifyConfig {
	if g.ConcernNotifyMap == nil {
		g.ConcernNotifyMap = make(map[mt.TargetType]*ConcernNotifyConfig)
	}
	var v *ConcernNotifyConfig
	v = g.ConcernNotifyMap[targetType]
	if v == nil {
		v = new(ConcernNotifyConfig)
		g.ConcernNotifyMap[targetType] = v
	}
	return v
}

// GetConcernFilter 返回 ConcernFilterConfig，总是返回 non-nil
func (g *ConcernConfig) GetConcernFilter(targetType mt.TargetType) *ConcernFilterConfig {
	if g.ConcernFilterMap == nil {
		g.ConcernFilterMap = make(map[mt.TargetType]*ConcernFilterConfig)
	}
	var v *ConcernFilterConfig
	v = g.ConcernFilterMap[targetType]
	if v == nil {
		v = new(ConcernFilterConfig)
		g.ConcernFilterMap[targetType] = v
	}
	return v
}

// ToString 将 ConcernConfig 通过json序列化成string
func (g *ConcernConfig) ToString() string {
	b, e := json.Marshal(g)
	if e != nil {
		panic(e)
	}
	return string(b)
}

// NewConcernConfigFromString 从json格式反序列化 ConcernConfig
func NewConcernConfigFromString(s string) (*ConcernConfig, error) {
	var concernConfig *ConcernConfig
	decoder := json.NewDecoder(strings.NewReader(s))
	decoder.UseNumber()
	err := decoder.Decode(&concernConfig)
	return concernConfig, err
}
