package concern

import (
	"golang.org/x/exp/constraints"

	"github.com/Sora233/DDBOT/v2/lsp/concern_type"
)

type AtSomeone[UT constraints.Integer] struct {
	Ctype  concern_type.Type `json:"ctype"`
	AtList []UT              `json:"at_list"`
}

// GroupConcernAtConfig @配置
type GroupConcernAtConfig[UT constraints.Integer] struct {
	AtAll     concern_type.Type `json:"at_all"`
	AtSomeone []*AtSomeone[UT]  `json:"at_someone"`
}

func (g *GroupConcernAtConfig[UT]) CheckAtAll(ctype concern_type.Type) bool {
	if g == nil {
		return false
	}
	return g.AtAll.ContainAll(ctype)
}

func (g *GroupConcernAtConfig[UT]) GetAtSomeoneList(ctype concern_type.Type) []UT {
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

func (g *GroupConcernAtConfig[UT]) SetAtSomeoneList(ctype concern_type.Type, ids []UT) {
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
		g.AtSomeone = append(g.AtSomeone, &AtSomeone[UT]{
			Ctype:  ctype,
			AtList: ids,
		})
	}
}

func (g *GroupConcernAtConfig[UT]) MergeAtSomeoneList(ctype concern_type.Type, ids []UT) {
	if g == nil {
		return
	}
	var found = false
	for _, at := range g.AtSomeone {
		if at.Ctype.ContainAll(ctype) {
			found = true
			var qqSet = make(map[UT]bool)
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
		g.AtSomeone = append(g.AtSomeone, &AtSomeone[UT]{
			Ctype:  ctype,
			AtList: ids,
		})
	}
}

func (g *GroupConcernAtConfig[UT]) RemoveAtSomeoneList(ctype concern_type.Type, ids []UT) {
	if g == nil {
		return
	}
	for _, at := range g.AtSomeone {
		if at.Ctype.ContainAll(ctype) {
			var qqSet = make(map[UT]bool)
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

func (g *GroupConcernAtConfig[UT]) ClearAtSomeoneList(ctype concern_type.Type) {
	if g == nil {
		return
	}
	var newList []*AtSomeone[UT]
	for _, at := range g.AtSomeone {
		if at.Ctype.ContainAll(ctype) {
			continue
		}
		newList = append(newList, at)
	}
	g.AtSomeone = newList
}
