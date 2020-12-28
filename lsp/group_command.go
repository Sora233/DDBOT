package lsp

import (
	"fmt"
	"github.com/Mrs4s/MiraiGo/client"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Sora233/Sora233-MiraiGo/lsp/aliyun"
	"github.com/Sora233/Sora233-MiraiGo/lsp/bilibili"
	"github.com/Sora233/Sora233-MiraiGo/lsp/concern"
	"github.com/forestgiant/sliceutil"
	"math/rand"
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
	if text, ok := lgc.msg.Elements[0].(*message.TextElement); ok {
		args := strings.Split(text.Content, " ")
		switch args[0] {
		case "/Lsp":
			lgc.LspCommand()
		case "/色图":
			lgc.SetuCommand()
		case "/watch":
			lgc.WatchCommand(false)
		case "/unwatch":
			lgc.WatchCommand(true)
		case "/list":
			lgc.ListLiving()
		case "/roll":
			lgc.RollCommand()
		default:
		}
	} else {
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
			lgc.textReply("参数错误 - 未知的id：" + sid)
			return
		}
		ids = append(ids, id)
	}

	switch site {
	case "bilibili":
		for _, id := range ids {
			if !remove {
				// watch
				infoResp, err := bilibili.XSpaceAccInfo(id)
				if err != nil {
					log.WithField("mid", id).Error(err)
					lgc.textReply(fmt.Sprintf("查询用户信息失败 %v - %v", id, err))
					continue
				}

				name := infoResp.GetData().GetName()

				if sliceutil.Contains([]int64{491474049}, id) {
					lgc.textReply(fmt.Sprintf("watch失败 - 用户 %v 禁止watch", name))
					continue
				}
				if bilibili.RoomStatus(infoResp.GetData().GetLiveRoom().GetRoomStatus()) == bilibili.RoomStatus_NonExist {
					lgc.textReply(fmt.Sprintf("watch失败 - 用户 %v 暂未开通直播间", name))
					continue
				}
				if err := lgc.l.BilibiliConcern.AddLiveRoom(groupCode, id,
					infoResp.GetData().GetLiveRoom().GetRoomid(),
				); err != nil {

					log.WithField("mid", id).Errorf("watch error %v", err)
					lgc.textReply(fmt.Sprintf("watch失败 - %v", err))
					continue
				}
				log.WithField("mid", id).Debugf("watch success")
				lgc.textReply(fmt.Sprintf("watch成功 - Bilibili用户 %v", name))
			} else {
				// unwatch
				if err := lgc.l.BilibiliConcern.Remove(groupCode, id, concern.BibiliLive); err != nil {
					lgc.textReply(fmt.Sprintf("unwatch失败 - %v", err))
					continue
				} else {
					log.WithField("mid", id).Debugf("unwatch success")
					lgc.textReply("unwatch成功")
				}
			}
		}
	default:
		log.WithField("site", site).Error("unsupported")
		lgc.textReply("未支持的网站")
	}
}

func (lgc *LspGroupCommand) ListLiving() {
	msg := lgc.msg
	groupCode := msg.GroupCode

	log := logger.WithField("GroupCode", groupCode)
	log.Infof("run lisg living command")

	text := msg.Elements[0].(*message.TextElement).Content

	args := strings.Split(text, " ")[1:]

	all := false

	if len(args) >= 1 {
		if args[0] == "all" {
			all = true
		}
	}

	living, err := lgc.l.BilibiliConcern.ListLiving(groupCode, all)
	if err != nil {
		log.Debugf("list living failed %v", err)
		lgc.textReply(fmt.Sprintf("list living 失败 - %v", err))
		return
	}
	listMsg := message.NewSendingMessage()
	for idx, liveInfo := range living {
		if idx != 0 {
			listMsg.Append(message.NewText("\n"))
		}
		notifyMsg := lgc.l.NotifyMessage(lgc.qqClient, liveInfo)
		for _, msg := range notifyMsg {
			listMsg.Append(msg)
		}
	}
	lgc.reply(listMsg)

}

func (lgc *LspGroupCommand) RollCommand() {
	msg := lgc.msg
	groupCode := msg.GroupCode

	log := logger.WithField("GroupCode", groupCode)
	log.Infof("run roll command")

	text := msg.Elements[0].(*message.TextElement).Content

	args := strings.Split(text, " ")[1:]

	var (
		max int64 = 100
		min int64 = 1
		err error
	)
	if len(args) >= 2 {
		lgc.textReply(fmt.Sprintf("多余的参数 - %v", args[1:]))
		return
	}

	if len(args) == 1 {
		rollarg := args[0]
		if strings.Contains(rollarg, "-") {
			rolls := strings.Split(rollarg, "-")
			if len(rolls) != 2 {
				lgc.textReply(fmt.Sprintf("参数解析错误 - %v", rollarg))
				return
			}
			min, err = strconv.ParseInt(rolls[0], 10, 64)
			if err != nil {
				lgc.textReply(fmt.Sprintf("参数解析错误 - %v", rollarg))
				return
			}
			max, err = strconv.ParseInt(rolls[1], 10, 64)
			if err != nil {
				lgc.textReply(fmt.Sprintf("参数解析错误 - %v", rollarg))
				return
			}
		} else {
			max, err = strconv.ParseInt(rollarg, 10, 64)
			if err != nil {
				lgc.textReply(fmt.Sprintf("参数解析错误 - %v", rollarg))
				return
			}
		}
	}
	if min > max {
		lgc.textReply(fmt.Sprintf("参数解析错误 - %v", args))
		return
	}
	result := rand.Int63n(max-min+1) + min
	lgc.textReply(strconv.FormatInt(result, 10))
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

func (lgc *LspGroupCommand) reply(msg *message.SendingMessage) {
	lgc.qqClient.SendGroupMessage(lgc.msg.GroupCode, msg)
}
