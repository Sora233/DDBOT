package huya

import (
	"github.com/Sora233/DDBOT/v2/lsp/concern"
)

func init() {
	concern.RegisterConcern(NewConcern(concern.GetNotifyChan()))
}
