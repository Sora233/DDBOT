package msg

import (
	"github.com/Mrs4s/MiraiGo/client"
	"github.com/Mrs4s/MiraiGo/message"
)

// TypedElement 根据TargetType选择不同的element，不解决循环问题，使用不当可能导致堆栈溢出
type TypedElement struct {
	E map[TargetType]message.IMessageElement
}

func NewTypedElement() *TypedElement {
	return new(TypedElement)
}

func NewGroupElement(e message.IMessageElement) *TypedElement {
	return NewTypedElement().OnGroup(e)
}

func NewPrivateElement(e message.IMessageElement) *TypedElement {
	return NewTypedElement().OnGroup(e)
}

func (t *TypedElement) Type() message.ElementType {
	return Typed
}

func (t *TypedElement) PackToElement(client *client.QQClient, target Target) message.IMessageElement {
	if t.E == nil {
		return nil
	}
	e := t.E[target.TargetType()]
	if e == nil {
		return e
	}
	if ce, ok := e.(CustomElement); ok {
		return ce.PackToElement(client, target)
	}
	return e
}

func (t *TypedElement) OnPrivate(e message.IMessageElement) *TypedElement {
	if t == e {
		panic("TypedElement can not type self")
	}
	if t.E == nil {
		t.E = make(map[TargetType]message.IMessageElement)
	}
	t.E[TargetPrivate] = e
	return t
}

func (t *TypedElement) OnGroup(e message.IMessageElement) *TypedElement {
	if t.E == nil {
		t.E = make(map[TargetType]message.IMessageElement)
	}
	t.E[TargetGroup] = e
	return t
}
