package concern

import (
	"errors"
)

const (
	FilterTypeType    = "type"
	FilterTypeNotType = "not_type"
	FilterTypeText    = "text"
)

type GroupConcernFilterConfigByType struct {
	Type []string `json:"type"`
}

func (g *GroupConcernFilterConfigByType) ToString() string {
	b, _ := json.Marshal(g)
	return string(b)
}

type GroupConcernFilterConfigByText struct {
	Text []string `json:"text"`
}

func (g *GroupConcernFilterConfigByText) ToString() string {
	b, _ := json.Marshal(g)
	return string(b)
}

// GroupConcernFilterConfig 过滤器配置
type GroupConcernFilterConfig struct {
	Type   string `json:"type"`
	Config string `json:"config"`
}

func (g *GroupConcernFilterConfig) Empty() bool {
	return g.Type == "" || g.Config == ""
}

func (g *GroupConcernFilterConfig) GetFilterByType() (*GroupConcernFilterConfigByType, error) {
	if g.Type != FilterTypeType && g.Type != FilterTypeNotType {
		return nil, errors.New("filter type mismatched")
	}
	var result = new(GroupConcernFilterConfigByType)
	err := json.Unmarshal([]byte(g.Config), result)
	return result, err
}

func (g *GroupConcernFilterConfig) GetFilterByText() (*GroupConcernFilterConfigByText, error) {
	if g.Type != FilterTypeText {
		return nil, errors.New("filter type mismatched")
	}
	var result = new(GroupConcernFilterConfigByText)
	err := json.Unmarshal([]byte(g.Config), result)
	return result, err
}
