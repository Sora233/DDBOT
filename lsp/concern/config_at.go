package concern

type AtSomeone struct {
	Ctype  Type    `json:"ctype"`
	AtList []int64 `json:"at_list"`
}

// GroupConcernAtConfig @配置
type GroupConcernAtConfig struct {
	AtAll     Type         `json:"at_all"`
	AtSomeone []*AtSomeone `json:"at_someone"`
}

func (g *GroupConcernAtConfig) CheckAtAll(ctype Type) bool {
	if g == nil {
		return false
	}
	return g.AtAll.ContainAll(ctype)
}

func (g *GroupConcernAtConfig) GetAtSomeoneList(ctype Type) []int64 {
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

func (g *GroupConcernAtConfig) SetAtSomeoneList(ctype Type, ids []int64) {
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

func (g *GroupConcernAtConfig) MergeAtSomeoneList(ctype Type, ids []int64) {
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

func (g *GroupConcernAtConfig) RemoveAtSomeoneList(ctype Type, ids []int64) {
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

func (g *GroupConcernAtConfig) ClearAtSomeoneList(ctype Type) {
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
