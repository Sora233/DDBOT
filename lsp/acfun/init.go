package acfun

import (
	"github.com/Sora233/DDBOT/lsp/concern"
	"github.com/Sora233/DDBOT/lsp/concern_type"
)

func init() {
	concern.RegisterConcernManager(
		NewConcern(concern.GetNotifyChan()),
		[]concern_type.Type{Live},
	)
}
