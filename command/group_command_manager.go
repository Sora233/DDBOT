package command

import (
	"github.com/Mrs4s/MiraiGo/client"
	"github.com/Mrs4s/MiraiGo/message"
)

type GroupCommand interface {
	Usage() string
	Execute(*client.QQClient, *message.GroupMessage) error
}

type GroupCommandManager interface {
	Execute(*message.GroupMessage) error
	Register(string, GroupCommand) error
}
