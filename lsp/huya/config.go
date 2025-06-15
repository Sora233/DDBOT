package huya

import (
	"github.com/Sora233/DDBOT/v2/lsp/concern"
)

type GroupConcernConfig struct {
	concern.IConfig
}

func NewGroupConcernConfig(g concern.IConfig) *GroupConcernConfig {
	return &GroupConcernConfig{g}
}
