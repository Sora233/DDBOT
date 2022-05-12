package douyu

import (
	"github.com/Sora233/DDBOT/lsp/concern"
)

type GroupConcernConfig struct {
	concern.IConfig
}

func NewConcernConfig(g concern.IConfig) *GroupConcernConfig {
	return &GroupConcernConfig{g}
}
