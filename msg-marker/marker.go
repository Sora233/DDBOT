package msg_marker

import (
	"github.com/Mrs4s/MiraiGo/client"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Sora233/MiraiGo-Template/bot"
	"github.com/Sora233/MiraiGo-Template/config"
	"github.com/Sora233/MiraiGo-Template/utils"
	"sync"
)

func init() {
	instance = new(marker)
	bot.RegisterModule(instance)
}

const moduleId = "sora233.message-read-marker"

type marker struct{}

var instance *marker

var logger = utils.GetModuleLogger(moduleId)

func (m *marker) MiraiGoModule() bot.ModuleInfo {
	return bot.ModuleInfo{
		ID:       moduleId,
		Instance: instance,
	}
}

func (m *marker) Init() {
}

func (m *marker) PostInit() {
}

func (m *marker) Serve(bot *bot.Bot) {
	if config.GlobalConfig.GetBool("message-marker.disable") {
		logger.Debug("自动已读被禁用")
		return
	}
	logger.Debug("自动已读已开启")
	bot.GroupMessageEvent.Subscribe(func(client *client.QQClient, message *message.GroupMessage) {
		if message.Sender.Uin != client.Uin {
			client.MarkGroupMessageReaded(message.GroupCode, int64(message.Id))
		}
	})
	bot.PrivateMessageEvent.Subscribe(func(qqClient *client.QQClient, privateMessage *message.PrivateMessage) {
		if privateMessage.Sender.Uin != qqClient.Uin {
			qqClient.MarkPrivateMessageReaded(privateMessage.Sender.Uin, int64(privateMessage.Time))
		}
	})
}

func (m *marker) Start(bot *bot.Bot) {
}

func (m *marker) Stop(bot *bot.Bot, wg *sync.WaitGroup) {
	wg.Done()
}
