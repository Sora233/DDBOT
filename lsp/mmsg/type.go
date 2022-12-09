package mmsg

import (
	"github.com/Mrs4s/MiraiGo/message"
)

const (
	ImageBytes message.ElementType = 10000 + iota
	Typed
	Cut
	At
	Poke
)

type CustomElement interface {
	PackToElement(target Target) message.IMessageElement
}
