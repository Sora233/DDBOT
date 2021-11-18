package bilibili

import (
	"github.com/Sora233/DDBOT/lsp/concern"
)

func init() {
	concern.RegisterConcern(NewConcern(concern.GetNotifyChan()))
}
