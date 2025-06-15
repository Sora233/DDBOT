package mmsg

import (
	"github.com/LagrangeDev/LagrangeGo/message"
	"github.com/sirupsen/logrus"
)

type IMsgCtx interface {
	TextSend(text string) interface{}
	TextReply(text string) interface{}
	Reply(m *MSG) interface{}
	Send(m *MSG) interface{}
	NoPermissionReply() interface{}
	GetLog() *logrus.Entry
	GetTarget() Target
	GetSender() *message.Sender
}
