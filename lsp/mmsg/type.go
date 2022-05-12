package mmsg

import (
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Sora233/DDBOT/lsp/mmsg/mt"
)

const (
	ImageBytes message.ElementType = 10000 + iota
	Typed
	Cut
	At
)

type CustomElement interface {
	PackToElement(target mt.Target) message.IMessageElement
}
