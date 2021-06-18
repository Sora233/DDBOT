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

type GroupConcernConfig struct {
	defaultHook
	GroupConcernAt     GroupConcernAtConfig     `json:"group_concern_at"`
	GroupConcernNotify GroupConcernNotifyConfig `json:"group_concern_notify"`
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
