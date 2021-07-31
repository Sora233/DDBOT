package msg_marker

import (
	"github.com/Logiase/MiraiGo-Template/bot"
	"github.com/Mrs4s/MiraiGo/client"
	"github.com/Mrs4s/MiraiGo/message"
	"sync"
)

func init() {
	instance = new(marker)
	bot.RegisterModule(instance)
}

type marker struct{}

var instance *marker

func (m *marker) MiraiGoModule() bot.ModuleInfo {
	return bot.ModuleInfo{
		ID:       "sora233.message-readed-marker",
		Instance: instance,
	}
}

func (m *marker) Init() {
}

func (m *marker) PostInit() {
}

func (m *marker) Serve(bot *bot.Bot) {
	bot.OnGroupMessage(func(client *client.QQClient, message *message.GroupMessage) {
		if message.Sender.Uin != client.Uin {
			client.MarkGroupMessageReaded(message.GroupCode, int64(message.Id))
		}
	})
	bot.OnPrivateMessage(func(qqClient *client.QQClient, privateMessage *message.PrivateMessage) {
		if privateMessage.Sender.Uin != qqClient.Uin {
			qqClient.MarkPrivateMessageReaded(privateMessage.Sender.Uin, int64(privateMessage.Id))
		}
	})
}

func (m *marker) Start(bot *bot.Bot) {
}

func (m *marker) Stop(bot *bot.Bot, wg *sync.WaitGroup) {
	wg.Done()
}
