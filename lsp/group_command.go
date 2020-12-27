package lsp

import (
	"github.com/Mrs4s/MiraiGo/client"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Sora233/Sora233-MiraiGo/lsp/aliyun"
	"github.com/Sora233/Sora233-MiraiGo/lsp/bilibili"
	"strconv"
	"strings"
)

type LspGroupCommand struct {
	qqClient *client.QQClient
	msg      *message.GroupMessage
	l        *Lsp
}

func NewLspGroupCommand(qqClient *client.QQClient, msg *message.GroupMessage, l *Lsp) *LspGroupCommand {
	return &LspGroupCommand{
		qqClient: qqClient,
		msg:      msg,
		l:        l,
	}
}

func (lgc *LspGroupCommand) Execute() {
	args := strings.Split(lgc.msg.Elements[0].(*message.TextElement).Content, " ")
	switch args[0] {
	case "/Lsp":
		lgc.LspCommand()
	case "/色图":
		lgc.SetuCommand()
	case "/watch":
		lgc.WatchCommand(false)
	case "/unwatch":
		lgc.WatchCommand(true)
	default:
		if lgc.msg.Sender.Uin != lgc.qqClient.Uin {
			lgc.ImageContent()
		}
	}
}

func (lgc *LspGroupCommand) LspCommand() {
	msg := lgc.msg
	qqClient := lgc.qqClient
	groupCode := msg.GroupCode

	log := logger.WithField("GroupCode", groupCode)
	log.Infof("run lsp command")

	sendingMsg := message.NewSendingMessage()
	sendingMsg.Append(message.NewReply(msg))
	sendingMsg.Append(message.NewText("LSP竟然是你"))
	qqClient.SendGroupMessage(groupCode, sendingMsg)
	return
}

func (lgc *LspGroupCommand) SetuCommand() {
	msg := lgc.msg
	qqClient := lgc.qqClient
	groupCode := msg.GroupCode

	log := logger.WithField("GroupCode", groupCode)
	log.Infof("run setu command")

	sendingMsg := message.NewSendingMessage()
	sendingMsg.Append(message.NewReply(msg))
	img, err := lgc.l.getHImage()
	if err != nil {
		log.Errorf("can not get HImage")
		return
	}
	groupImage, err := qqClient.UploadGroupImage(groupCode, img)
	if err != nil {
		log.Errorf("upload group image failed %v", err)
		return
	}
	sendingMsg.Append(groupImage)
	qqClient.SendGroupMessage(groupCode, sendingMsg)
	return

}

func (lgc *LspGroupCommand) WatchCommand(remove bool) {
	msg := lgc.msg
	groupCode := msg.GroupCode

	log := logger.WithField("GroupCode", groupCode)
	log.Infof("run watch command")

	text := msg.Elements[0].(*message.TextElement).Content

	args := strings.Split(text, " ")[1:]

	if len(args) == 0 {
		log.WithField("content", text).Errorf("watch need args")
		lgc.textReply("参数错误 - Usage: /watch [channel name] id [id...]")
		return
	}
	site := "bilibili"
	var ids []int64

	id, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		switch args[0] {
		case "bilibili":
			site = "bilibili"
		default:
			log.WithField("content", text).Errorf("watch args error")
			lgc.textReply("参数错误 - 支持的channel name：['bilibili']")
			return

		}
	} else {
		ids = append(ids, id)
	}
	for _, sid := range args[1:] {
		id, err := strconv.ParseInt(sid, 10, 64)
		if err != nil {
			log.WithField("content", text).Errorf("watch args error")
			lgc.textReply("参数错误\n未知的id：" + sid)
			return
		}
		ids = append(ids, id)
	}

	if site == "bilibili" {
		for _, id := range ids {
			if err := lgc.l.BilibiliConcern.Add(groupCode, id, bilibili.ConcernLive); err != nil {
				log.WithField("mid", id).Errorf("watch error %v", err)
				return
			}
			log.WithField("mid", id).Debugf("watch success")
		}
	}
	if remove {
		lgc.textReply("unwatch done")
	} else {
		lgc.textReply("watch done")
	}
}

func (lgc *LspGroupCommand) ImageContent() {
	msg := lgc.msg
	qqClient := lgc.qqClient

	groupCode := msg.GroupCode

	log := logger.WithField("GroupCode", groupCode)

	for _, e := range msg.Elements {
		if e.Type() == message.Image {
			if img, ok := e.(*message.ImageElement); ok {
				rating := lgc.l.checkImage(img)
				if rating == aliyun.SceneSexy {
					sendingMsg := message.NewSendingMessage()
					sendingMsg.Append(message.NewReply(msg))
					sendingMsg.Append(message.NewText("就这"))
					qqClient.SendGroupMessage(groupCode, sendingMsg)
					return
				} else if rating == aliyun.ScenePorn {
					sendingMsg := message.NewSendingMessage()
					sendingMsg.Append(message.NewReply(msg))
					sendingMsg.Append(message.NewText("多发点"))
					qqClient.SendGroupMessage(groupCode, sendingMsg)
					return
				}
			} else {
				log.Error("can not cast element to GroupImageElement")
			}
		}
	}
}

func (lgc *LspGroupCommand) textReply(text string) {
	msg := lgc.msg
	qqClient := lgc.qqClient

	sendingMsg := message.NewSendingMessage()
	sendingMsg.Append(message.NewReply(msg))
	sendingMsg.Append(message.NewText(text))
	qqClient.SendGroupMessage(msg.GroupCode, sendingMsg)
}
