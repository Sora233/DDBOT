package concern_manager

import (
	"strings"
)

type GroupConcernConfig struct {
	defaultCallback
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
