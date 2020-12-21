package lsp

import (
	"github.com/Logiase/MiraiGo-Template/bot"
	"github.com/Mrs4s/MiraiGo/client"
	"github.com/Mrs4s/MiraiGo/message"
	"strings"
	"sync"
)

const ModuleName = "me.sora233.lsp"

type lsp struct{}

func (l *lsp) MiraiGoModule() bot.ModuleInfo {
	return bot.ModuleInfo{
		ID:       ModuleName,
		Instance: instance,
	}
}

func (l *lsp) Init() {
}

func (l *lsp) PostInit() {
}

func (l *lsp) Serve(bot *bot.Bot) {
	bot.OnGroupMessage(func(client *client.QQClient, msg *message.GroupMessage) {
		if strings.ToLower(msg.ToString()) == "/lsp" {
			groupCode := msg.GroupCode
			sendingMsg := message.NewSendingMessage()
			sendingMsg.Append(message.NewReply(msg))
			sendingMsg.Append(message.NewText("LSP竟然是你"))
			client.SendGroupMessage(groupCode, sendingMsg)
		}
	})
}

func (l *lsp) Start(bot *bot.Bot) {
}

func (l *lsp) Stop(bot *bot.Bot, wg *sync.WaitGroup) {
	defer wg.Done()
}

var instance *lsp

func init() {
	instance = &lsp{}
	bot.RegisterModule(instance)
}
