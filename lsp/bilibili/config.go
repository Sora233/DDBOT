package bilibili

import (
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Sora233/DDBOT/concern"
	"github.com/Sora233/DDBOT/lsp/concern_manager"
	"github.com/Sora233/DDBOT/utils"
	"strconv"
	"strings"
)

type GroupConcernConfig struct {
	concern_manager.GroupConcernConfig
}

func (g *GroupConcernConfig) NotifyBeforeCallback(inotify concern.Notify) {
	if inotify.Type() != concern.BilibiliNews {
		return
	}
}

func (g *GroupConcernConfig) NotifyAfterCallback(inotify concern.Notify, message *message.GroupMessage) {
	if inotify.Type() != concern.BilibiliNews {
		return
	}
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
	// 没设置过滤，pass
	if g.GroupConcernFilter.Empty() {
		hook.Pass = true
		return
	}
	switch n := notify.(type) {
	case *ConcernLiveNotify:
		hook.Pass = true
		return
	case *ConcernNewsNotify:
		logger := notify.Logger().WithField("FilterType", g.GroupConcernFilter.Type)
		switch g.GroupConcernFilter.Type {
		case concern_manager.FilterTypeText:
			textFilter, err := g.GroupConcernFilter.GetFilterByText()
			if err != nil {
				logger.WithField("GroupConcernTextFilter", g.GroupConcernFilter.Config).
					Errorf("get text filter error %v", err)
				hook.Pass = true
			} else {
				for _, text := range textFilter.Text {
					if strings.Contains(utils.MsgToString(notify.ToMessage()), text) {
						hook.Pass = true
						break
					}
				}
				if !hook.Pass {
					logger.WithField("TextFilter", textFilter.Text).
						Debug("news notify filtered by textFilter")
					hook.Reason = "TextFilter All pattern match failed"
				} else {
					logger.Debugf("news notify NewsFilterHook pass")
				}
			}
		case concern_manager.FilterTypeType, concern_manager.FilterTypeNotType:
			typeFilter, err := g.GroupConcernFilter.GetFilterByType()
			if err != nil {
				logger.WithField("GroupConcernFilterConfig", g.GroupConcernFilter.Config).
					Errorf("get type filter error %v", err)
				hook.Pass = true
			} else {
				var convTypes []DynamicDescType
				for _, tp := range typeFilter.Type {
					if types, _ := PredefinedType[tp]; types != nil {
						convTypes = append(convTypes, types...)
					} else {
						if t, err := strconv.ParseInt(tp, 10, 32); err == nil {
							convTypes = append(convTypes, DynamicDescType(t))
						}
					}
				}

				var ok bool
				switch g.GroupConcernFilter.Type {
				case concern_manager.FilterTypeType:
					ok = false
					for _, tp := range convTypes {
						if n.Card.GetDesc().GetType() == tp {
							ok = true
							break
						}
					}
				case concern_manager.FilterTypeNotType:
					ok = true
					for _, tp := range convTypes {
						if n.Card.GetDesc().GetType() == tp {
							ok = false
							break
						}
					}
				}
				if ok {
					logger.Debugf("news notify NewsFilterHook pass")
					hook.Pass = true
				} else {
					logger.WithField("TypeFilter", convTypes).
						Debug("news notify NewsFilterHook filtered")
					hook.Reason = "filtered by TypeFilter"
				}
			}
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

const (
	Zhuanlan      = "专栏"
	Zhuanfa       = "转发"
	Tougao        = "投稿"
	Wenzi         = "文字"
	Tupian        = "图片"
	Zhibofenxiang = "直播分享"
)

var PredefinedType = map[string][]DynamicDescType{
	Zhuanlan:      {DynamicDescType_WithPost},
	Zhuanfa:       {DynamicDescType_WithOrigin},
	Tougao:        {DynamicDescType_WithVideo, DynamicDescType_WithMusic, DynamicDescType_WithPost},
	Wenzi:         {DynamicDescType_TextOnly},
	Tupian:        {DynamicDescType_WithImage},
	Zhibofenxiang: {DynamicDescType_WithLive, DynamicDescType_WithLiveV2},
}

func CheckTypeDefine(types []string) (invalid []string) {
	for _, t := range types {
		if PredefinedType[t] != nil {
			continue
		}
		tp, err := strconv.ParseInt(t, 10, 32)
		if err != nil {
			invalid = append(invalid, t)
			continue
		}
		if _, found := DynamicDescType_name[int32(tp)]; tp != 0 && found {
			continue
		}
		invalid = append(invalid, t)
	}
	return
}
