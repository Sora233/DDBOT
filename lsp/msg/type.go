package msg

import (
	"github.com/Mrs4s/MiraiGo/client"
	"github.com/Mrs4s/MiraiGo/message"
)

const (
	ImageBytes message.ElementType = 10000 + iota
	Typed
)

type CustomElement interface {
	PackToElement(client *client.QQClient, target Target) message.IMessageElement
}
