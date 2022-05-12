package mmsg

import (
	"github.com/Sora233/DDBOT/lsp/mmsg/mt"
	"github.com/sirupsen/logrus"
)

type MessageSender interface {
	Uin() int64
	Name() string
}

type IMsgCtx interface {
	TextSend(text string) interface{}
	TextReply(text string) interface{}
	Reply(m *MSG) interface{}
	Send(m *MSG) interface{}
	NoPermissionReply() interface{}
	NotImplReply() interface{}
	GetLog() *logrus.Entry
	GetSource() mt.TargetType
	GetSender() MessageSender
}
