package mmsg

import (
	"github.com/LagrangeDev/LagrangeGo/message"
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
