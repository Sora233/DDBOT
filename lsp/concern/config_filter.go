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

func (c *GroupConcernFilterConfig) GetFilterByType() (*GroupConcernFilterConfigByType, error) {
	if c.Type != FilterTypeType && c.Type != FilterTypeNotType {
		return nil, errors.New("filter type mismatched")
	}
	var result = new(GroupConcernFilterConfigByType)
	err := json.Unmarshal([]byte(c.Config), result)
	return result, err
}

func (c *GroupConcernFilterConfig) GetFilterByText() (*GroupConcernFilterConfigByText, error) {
	if c.Type != FilterTypeText {
		return nil, errors.New("filter type mismatched")
	}
	var result = new(GroupConcernFilterConfigByText)
	err := json.Unmarshal([]byte(c.Config), result)
	return result, err
}
