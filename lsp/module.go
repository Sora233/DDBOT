package lsp

import (
	"fmt"
	"github.com/Logiase/MiraiGo-Template/bot"
	"github.com/Logiase/MiraiGo-Template/config"
	"github.com/Logiase/MiraiGo-Template/utils"
	"github.com/Mrs4s/MiraiGo/client"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Sora233/Sora233-MiraiGo/concern"
	"github.com/Sora233/Sora233-MiraiGo/image_pool"
	"github.com/Sora233/Sora233-MiraiGo/image_pool/local_pool"
	"github.com/Sora233/Sora233-MiraiGo/image_pool/lolicon_pool"
	"github.com/Sora233/Sora233-MiraiGo/lsp/aliyun"
	"github.com/Sora233/Sora233-MiraiGo/lsp/bilibili"
	localdb "github.com/Sora233/Sora233-MiraiGo/lsp/buntdb"
	"github.com/Sora233/Sora233-MiraiGo/lsp/douyu"
	"github.com/Sora233/Sora233-MiraiGo/proxy_pool"
	"github.com/Sora233/Sora233-MiraiGo/proxy_pool/py"
	"github.com/Sora233/Sora233-MiraiGo/proxy_pool/zhima"
	localutils "github.com/Sora233/Sora233-MiraiGo/utils"
	zhimaproxypool "github.com/Sora233/zhima-proxy-pool"
	"github.com/asmcos/requests"
	"strings"
	"sync"
	"time"
)

const ModuleName = "me.sora233.Lsp"

var logger = utils.GetModuleLogger(ModuleName)

type Lsp struct {
	bilibiliConcern *bilibili.Concern
	douyuConcern    *douyu.Concern
	pool            image_pool.Pool
	concernNotify   chan concern.Notify
	stop            chan interface{}
}

func (l *Lsp) MiraiGoModule() bot.ModuleInfo {
	return bot.ModuleInfo{
		ID:       ModuleName,
		Instance: Instance,
	}
}

func (l *Lsp) Init() {
	if err := localdb.InitBuntDB(); err != nil {
		panic(err)
	}
	aliyun.InitAliyun()
	l.bilibiliConcern = bilibili.NewConcern(l.concernNotify)
	l.douyuConcern = douyu.NewConcern(l.concernNotify)

	imagePoolType := config.GlobalConfig.GetString("imagePool.type")
	log := logger.WithField("image_pool_type", imagePoolType)

	switch imagePoolType {
	case "loliconPool":
		apikey := config.GlobalConfig.GetString("loliconPool.apikey")
		pool, err := lolicon_pool.NewLoliconPool(apikey)
		if err != nil {
			log.Errorf("can not init pool %v", err)
		} else {
			l.pool = pool
			log.Debugf("init image pool")
		}
	case "localPool":
		pool, err := local_pool.NewLocalPool(config.GlobalConfig.GetString("localPool.imageDir"))
		if err != nil {
			log.Errorf("can not init pool %v", err)
		} else {
			l.pool = pool
			log.Debugf("init image pool")
		}
	default:
		log.Errorf("unknown pool")
	}

	proxyType := config.GlobalConfig.GetString("proxy.type")
	log = logger.WithField("proxy_type", proxyType)
	switch proxyType {
	case "pyProxyPool":
		host := config.GlobalConfig.GetString("pyProxyPool.host")
		log := log.WithField("host", host)
		pyPool, err := py.NewPYProxyPool(host)
		if err != nil {
			log.Errorf("init py pool err %v", err)
		} else {
			proxy_pool.Init(pyPool)
		}
	case "zhimaProxyPool":
		api := config.GlobalConfig.GetString("zhimaProxyPool.api")
		log.WithField("api", api).Debug("debug")
		cfg := &zhimaproxypool.Config{
			ApiAddr:   api,
			BackUpCap: config.GlobalConfig.GetInt("zhimaProxyPool.BackUpCap"),
			ActiveCap: config.GlobalConfig.GetInt("zhimaProxyPool.ActiveCap"),
			ClearTime: time.Second * time.Duration(config.GlobalConfig.GetInt("zhimaProxyPool.ClearTime")),
			TimeLimit: time.Minute * time.Duration(config.GlobalConfig.GetInt("zhimaProxyPool.TimeLimit")),
		}
		zhimaPool := zhimaproxypool.NewZhimaProxyPool(cfg, zhima.NewBuntdbPersister())
		proxy_pool.Init(zhima.NewZhimaWrapper(zhimaPool, 15))
	default:
		log.Errorf("unknown proxy type")
	}
}

func (l *Lsp) PostInit() {
}

func (l *Lsp) Serve(bot *bot.Bot) {
	bot.OnGroupInvited(func(qqClient *client.QQClient, request *client.GroupInvitedRequest) {
		logger.WithField("group_code", request.GroupCode).
			WithField("group_name", request.GroupName).
			WithField("invitor_uin", request.InvitorUin).
			WithField("invitor_nick", request.InvitorNick).
			Debug("new group invited")
		request.Accept()
	})

	bot.OnNewFriendRequest(func(qqClient *client.QQClient, request *client.NewFriendRequest) {
		logger.WithField("uin", request.RequesterUin).
			WithField("nickname", request.RequesterNick).
			WithField("message", request.Message).
			Debug("new friend")
		request.Accept()
	})

	bot.OnJoinGroup(func(qqClient *client.QQClient, info *client.GroupInfo) {
		logger.WithField("group_code", info.Code).Debugf("join group")
		l.FreshIndex()
	})
	bot.OnLeaveGroup(func(qqClient *client.QQClient, event *client.GroupLeaveEvent) {
		logger.WithField("group_code", event.Group.Code).Debugf("leave group")
		l.RemoveAll(event.Group.Code)
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
	l.douyuConcern.Start()
	go l.ConcernNotify(bot)
}

func (l *Lsp) Stop(bot *bot.Bot, wg *sync.WaitGroup) {
	defer wg.Done()
	if l.stop != nil {
		close(l.stop)
	}
	proxy_pool.Stop()
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
	if len(resp.Data.Results[0].SubResults) == 0 {
		logger.Errorf("aliyun response empty subResults")
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
				logger.WithField("site", bilibili.Site).
					WithField("GroupCode", notify.GroupCode).
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
				notify := (inotify).(*bilibili.ConcernNewsNotify)
				logger.WithField("site", bilibili.Site).
					WithField("GroupCode", notify.GroupCode).
					WithField("Name", notify.Name).
					WithField("NewsCount", len(notify.Cards)).
					Info("notify")
				sendingMsg := message.NewSendingMessage()
				notifyMsg := l.NotifyMessage(bot, notify)
				for _, msg := range notifyMsg {
					sendingMsg.Append(msg)
				}
				bot.SendGroupMessage(notify.GroupCode, sendingMsg)
			case concern.DouyuLive:
				notify := (inotify).(*douyu.ConcernLiveNotify)
				logger.WithField("site", douyu.Site).
					WithField("GroupCode", notify.GroupCode).
					WithField("Name", notify.Nickname).
					WithField("Title", notify.RoomName).
					WithField("Status", notify.ShowStatus.String()).
					Info("notify")
				if notify.ShowStatus == douyu.ShowStatus_Living {
					sendingMsg := message.NewSendingMessage()
					notifyMsg := l.NotifyMessage(bot, notify)
					for _, msg := range notifyMsg {
						sendingMsg.Append(msg)
					}
					bot.SendGroupMessage(notify.GroupCode, sendingMsg)
				}
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
		notify := (inotify).(*bilibili.ConcernNewsNotify)
		for index, card := range notify.Cards {
			dynamicUrl := bilibili.DynamicUrl(card.GetDesc().GetDynamicIdStr())
			date := time.Unix(int64(card.GetDesc().GetTimestamp()), 0).Format("2006-01-02 15:04:05")
			switch card.GetDesc().GetType() {
			case bilibili.DynamicDescType_WithOrigin:
				cardOrigin, err := notify.GetCardWithOrig(index)
				if err != nil {
					logger.WithField("name", notify.Name).WithField("card", card).Errorf("cast failed %v", err)
					continue
				}
				originName := cardOrigin.GetOriginUser().GetInfo().GetUname()
				result = append(result, message.NewText(fmt.Sprintf("%v转发了%v的动态，并说：\n%v\n", notify.Name, originName, cardOrigin.GetItem().GetContent())))
			case bilibili.DynamicDescType_WithImage:
				cardImage, err := notify.GetCardWithImage(index)
				if err != nil {
					logger.WithField("name", notify.Name).WithField("card", card).Errorf("cast failed %v", err)
					continue
				}
				cardImage.GetItem()
				result = append(result, message.NewText(fmt.Sprintf("%v发布了新态：\n%v\n%v\n", notify.Name, date, cardImage.GetItem().GetDescription())))
				for _, pic := range cardImage.GetItem().GetPictures() {
					img, err := localutils.ImageGetAndNorm(pic.GetImgSrc())
					if err != nil {
						continue
					}
					pic, err := bot.UploadGroupImage(notify.GroupCode, img)
					if err != nil {
						continue
					}
					result = append(result, pic)
				}
			case bilibili.DynamicDescType_TextOnly:
				cardText, err := notify.GetCardTextOnly(index)
				if err != nil {
					logger.WithField("name", notify.Name).WithField("card", card).Errorf("cast failed %v", err)
					continue
				}
				result = append(result, message.NewText(fmt.Sprintf("%v发布了新动态：\n%v\n%v\n", notify.Name, date, cardText.GetItem().GetContent())))
			case bilibili.DynamicDescType_WithVideo:
				cardVideo, err := notify.GetCardWithVideo(index)
				if err != nil {
					logger.WithField("name", notify.Name).WithField("card", card).Errorf("cast failed %v", err)
					continue
				}
				result = append(result, message.NewText(fmt.Sprintf("%v发布了新视频：\n%v\n%v\n", notify.Name, date, cardVideo.GetItem().GetTitle())))
				img, err := localutils.ImageGetAndNorm(cardVideo.GetItem().GetPic())
				if err != nil {
					continue
				}
				cover, err := bot.UploadGroupImage(notify.GroupCode, img)
				if err != nil {
					continue
				}
				result = append(result, cover)
			}
			result = append(result, message.NewText(dynamicUrl))
		}
	case concern.DouyuLive:
		notify := (inotify).(*douyu.ConcernLiveNotify)
		switch notify.ShowStatus {
		case douyu.ShowStatus_Living:
			result = append(result, message.NewText(fmt.Sprintf("斗鱼-%s正在直播【%s】\n", notify.Nickname, notify.RoomName)))
			result = append(result, message.NewText(notify.RoomUrl))
			coverResp, err := requests.Get(notify.GetAvatar().GetBig())
			if err == nil {
				if cover, err := bot.UploadGroupImage(notify.GroupCode, coverResp.Content()); err == nil {
					result = append(result, cover)
				}
			}
		case douyu.ShowStatus_NoLiving:
			result = append(result, message.NewText(fmt.Sprintf("斗鱼-%s暂未直播\n", notify.Nickname)))
			result = append(result, message.NewText(notify.RoomUrl))
		}
	}

	return result
}

func (l *Lsp) FreshIndex() {
	l.bilibiliConcern.FreshIndex()
	l.douyuConcern.FreshIndex()
}

func (l *Lsp) RemoveAll(groupCode int64) {
	l.bilibiliConcern.RemoveAll(groupCode)
	l.douyuConcern.RemoveAll(groupCode)
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
