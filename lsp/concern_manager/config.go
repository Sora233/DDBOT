package concern_manager

import (
	"encoding/json"
	"github.com/Sora233/DDBOT/concern"
	"strings"
)

type AtSomeone struct {
	Ctype  concern.Type `json:"ctype"`
	AtList []int64      `json:"at_list"`
}

// GroupConcernAtConfig @配置
type GroupConcernAtConfig struct {
	AtAll     concern.Type `json:"at_all"`
	AtSomeone []*AtSomeone `json:"at_someone"`
}

func (g *GroupConcernAtConfig) CheckAtAll(ctype concern.Type) bool {
	if g == nil {
		return false
	}
	return g.AtAll.ContainAll(ctype)
}

func (g *GroupConcernAtConfig) GetAtSomeoneList(ctype concern.Type) []int64 {
	if g == nil {
		return nil
	}
	for _, at := range g.AtSomeone {
		if at.Ctype.ContainAll(ctype) {
			return at.AtList
		}
	}
	return nil
}

func (g *GroupConcernAtConfig) SetAtSomeoneList(ctype concern.Type, ids []int64) {
	if g == nil {
		return
	}
	var found = false
	for _, at := range g.AtSomeone {
		if at.Ctype.ContainAll(ctype) {
			found = true
			at.AtList = ids
		}
	}
	if !found {
		g.AtSomeone = append(g.AtSomeone, &AtSomeone{
			Ctype:  ctype,
			AtList: ids,
		})
	}
}

func (g *GroupConcernAtConfig) MergeAtSomeoneList(ctype concern.Type, ids []int64) {
	if g == nil {
		return
	}
	var found = false
	for _, at := range g.AtSomeone {
		if at.Ctype.ContainAll(ctype) {
			found = true
			var qqSet = make(map[int64]bool)
			for _, id := range at.AtList {
				qqSet[id] = true
			}
			for _, id := range ids {
				qqSet[id] = true
			}
			at.AtList = nil
			for id := range qqSet {
				at.AtList = append(at.AtList, id)
			}
		}
	}
	if !found {
		g.AtSomeone = append(g.AtSomeone, &AtSomeone{
			Ctype:  ctype,
			AtList: ids,
		})
	}
}

func (g *GroupConcernAtConfig) RemoveAtSomeoneList(ctype concern.Type, ids []int64) {
	if g == nil {
		return
	}
	for _, at := range g.AtSomeone {
		if at.Ctype.ContainAll(ctype) {
			var qqSet = make(map[int64]bool)
			for _, id := range at.AtList {
				qqSet[id] = true
			}
			for _, id := range ids {
				delete(qqSet, id)
			}
			at.AtList = nil
			for id := range qqSet {
				at.AtList = append(at.AtList, id)
			}
		}
	}
}

func (g *GroupConcernAtConfig) ClearAtSomeoneList(ctype concern.Type) {
	if g == nil {
		return
	}
	var newList []*AtSomeone
	for _, at := range g.AtSomeone {
		if at.Ctype.ContainAll(ctype) {
			continue
		}
		newList = append(newList, at)
	}
	g.AtSomeone = newList
}

// GroupConcernNotifyConfig 推送配置
type GroupConcernNotifyConfig struct {
	TitleChangeNotify concern.Type `json:"title_change_notify"`
	OfflineNotify     concern.Type `json:"offline_notify"`
}

func (g *GroupConcernNotifyConfig) CheckTitleChangeNotify(ctype concern.Type) bool {
	return g.TitleChangeNotify.ContainAll(ctype)
}

func (g *GroupConcernNotifyConfig) CheckOfflineNotify(ctype concern.Type) bool {
	return g.OfflineNotify.ContainAll(ctype)
}

// GroupConcernFilterConfig 过滤器配置
type GroupConcernFilterConfig struct {
	Type   string `json:"type"`
	Config string `json:"config"`
}

type GroupConcernConfig struct {
	defaultHook
	GroupConcernAt     GroupConcernAtConfig     `json:"group_concern_at"`
	GroupConcernNotify GroupConcernNotifyConfig `json:"group_concern_notify"`
	GroupConcernFilter GroupConcernFilterConfig `json:"group_concern_filter"`
}

func NewGroupConcernConfigFromString(s string) (*GroupConcernConfig, error) {
	var concernConfig *GroupConcernConfig
	decoder := json.NewDecoder(strings.NewReader(s))
	decoder.UseNumber()
	err := decoder.Decode(&concernConfig)
	return concernConfig, err
}

func (g *GroupConcernConfig) ToString() string {
	b, e := json.Marshal(g)
	if e != nil {
		panic(e)
	}
	return string(b)
}
