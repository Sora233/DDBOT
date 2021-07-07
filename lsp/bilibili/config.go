package bilibili

import (
	"fmt"
	"github.com/Sora233/DDBOT/concern"
	"github.com/Sora233/DDBOT/lsp/concern_manager"
	"github.com/Sora233/DDBOT/utils"
	"strconv"
	"strings"
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
					hook.Reason = "TextFilter All pattern match failed"
				}
			}
		case concern_manager.FilterTypeType, concern_manager.FilterTypeNotType:
			typeFilter, err := g.GroupConcernFilter.GetFilterByType()
			if err != nil {
				logger.WithField("GroupConcernFilterConfig", g.GroupConcernFilter.Config).
					Errorf("get type filter error %v", err)
				hook.Pass = true
			} else {
				var originSize = len(n.Cards)
				var convTypes []DynamicDescType
				var filteredCards []*Card
				for _, tp := range typeFilter.Type {
					if types, _ := PredefinedType[tp]; types != nil {
						convTypes = append(convTypes, types...)
					} else {
						if t, err := strconv.ParseInt(tp, 10, 64); err == nil {
							convTypes = append(convTypes, DynamicDescType(t))
						}
					}
				}

				for _, card := range n.Cards {
					var ok bool
					switch g.GroupConcernFilter.Type {
					case concern_manager.FilterTypeType:
						ok = false
						for _, tp := range convTypes {
							if card.GetDesc().GetType() == tp {
								ok = true
								break
							}
						}
					case concern_manager.FilterTypeNotType:
						ok = true
						for _, tp := range convTypes {
							if card.GetDesc().GetType() == tp {
								ok = false
								break
							}
						}
					}
					if ok {
						filteredCards = append(filteredCards, card)
					} else {
						logger.WithField("CardType", card.GetDesc().GetType()).
							WithField("DynamicId", card.GetDesc().GetDynamicIdStr()).
							WithField("TypeFilter", convTypes).
							Debug("Card Filtered by typeFilter")
					}
				}
				n.Cards = filteredCards
				hook.PassOrReason(len(n.Cards) > 0, fmt.Sprintf("All %v cards are filtered", originSize))
				if hook.Pass {
					logger.Debugf("NewsFilterHook done, size %v -> %v", originSize, len(n.Cards))
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
	Yinpin        = "音频"
	Zhibofenxiang = "直播分享"
)

var PredefinedType = map[string][]DynamicDescType{
	Zhuanlan:      {DynamicDescType_WithPost},
	Zhuanfa:       {DynamicDescType_WithOrigin},
	Tougao:        {DynamicDescType_WithVideo},
	Wenzi:         {DynamicDescType_TextOnly},
	Tupian:        {DynamicDescType_WithImage},
	Yinpin:        {DynamicDescType_WithMusic},
	Zhibofenxiang: {DynamicDescType_WithLive, DynamicDescType_WithLiveV2},
}
