package bilibili

import (
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Sora233/DDBOT/concern"
	localdb "github.com/Sora233/DDBOT/lsp/buntdb"
	"github.com/Sora233/DDBOT/lsp/concern_manager"
	"github.com/Sora233/DDBOT/utils"
	"strconv"
	"strings"
	"sync/atomic"
)

type GroupConcernConfig struct {
	concern_manager.GroupConcernConfig
	Concern *Concern
}

func (g *GroupConcernConfig) NotifyBeforeCallback(inotify concern.Notify) {
	if inotify.Type() != concern.BilibiliNews {
		return
	}
	notify := inotify.(*ConcernNewsNotify)
	switch notify.Card.GetDesc().GetType() {
	case DynamicDescType_WithVideo:
		videoCard, err := notify.Card.GetCardWithVideo()
		if err != nil {
			return
		}
		videoOrigin := videoCard.GetOrigin()
		if videoOrigin == nil {
			return
		}
		notify.compactKey = videoOrigin.GetBvid()
		// 解决联合投稿的时候刷屏
		err = g.Concern.SetGroupCompactMarkIfNotExist(notify.GetGroupCode(), videoOrigin.GetBvid())
		if localdb.IsRollback(err) {
			notify.shouldCompact = true
		}
	case DynamicDescType_WithOrigin:
		// 解决一起转发的时候刷屏
		origDyId := notify.Card.GetDesc().GetOrigDyIdStr()
		notify.compactKey = origDyId
		err := g.Concern.SetGroupCompactMarkIfNotExist(notify.GetGroupCode(), origDyId)
		if localdb.IsRollback(err) {
			notify.shouldCompact = true
		}
	default:
		// 其他动态也设置一下
		notify.compactKey = notify.Card.GetDesc().GetDynamicIdStr()
		err := g.Concern.SetGroupCompactMarkIfNotExist(notify.GetGroupCode(), notify.Card.GetDesc().GetDynamicIdStr())
		if err != nil && !localdb.IsRollback(err) {
			logger.Errorf("SetGroupOriginMarkIfNotExist error %v", err)
		}
	}
}

func (g *GroupConcernConfig) NotifyAfterCallback(inotify concern.Notify, msg *message.GroupMessage) {
	if inotify.Type() != concern.BilibiliNews || msg == nil || msg.Id == -1 {
		return
	}
	notify := inotify.(*ConcernNewsNotify)
	if notify.shouldCompact || len(notify.compactKey) == 0 {
		return
	}
	err := g.Concern.SetNotifyMsg(notify.compactKey, msg)
	if err != nil && !localdb.IsRollback(err) {
		notify.Logger().Errorf("set notify msg error %v", err)
	}
}

func (g *GroupConcernConfig) AtBeforeHook(notify concern.Notify) (hook *concern_manager.HookResult) {
	hook = new(concern_manager.HookResult)
	if g.Concern != nil && atomic.LoadInt32(&g.Concern.unsafeStart) != 0 {
		hook.Reason = "unsafe start status"
		return
	}
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
	switch n := notify.(type) {
	case *ConcernLiveNotify:
		hook.Pass = true
		return
	case *ConcernNewsNotify:
		// 2021-08-15 发现好像是系统推荐的直播间，非人为操作，选择不推送
		if n.Card.GetDesc().GetType() == DynamicDescType_WithLiveV2 {
			hook.Reason = "WithLiveV2 news notify filtered"
			return
		}

		// 没设置过滤，pass
		if g.GroupConcernFilter.Empty() {
			hook.Pass = true
			return
		}

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

func NewGroupConcernConfig(g *concern_manager.GroupConcernConfig, c *Concern) *GroupConcernConfig {
	return &GroupConcernConfig{*g, c}
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
