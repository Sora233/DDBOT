package lsp

import (
	"github.com/Logiase/MiraiGo-Template/bot"
	"github.com/Logiase/MiraiGo-Template/utils"
	"github.com/Mrs4s/MiraiGo/client"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Sora233/Sora233-MiraiGo/lsp/aliyun"
	"strings"
	"sync"
)

const ModuleName = "me.sora233.lsp"

var logger = utils.GetModuleLogger(ModuleName)

type lsp struct{}

func (l *lsp) MiraiGoModule() bot.ModuleInfo {
	return bot.ModuleInfo{
		ID:       ModuleName,
		Instance: instance,
	}
}

func (l *lsp) Init() {
	//apiKey := config.GlobalConfig.GetString("moderatecontent.apikey")
	//moderate.InitModerate(apiKey)
	aliyun.InitAliyun()
}

func (l *lsp) PostInit() {
}

func (l *lsp) Serve(bot *bot.Bot) {
	bot.OnGroupMessage(func(qqClient *client.QQClient, msg *message.GroupMessage) {
		groupCode := msg.GroupCode
		if strings.ToLower(msg.ToString()) == "/lsp" {
			sendingMsg := message.NewSendingMessage()
			sendingMsg.Append(message.NewReply(msg))
			sendingMsg.Append(message.NewText("LSP竟然是你"))
			qqClient.SendGroupMessage(groupCode, sendingMsg)
			return
		}
		for _, e := range msg.Elements {
			if e.Type() == message.Image {
				if img, ok := e.(*message.ImageElement); ok {
					rating := l.checkImage(img)
					if rating == aliyun.SceneSexy {
						sendingMsg := message.NewSendingMessage()
						sendingMsg.Append(message.NewReply(msg))
						sendingMsg.Append(message.NewText("就这"))
						qqClient.SendGroupMessage(groupCode, sendingMsg)
						return
					} else if rating == aliyun.ScenePorn {
						sendingMsg := message.NewSendingMessage()
						sendingMsg.Append(message.NewReply(msg))
						sendingMsg.Append(message.NewText("再来点"))
						qqClient.SendGroupMessage(groupCode, sendingMsg)
						return
					}
				} else {
					logger.Error("can not cast element to GroupImageElement")
				}
			}
		}
	})

	bot.OnPrivateMessage(func(qqClient *client.QQClient, msg *message.PrivateMessage) {
		for _, e := range msg.Elements {
			if e.Type() == message.Image {
				if img, ok := e.(*message.ImageElement); ok {
					rating := l.checkImage(img)
					if rating == aliyun.SceneSexy {
						sendingMsg := message.NewSendingMessage()
						sendingMsg.Append(message.NewText("就这"))
						qqClient.SendPrivateMessage(msg.Sender.Uin, sendingMsg)
						return
					} else if rating == aliyun.ScenePorn {
						sendingMsg := message.NewSendingMessage()
						sendingMsg.Append(message.NewText("再来点"))
						qqClient.SendPrivateMessage(msg.Sender.Uin, sendingMsg)
						return
					}
				} else {
					logger.Error("can not cast element to GroupImageElement")
				}
			}
		}
	})
}

//func (l *lsp) checkImage(img *message.ImageElement) int {
//	logger.WithField("image_url", img.Url).Info("image here")
//	resp, err := moderate.Anime(img.Url)
//	if err != nil {
//		logger.Errorf("moderate request error %v", err)
//	} else if resp.ErrorCode != 0 {
//		logger.Errorf("moderate response code %v, msg %v", resp.ErrorCode, resp.Error)
//	}
//
//	logger.WithField("teen", resp.Predictions.Teen).
//		WithField("everyone", resp.Predictions.Everyone).
//		WithField("adult", resp.Predictions.Adult).
//		Debug("detect done")
//	return resp.RatingIndex
//}
func (l *lsp) checkImage(img *message.ImageElement) string {
	logger.WithField("image_url", img.Url).Info("image here")
	resp, err := aliyun.Audit(img.Url)
	if err != nil {
		logger.Errorf("moderate request error %v", err)
		return ""
	} else if resp.Data.Results[0].Code != 0 {
		logger.Errorf("aliyun response code %v, msg %v", resp.Data.Results[0].Code, resp.Data.Results[0].Message)
		return ""
	}

	logger.WithField("label", resp.Data.Results[0].SubResults[0].Label).
		WithField("rate", resp.Data.Results[0].SubResults[0].Rate).
		Debug("detect done")
	return resp.Data.Results[0].SubResults[0].Label
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
