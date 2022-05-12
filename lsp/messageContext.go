package lsp

import (
	"github.com/Sora233/DDBOT/lsp/mmsg"
	"github.com/Sora233/DDBOT/lsp/mmsg/mt"
	"github.com/sirupsen/logrus"
)

type sender struct {
	uin  int64
	name string
}

func (s *sender) Uin() int64 {
	return s.uin
}

func (s *sender) Name() string {
	return s.name
}

func NewMessageSender(uin int64, name string) mmsg.MessageSender {
	return &sender{
		uin:  uin,
		name: name,
	}
}

type MessageContext struct {
	ReplyFunc             func(m *mmsg.MSG) interface{}
	SendFunc              func(m *mmsg.MSG) interface{}
	NoPermissionReplyFunc func() interface{}
	DisabledReply         func() interface{}
	GlobalDisabledReply   func() interface{}
	NotImplReplyFunc      func() interface{}
	Lsp                   *Lsp
	Log                   *logrus.Entry
	Source                mt.TargetType
	Sender                mmsg.MessageSender
}

func (c *MessageContext) GetSource() mt.TargetType {
	return c.Source
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

func (c *MessageContext) NotImplReply() interface{} {
	return c.NotImplReplyFunc()
}

func (c *MessageContext) GetLog() *logrus.Entry {
	return c.Log
}

func (c *MessageContext) GetSender() mmsg.MessageSender {
	return c.Sender
}

func (c *MessageContext) IsFromPrivate() bool {
	return c.Source == mt.TargetPrivate
}

func (c *MessageContext) IsFromGroup() bool {
	return c.Source == mt.TargetGroup
}

func (c *MessageContext) IsFromGulid() bool {
	return c.Source == mt.TargetGulid
}

func NewMessageContext() *MessageContext {
	return new(MessageContext)
}
