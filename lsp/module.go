package lsp

import (
	"fmt"
	"github.com/Logiase/MiraiGo-Template/bot"
	"github.com/Logiase/MiraiGo-Template/config"
	"github.com/Logiase/MiraiGo-Template/utils"
	"github.com/Mrs4s/MiraiGo/client"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Sora233/Sora233-MiraiGo/lsp/aliyun"
	"github.com/Sora233/Sora233-MiraiGo/lsp/bilibili"
	localdb "github.com/Sora233/Sora233-MiraiGo/lsp/buntdb"
	"github.com/Sora233/Sora233-MiraiGo/lsp/concern"
	"github.com/Sora233/Sora233-MiraiGo/lsp/image_pool"
	"github.com/Sora233/Sora233-MiraiGo/lsp/image_pool/local_pool"
	"github.com/Sora233/Sora233-MiraiGo/lsp/image_pool/lolicon_pool"
	"github.com/asmcos/requests"
	"strings"
	"sync"
)

const ModuleName = "me.sora233.Lsp"

var logger = utils.GetModuleLogger(ModuleName)

type Lsp struct {
	pool            image_pool.Pool
	bilibiliConcern *bilibili.Concern

	concernNotify chan concern.Notify
	stop          chan interface{}
}

func (l *Lsp) MiraiGoModule() bot.ModuleInfo {
	return bot.ModuleInfo{
		ID:       ModuleName,
		Instance: Instance,
	}
}

func (l *Lsp) Init() {
	aliyun.InitAliyun()
	l.bilibiliConcern = bilibili.NewConcern(l.concernNotify)

	poolType := config.GlobalConfig.GetString("imagePool.type")
	log := logger.WithField("pool_type", poolType)

	switch poolType {
	case "loliconPool":
		apikey := config.GlobalConfig.GetString("loliconPool.apikey")
		pool, err := lolicon_pool.NewLoliconPool(apikey)
		if err != nil {
			log.Errorf("can not init pool %v", err)
		} else {
			l.pool = pool
			log.Debugf("init pool")
		}
	case "localPool":
		pool, err := local_pool.NewLocalPool(config.GlobalConfig.GetString("localPool.imageDir"))
		if err != nil {
			log.Errorf("can not init pool %v", err)
		} else {
			l.pool = pool
			log.Debugf("init pool")
		}
	default:
		log.Errorf("unknown pool")
	}

}

func (l *Lsp) PostInit() {
	if err := localdb.InitBuntDB(); err != nil {
		panic(err)
	}
}

func (l *Lsp) Serve(bot *bot.Bot) {
	bot.OnGroupInvited(func(qqClient *client.QQClient, request *client.GroupInvitedRequest) {
		request.Accept()
	})

	bot.OnNewFriendRequest(func(qqClient *client.QQClient, request *client.NewFriendRequest) {
		request.Accept()
	})

	bot.OnJoinGroup(func(qqClient *client.QQClient, info *client.GroupInfo) {
		logger.WithField("group_code", info.Code).Debugf("join group")
		l.bilibiliConcern.FreshIndex()
	})
	bot.OnLeaveGroup(func(qqClient *client.QQClient, event *client.GroupLeaveEvent) {
		logger.WithField("group_code", event.Group.Code).Debugf("leave group")
		l.bilibiliConcern.FreshIndex()
	})

	bot.OnGroupMessage(func(qqClient *client.QQClient, msg *message.GroupMessage) {
		if len(msg.Elements) <= 0 {
			return
		}
		cmd := NewLspGroupCommand(bot, msg, l)
		go cmd.Execute()
	})

	bot.OnPrivateMessage(func(qqClient *client.QQClient, msg *message.PrivateMessage) {
		cmds := strings.Split(msg.ToString(), " ")
		if cmds[0] == "/ping" {
			sendingMsg := message.NewSendingMessage()
			sendingMsg.Append(message.NewText("pong"))
			qqClient.SendPrivateMessage(msg.Sender.Uin, sendingMsg)
		}
	})
}

func (l *Lsp) Start(bot *bot.Bot) {
	l.bilibiliConcern.Start()
	go l.ConcernNotify(bot)
}

func (l *Lsp) Stop(bot *bot.Bot, wg *sync.WaitGroup) {
	defer wg.Done()
	if l.stop != nil {
		close(l.stop)
	}
	l.bilibiliConcern.Stop()
	if err := localdb.Close(); err != nil {
		logger.Errorf("close db err %v", err)
	}
}

func (l *Lsp) checkImage(img *message.ImageElement) string {
	logger.WithField("image_url", img.Url).Info("image here")
	resp, err := aliyun.Audit(img.Url)
	if err != nil {
		logger.Errorf("aliyun request error %v", err)
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

func (l *Lsp) ConcernNotify(bot *bot.Bot) {
	for {
		select {
		case inotify := <-l.concernNotify:
			switch inotify.Type() {
			case concern.BibiliLive:
				notify := (inotify).(*bilibili.ConcernLiveNotify)
				logger.WithField("GroupCode", notify.GroupCode).
					WithField("Name", notify.Name).
					WithField("Title", notify.LiveTitle).
					WithField("Status", notify.Status.String()).
					Info("notify")
				if notify.Status == bilibili.LiveStatus_Living {
					sendingMsg := message.NewSendingMessage()
					notifyMsg := l.NotifyMessage(bot, notify)
					for _, msg := range notifyMsg {
						sendingMsg.Append(msg)
					}
					bot.SendGroupMessage(notify.GroupCode, sendingMsg)
				}
			case concern.BilibiliNews:
				// TODO
			}
		}
	}
}

func (l *Lsp) NotifyMessage(bot *bot.Bot, inotify concern.Notify) []message.IMessageElement {
	var result []message.IMessageElement
	switch inotify.Type() {
	case concern.BibiliLive:
		notify := (inotify).(*bilibili.ConcernLiveNotify)
		switch notify.Status {
		case bilibili.LiveStatus_Living:
			result = append(result, message.NewText(fmt.Sprintf("%s正在直播【%s】\n", notify.Name, notify.LiveTitle)))
			result = append(result, message.NewText(notify.RoomUrl))
			coverResp, err := requests.Get(notify.Cover)
			if err == nil {
				if cover, err := bot.UploadGroupImage(notify.GroupCode, coverResp.Content()); err == nil {
					result = append(result, cover)
				}
			}
		case bilibili.LiveStatus_NoLiving:
			result = append(result, message.NewText(fmt.Sprintf("%s暂未直播\n", notify.Name)))
			result = append(result, message.NewText(notify.RoomUrl))
		}
	case concern.BilibiliNews:

	}

	return result
}

func (l *Lsp) FreshIndex() {
	l.bilibiliConcern.FreshIndex()
}

func (l *Lsp) GetImageFromPool(options ...image_pool.OptionFunc) (image_pool.Image, error) {
	return l.pool.Get(options...)
}

var Instance *Lsp

func init() {
	Instance = &Lsp{
		concernNotify: make(chan concern.Notify, 500),
		stop:          make(chan interface{}),
	}
	bot.RegisterModule(Instance)
}
