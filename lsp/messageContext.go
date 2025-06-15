package lsp

import (
	"github.com/LagrangeDev/LagrangeGo/message"
	"github.com/sirupsen/logrus"

	"github.com/Sora233/DDBOT/v2/lsp/mmsg"
)

type MessageContext struct {
	ReplyFunc             func(m *mmsg.MSG) interface{}
	SendFunc              func(m *mmsg.MSG) interface{}
	NoPermissionReplyFunc func() interface{}
	DisabledReply         func() interface{}
	GlobalDisabledReply   func() interface{}
	Lsp                   *Lsp
	Log                   *logrus.Entry
	Target                mmsg.Target
	Sender                *message.Sender
}

func (c *MessageContext) TextSend(text string) interface{} {
	return c.SendFunc(mmsg.NewText(text))
}

func (c *MessageContext) TextReply(text string) interface{} {
	return c.ReplyFunc(mmsg.NewText(text))
}

func (c *MessageContext) Reply(m *mmsg.MSG) interface{} {
	return c.ReplyFunc(m)
}

func (c *MessageContext) Send(m *mmsg.MSG) interface{} {
	return c.SendFunc(m)
}

func (c *MessageContext) NoPermissionReply() interface{} {
	return c.NoPermissionReplyFunc()
}

func (c *MessageContext) GetLog() *logrus.Entry {
	return c.Log
}

func (c *MessageContext) GetTarget() mmsg.Target {
	return c.Target
}

func (c *MessageContext) GetSender() *message.Sender {
	return c.Sender
}

func (c *MessageContext) IsFromPrivate() bool {
	return c.Target.TargetType() == mmsg.TargetPrivate
}

func (c *MessageContext) IsFromGroup() bool {
	return c.Target.TargetType() == mmsg.TargetGroup
}

func NewMessageContext() *MessageContext {
	return new(MessageContext)
}
