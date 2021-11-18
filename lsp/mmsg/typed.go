package mmsg

import (
	"github.com/Mrs4s/MiraiGo/message"
)

// TypedElement 根据TargetType选择不同的element，不解决循环问题，使用不当可能导致堆栈溢出
// 可以同时设置 OnGroup 和 OnPrivate ，发送时会根据目标自动选择
// 如果只设置了一个，发送另一个时会返回 nil ，即这里什么也不发送
type TypedElement struct {
	//E map[TargetType]message.IMessageElement
	privateE message.IMessageElement
	groupE   message.IMessageElement
}

func NewTypedElement() *TypedElement {
	return new(TypedElement)
}

func NewGroupElement(e message.IMessageElement) *TypedElement {
	return NewTypedElement().OnGroup(e)
}

func NewPrivateElement(e message.IMessageElement) *TypedElement {
	return NewTypedElement().OnPrivate(e)
}

func (t *TypedElement) Type() message.ElementType {
	return Typed
}

func (t *TypedElement) PackToElement(target Target) message.IMessageElement {
	if t.privateE == nil && t.groupE == nil {
		return nil
	}
	var e message.IMessageElement
	switch target.TargetType() {
	case TargetPrivate:
		e = t.privateE
	case TargetGroup:
		e = t.groupE
	}
	if e == nil {
		return e
	}
	if ce, ok := e.(CustomElement); ok {
		return ce.PackToElement(target)
	}
	return e
}

func (t *TypedElement) OnPrivate(e message.IMessageElement) *TypedElement {
	if t == e {
		panic("TypedElement can not type self")
	}
	t.privateE = e
	return t
}

func (t *TypedElement) OnGroup(e message.IMessageElement) *TypedElement {
	if t == e {
		panic("TypedElement can not type self")
	}
	t.groupE = e
	return t
}
