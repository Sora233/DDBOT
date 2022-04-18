package mmsg

import (
	"github.com/Mrs4s/MiraiGo/message"
)

const (
	ImageBytes message.ElementType = 10000 + iota
	Typed
	Cut
	At
)

type CustomElement interface {
	PackToElement(target Target) message.IMessageElement
}
