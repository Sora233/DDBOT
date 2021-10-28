package lsp

import (
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Sora233/DDBOT/lsp/msg"
	"github.com/sirupsen/logrus"
)

type MessageContext struct {
	ReplyFunc             func(m *msg.MSG) interface{}
	SendFunc              func(m *msg.MSG) interface{}
	NoPermissionReplyFunc func() interface{}
	DisabledReply         func() interface{}
	GlobalDisabledReply   func() interface{}
	Lsp                   *Lsp
	Log                   *logrus.Entry
	Target                msg.Target
	Sender                *message.Sender
}

func (c *MessageContext) TextSend(text string) interface{} {
	return c.SendFunc(msg.NewText(text))
}

func (c *MessageContext) TextReply(text string) interface{} {
	return c.ReplyFunc(msg.NewText(text))
}

func (c *MessageContext) Reply(m *msg.MSG) interface{} {
	return c.ReplyFunc(m)
}

func (c *MessageContext) Send(m *msg.MSG) interface{} {
	return c.SendFunc(m)
}

func (c *MessageContext) NoPermissionReply() interface{} {
	return c.NoPermissionReplyFunc()
}

func (c *MessageContext) GetLog() *logrus.Entry {
	return c.Log
}

func (c *MessageContext) GetTarget() msg.Target {
	return c.Target
}

func (c *MessageContext) GetSender() *message.Sender {
	return c.Sender
}

func (c *MessageContext) IsFromPrivate() bool {
	return c.Target.TargetType() == msg.TargetPrivate
}

func (c *MessageContext) IsFromGroup() bool {
	return c.Target.TargetType() == msg.TargetGroup
}

func NewMessageContext() *MessageContext {
	return new(MessageContext)
}
