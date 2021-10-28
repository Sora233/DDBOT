package huya

import (
	"github.com/Sora233/DDBOT/lsp/concern_type"
	"github.com/Sora233/DDBOT/lsp/registry"
)

func init() {
	registry.RegisterConcernManager(NewConcern(registry.GetNotifyChan()), []concern_type.Type{Live})
}
