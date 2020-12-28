package lsp

import (
	"fmt"
	"github.com/Mrs4s/MiraiGo/client"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Sora233/Sora233-MiraiGo/lsp/aliyun"
	"github.com/Sora233/Sora233-MiraiGo/lsp/bilibili"
	localdb "github.com/Sora233/Sora233-MiraiGo/lsp/buntdb"
	"github.com/Sora233/Sora233-MiraiGo/lsp/concern"
	"github.com/forestgiant/sliceutil"
	"github.com/tidwall/buntdb"
	"math/rand"
	"strconv"
	"strings"
	"time"
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
			lgc.ListLivingCommand()
		case "/签到":
			lgc.CheckinCommand()
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
		lgc.textReply("参数错误 - Usage: /watch [bilibili] [news/live] id")
		return
	}
	site := "bilibili"
	watchType := concern.BibiliLive

	id, err := strconv.ParseInt(args[len(args)-1], 10, 64)
	if err != nil {
		log.WithField("content", text).Errorf("watch args error")
		lgc.textReply("参数错误 - 未知的id：" + args[len(args)-1])
		return
	}

	args = args[:len(args)-1]
	for _, arg := range args {
		switch arg {
		case "bilibili":
			site = "bilibili"
		case "news":
			watchType = concern.BilibiliNews
		case "live":
			watchType = concern.BibiliLive
		default:
			log.WithField("content", text).Errorf("watch need args")
			lgc.textReply("参数错误 - Usage: /watch [bilibili] [news/live] id")
		}
	}

	switch site {
	case "bilibili":
		if watchType == concern.BibiliLive {
			if remove {
				// unwatch
				if err := lgc.l.BilibiliConcern.Remove(groupCode, id, concern.BibiliLive); err != nil {
					lgc.textReply(fmt.Sprintf("unwatch失败 - %v", err))
					break
				} else {
					log.WithField("mid", id).Debugf("unwatch success")
					lgc.textReply("unwatch成功")
				}
				return
			}

			// watch
			infoResp, err := bilibili.XSpaceAccInfo(id)
			if err != nil {
				log.WithField("mid", id).Error(err)
				lgc.textReply(fmt.Sprintf("查询用户信息失败 %v - %v", id, err))
				break
			}

			name := infoResp.GetData().GetName()

			if sliceutil.Contains([]int64{491474049}, id) {
				lgc.textReply(fmt.Sprintf("watch失败 - 用户 %v 禁止watch", name))
				break
			}
			if bilibili.RoomStatus(infoResp.GetData().GetLiveRoom().GetRoomStatus()) == bilibili.RoomStatus_NonExist {
				lgc.textReply(fmt.Sprintf("watch失败 - 用户 %v 暂未开通直播间", name))
				break
			}
			if err := lgc.l.BilibiliConcern.AddLiveRoom(groupCode, id,
				infoResp.GetData().GetLiveRoom().GetRoomid(),
			); err != nil {

				log.WithField("mid", id).Errorf("watch error %v", err)
				lgc.textReply(fmt.Sprintf("watch失败 - %v", err))
				break
			}
			log.WithField("mid", id).Debugf("watch success")
			lgc.textReply(fmt.Sprintf("watch成功 - Bilibili用户 %v", name))
		}
	default:
		log.WithField("site", site).Error("unsupported")
		lgc.textReply("未支持的网站")
	}
}

func (lgc *LspGroupCommand) ListLivingCommand() {
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

func (lgc *LspGroupCommand) CheckinCommand() {
	msg := lgc.msg
	groupCode := msg.GroupCode

	log := logger.WithField("GroupCode", groupCode)
	log.Infof("run checkin command")

	db, err := localdb.GetClient()
	if err != nil {
		logger.Errorf("get db failed %v", err)
		return
	}
	date := time.Now().Format("20060102")

	db.Update(func(tx *buntdb.Tx) error {
		var score int64
		key := localdb.Key("Score", groupCode, msg.Sender.Uin)
		dateMarker := localdb.Key("ScoreDate", groupCode, msg.Sender.Uin, date, nil)

		_, err := tx.Get(dateMarker)
		if err != buntdb.ErrNotFound {
			lgc.textReply("明天再来吧")
			return nil
		}

		val, err := tx.Get(key)
		if err == buntdb.ErrNotFound {
			score = 0
		} else {
			score, err = strconv.ParseInt(val, 10, 64)
			if err != nil {
				log.WithField("value", val).Errorf("parse score failed %v", err)
				return err
			}
		}
		score += 1
		_, _, err = tx.Set(key, strconv.FormatInt(score, 10), nil)
		if err != nil {
			log.WithField("sender", msg.Sender.Uin).Errorf("update score failed %v", err)
			return err
		}

		_, _, err = tx.Set(dateMarker, "1", nil)
		if err != nil {
			log.WithField("sender", msg.Sender.Uin).Errorf("update score marker failed %v", err)
			return err
		}
		lgc.textReply(fmt.Sprintf("签到成功！获得1积分，当前积分为%v", score))
		return nil
	})
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
