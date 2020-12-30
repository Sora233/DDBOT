package lsp

import (
	"bytes"
	"fmt"
	miraiBot "github.com/Logiase/MiraiGo-Template/bot"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Sora233/Sora233-MiraiGo/image_pool"
	"github.com/Sora233/Sora233-MiraiGo/image_pool/lolicon_pool"
	"github.com/Sora233/Sora233-MiraiGo/lsp/aliyun"
	"github.com/Sora233/Sora233-MiraiGo/lsp/bilibili"
	localdb "github.com/Sora233/Sora233-MiraiGo/lsp/buntdb"
	"github.com/Sora233/Sora233-MiraiGo/lsp/concern"
	"github.com/forestgiant/sliceutil"
	"github.com/nfnt/resize"
	"github.com/tidwall/buntdb"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

type LspGroupCommand struct {
	bot *miraiBot.Bot
	msg *message.GroupMessage
	l   *Lsp
}

func NewLspGroupCommand(bot *miraiBot.Bot, msg *message.GroupMessage, l *Lsp) *LspGroupCommand {
	return &LspGroupCommand{
		bot: bot,
		msg: msg,
		l:   l,
	}
}

func (lgc *LspGroupCommand) Execute() {
	if text, ok := lgc.msg.Elements[0].(*message.TextElement); ok {
		args := strings.Split(text.Content, " ")
		switch args[0] {
		case "/lsp":
			lgc.LspCommand()
		case "/色图":
			lgc.SetuCommand(false)
		case "/黄图":
			lgc.SetuCommand(true)
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
		if lgc.msg.Sender.Uin != lgc.bot.Uin {
			lgc.ImageContent()
		}
	}
}

func (lgc *LspGroupCommand) LspCommand() {
	msg := lgc.msg
	bot := lgc.bot
	groupCode := msg.GroupCode

	log := logger.WithField("GroupCode", groupCode)
	log.Infof("run lsp command")
	defer log.Info("lsp command end")

	sendingMsg := message.NewSendingMessage()
	sendingMsg.Append(message.NewReply(msg))
	sendingMsg.Append(message.NewText("LSP竟然是你"))
	bot.SendGroupMessage(groupCode, sendingMsg)
	return
}

func (lgc *LspGroupCommand) SetuCommand(r18 bool) {
	msg := lgc.msg
	bot := lgc.bot
	groupCode := msg.GroupCode

	log := logger.WithField("GroupCode", groupCode)
	log.Info("run setu command")
	defer log.Info("setu command end")

	sendingMsg := message.NewSendingMessage()
	sendingMsg.Append(message.NewReply(msg))

	var options []image_pool.OptionFunc
	if r18 {
		options = append(options, lolicon_pool.R18Option(lolicon_pool.R18_ON))
	} else {
		options = append(options, lolicon_pool.R18Option(lolicon_pool.R18_OFF))
	}
	img, err := lgc.l.GetImageFromPool(options...)
	if err != nil {
		log.Errorf("get from pool failed %v", err)
		lgc.textReply("获取失败")
		return
	}
	imgBytes, err := img.Content()
	if err != nil {
		log.Errorf("get image bytes failed %v", err)
		lgc.textReply("获取失败")
		return
	}
	dImage, format, err := image.Decode(bytes.NewReader(imgBytes))
	if err != nil {
		log.Errorf("image decode failed %v", err)
		lgc.textReply("获取失败")
		return
	}
	log = log.WithField("format", format)
	resizedImage := resize.Thumbnail(1280, 860, dImage, resize.Lanczos3)
	resizedImageBuffer := bytes.NewBuffer(make([]byte, 0))

	switch format {
	case "jpeg":
		err = jpeg.Encode(resizedImageBuffer, resizedImage, nil)
	case "gif":
		err = gif.Encode(resizedImageBuffer, resizedImage, nil)
	case "png":
		err = png.Encode(resizedImageBuffer, resizedImage)
	}

	if err != nil {
		log.Errorf("resized image encode failed %v", err)
		lgc.textReply("获取失败")
		return
	}
	groupImage, err := bot.UploadGroupImage(groupCode, resizedImageBuffer.Bytes())
	if err != nil {
		log.Errorf("upload group image failed %v", err)
		lgc.textReply("上传失败")
		return
	}
	sendingMsg.Append(groupImage)
	if loliconImage, ok := img.(*lolicon_pool.Setu); ok {
		log.WithField("author", loliconImage.Author).
			WithField("r18", loliconImage.R18).
			WithField("pid", loliconImage.Pid).
			WithField("tags", loliconImage.Tags).
			WithField("title", loliconImage.Title).
			WithField("upload_url", groupImage.Url).
			Debug("debug image")
		sendingMsg.Append(message.NewText(fmt.Sprintf("标题：%v\n", loliconImage.Title)))
		sendingMsg.Append(message.NewText(fmt.Sprintf("作者：%v\n", loliconImage.Author)))
		sendingMsg.Append(message.NewText(fmt.Sprintf("PID：%v\n", loliconImage.Pid)))
		sendingMsg.Append(message.NewText(fmt.Sprintf("R18：%v", loliconImage.R18)))
	}
	lgc.reply(sendingMsg)
	return

}

func (lgc *LspGroupCommand) WatchCommand(remove bool) {
	msg := lgc.msg
	groupCode := msg.GroupCode

	log := logger.WithField("GroupCode", groupCode)
	log.Info("run watch command")
	defer log.Info("watch command end")

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
	if len(args) > 2 {
		log.WithField("content", text).Errorf("watch args error")
		lgc.textReply("参数错误 - Usage: /watch [bilibili] [news/live] id")
		return
	}

	for _, arg := range args {
		switch arg {
		case "bilibili":
			site = "bilibili"
		case "news":
			watchType = concern.BilibiliNews
		case "live":
			watchType = concern.BibiliLive
		default:
			log.WithField("content", text).Errorf("watch args error")
			lgc.textReply("参数错误 - Usage: /watch [bilibili] [news/live] id")
			return
		}
	}

	switch site {
	case "bilibili":
		if watchType == concern.BibiliLive {
			if remove {
				// unwatch
				if err := lgc.l.bilibiliConcern.Remove(groupCode, id, concern.BibiliLive); err != nil {
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
			if infoResp.Code != 0 {
				log.WithField("mid", id).WithField("code", infoResp.Code).Errorf(infoResp.Message)
				lgc.textReply(fmt.Sprintf("查询用户信息失败 %v - %v %v", id, infoResp.Code, infoResp.Message))
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
			if err := lgc.l.bilibiliConcern.AddLiveRoom(groupCode, id,
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
	log.Info("run list living command")
	defer log.Info("list living command end")

	text := msg.Elements[0].(*message.TextElement).Content

	args := strings.Split(text, " ")[1:]

	all := false

	if len(args) >= 1 {
		if args[0] == "all" {
			all = true
		}
	}

	living, err := lgc.l.bilibiliConcern.ListLiving(groupCode, all)
	if err != nil {
		log.Debugf("list living failed %v", err)
		lgc.textReply(fmt.Sprintf("list living 失败 - %v", err))
		return
	}
	if living == nil {
		lgc.textReply("关注列表为空，可以使用/watch命令关注")
		return
	}
	listMsg := message.NewSendingMessage()
	for idx, liveInfo := range living {
		if idx != 0 {
			listMsg.Append(message.NewText("\n"))
		}
		notifyMsg := lgc.l.NotifyMessage(lgc.bot, liveInfo)
		for _, msg := range notifyMsg {
			listMsg.Append(msg)
		}
	}
	if len(listMsg.Elements) == 0 {
		listMsg.Append(message.NewText("无人直播"))
	}
	lgc.reply(listMsg)

}

func (lgc *LspGroupCommand) RollCommand() {
	msg := lgc.msg
	groupCode := msg.GroupCode

	log := logger.WithField("GroupCode", groupCode)
	log.Info("run roll command")
	defer log.Info("roll command end")

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
	defer log.Info("checkin command end")

	db, err := localdb.GetClient()
	if err != nil {
		logger.Errorf("get db failed %v", err)
		return
	}
	date := time.Now().Format("20060102")

	err = db.Update(func(tx *buntdb.Tx) error {
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
	if err != nil {
		log.Errorf("签到失败")
	}
}

func (lgc *LspGroupCommand) ImageContent() {
	msg := lgc.msg
	bot := lgc.bot

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
					bot.SendGroupMessage(groupCode, sendingMsg)
					return
				} else if rating == aliyun.ScenePorn {
					sendingMsg := message.NewSendingMessage()
					sendingMsg.Append(message.NewReply(msg))
					sendingMsg.Append(message.NewText("多发点"))
					bot.SendGroupMessage(groupCode, sendingMsg)
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
	bot := lgc.bot

	sendingMsg := message.NewSendingMessage()
	sendingMsg.Append(message.NewReply(msg))
	sendingMsg.Append(message.NewText(text))
	bot.SendGroupMessage(msg.GroupCode, sendingMsg)
}

func (lgc *LspGroupCommand) reply(msg *message.SendingMessage) {
	lgc.bot.SendGroupMessage(lgc.msg.GroupCode, msg)
}
