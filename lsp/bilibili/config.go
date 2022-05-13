package bilibili

import (
	"fmt"
	"github.com/Mrs4s/MiraiGo/message"
	localdb "github.com/Sora233/DDBOT/lsp/buntdb"
	"github.com/Sora233/DDBOT/lsp/concern"
	"github.com/Sora233/DDBOT/lsp/mmsg/mt"
	"strconv"
	"strings"
)

type ConcernConfig struct {
	concern.IConfig
	concern *Concern
}

func (g *ConcernConfig) Validate() error {
	if !g.GetConcernFilter().Empty() {
		switch g.GetConcernFilter().Type {
		case concern.FilterTypeNotType, concern.FilterTypeType:
			filterByType, err := g.GetConcernFilter().GetFilterByType()
			if err != nil {
				return err
			}
			var invalid = CheckTypeDefine(filterByType.Type)
			if len(invalid) != 0 {
				return fmt.Errorf("未定义的类型：\n%v", strings.Join(invalid, " "))
			}
			return nil
		}
	}

	return g.IConfig.Validate()
}

func (g *ConcernConfig) NotifyBeforeCallback(inotify concern.Notify) {
	if inotify.Type() != News {
		return
	}
	if inotify.GetTarget().GetTargetType() != mt.TargetGroup {
		return
	}
	notify := inotify.(*ConcernNewsNotify)
	switch notify.Card.GetDesc().GetType() {
	case DynamicDescType_WithVideo:
		// 解决联合投稿的时候刷屏
		notify.compactKey = notify.Card.GetDesc().GetBvid()
		err := g.concern.SetTargetCompactMarkIfNotExist(notify.GetTarget(), notify.compactKey)
		if localdb.IsRollback(err) {
			notify.shouldCompact = true
		}
	case DynamicDescType_WithOrigin:
		// 解决一起转发的时候刷屏
		notify.compactKey = notify.Card.GetDesc().GetOrigDyIdStr()
		err := g.concern.SetTargetCompactMarkIfNotExist(notify.GetTarget(), notify.compactKey)
		if localdb.IsRollback(err) {
			notify.shouldCompact = true
		}
	default:
		// 其他动态也设置一下
		notify.compactKey = notify.Card.GetDesc().GetDynamicIdStr()
		err := g.concern.SetTargetCompactMarkIfNotExist(notify.GetTarget(), notify.Card.GetDesc().GetDynamicIdStr())
		if err != nil && !localdb.IsRollback(err) {
			logger.Errorf("SetGroupOriginMarkIfNotExist error %v", err)
		}
	}
}

func (g *ConcernConfig) NotifyAfterCallback(inotify concern.Notify, _msg interface{}) {
	if inotify.Type() != News {
		return
	}
	if inotify.GetTarget().GetTargetType() != mt.TargetGroup {
		return
	}
	msg := _msg.(*message.GroupMessage)
	if msg == nil || msg.Id == -1 {
		return
	}
	notify := inotify.(*ConcernNewsNotify)
	if notify.shouldCompact || len(notify.compactKey) == 0 {
		return
	}
	err := g.concern.SetNotifyMsg(notify.compactKey, msg)
	if err != nil && !localdb.IsRollback(err) {
		notify.Logger().Errorf("set notify msg error %v", err)
	}
}

func (g *ConcernConfig) AtBeforeHook(notify concern.Notify) (hook *concern.HookResult) {
	hook = new(concern.HookResult)
	if g.concern != nil && g.concern.unsafeStart.Load() {
		hook.Reason = "bilibili unsafe start status"
		return
	}
	return g.IConfig.AtBeforeHook(notify)
}

func (g *ConcernConfig) FilterHook(notify concern.Notify) (hook *concern.HookResult) {
	hook = new(concern.HookResult)
	switch n := notify.(type) {
	case *ConcernLiveNotify:
		hook.Pass = true
		return
	case *ConcernNewsNotify:
		// 没设置过滤，pass
		if g.GetConcernFilter().Empty() {
			hook.Pass = true
			return
		}

		logger := notify.Logger().WithField("FilterType", g.GetConcernFilter().Type)
		switch g.GetConcernFilter().Type {
		case concern.FilterTypeType, concern.FilterTypeNotType:
			typeFilter, err := g.GetConcernFilter().GetFilterByType()
			if err != nil {
				logger.WithField("ConcernFilterConfig", g.GetConcernFilter().Config).
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
				switch g.GetConcernFilter().Type {
				case concern.FilterTypeType:
					ok = false
					for _, tp := range convTypes {
						if n.Card.GetDesc().GetType() == tp {
							ok = true
							break
						}
					}
				case concern.FilterTypeNotType:
					ok = true
					for _, tp := range convTypes {
						if n.Card.GetDesc().GetType() == tp {
							ok = false
							break
						}
					}
				}
				if ok {
					logger.Debugf("news notify FilterHook pass")
					hook.Pass = true
				} else {
					logger.WithField("TypeFilter", convTypes).
						Debug("news notify FilterHook filtered")
					hook.Reason = "filtered by TypeFilter"
				}
			}
		default:
			hook = g.IConfig.FilterHook(notify)
		}
		return
	default:
		hook.Reason = "unknown notify type"
		return
	}
}

func NewConcernConfig(g concern.IConfig, c *Concern) *ConcernConfig {
	return &ConcernConfig{g, c}
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
