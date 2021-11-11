package youtube

import (
	"github.com/Sora233/DDBOT/lsp/concern"
)

func init() {
	concern.RegisterConcernManager(NewConcern(concern.GetNotifyChan()))
}
