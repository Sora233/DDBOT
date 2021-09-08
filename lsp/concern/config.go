package concern

import (
	"strings"
)

type IConfig interface {
	GetGroupConcernAt() *GroupConcernAtConfig
	GetGroupConcernNotify() *GroupConcernNotifyConfig
	GetGroupConcernFilter() *GroupConcernFilterConfig
}

type GroupConcernConfig struct {
	DefaultCallback
	DefaultHook
	GroupConcernAt     GroupConcernAtConfig     `json:"group_concern_at"`
	GroupConcernNotify GroupConcernNotifyConfig `json:"group_concern_notify"`
	GroupConcernFilter GroupConcernFilterConfig `json:"group_concern_filter"`
}

func (g *GroupConcernConfig) GetGroupConcernAt() *GroupConcernAtConfig {
	return &g.GroupConcernAt
}

func (g *GroupConcernConfig) GetGroupConcernNotify() *GroupConcernNotifyConfig {
	return &g.GroupConcernNotify
}

func (g *GroupConcernConfig) GetGroupConcernFilter() *GroupConcernFilterConfig {
	return &g.GroupConcernFilter
}

func (g *GroupConcernConfig) ToString() string {
	b, e := json.Marshal(g)
	if e != nil {
		panic(e)
	}
	return string(b)
}

func NewGroupConcernConfigFromString(s string) (*GroupConcernConfig, error) {
	var concernConfig *GroupConcernConfig
	decoder := json.NewDecoder(strings.NewReader(s))
	decoder.UseNumber()
	err := decoder.Decode(&concernConfig)
	return concernConfig, err
}
