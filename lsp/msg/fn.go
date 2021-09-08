package msg

import (
	"github.com/Mrs4s/MiraiGo/client"
	"github.com/Mrs4s/MiraiGo/message"
)

type FnElement struct {
	f func()
}

func NewFnElement() *FnElement {
	return new(FnElement)
}

func (f *FnElement) Type() message.ElementType {
	return Fn
}

func (f *FnElement) PackToElement(client *client.QQClient, target Target) message.IMessageElement {
	return f
}

func (f *FnElement) F(fn func()) *FnElement {
	f.f = fn
	return f
}
