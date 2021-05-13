package lsp

import (
	"fmt"
	"github.com/Logiase/MiraiGo-Template/bot"
	"github.com/Logiase/MiraiGo-Template/config"
	"github.com/Logiase/MiraiGo-Template/utils"
	"github.com/Mrs4s/MiraiGo/client"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Sora233/DDBOT/concern"
	"github.com/Sora233/DDBOT/image_pool"
	"github.com/Sora233/DDBOT/image_pool/local_pool"
	"github.com/Sora233/DDBOT/image_pool/lolicon_pool"
	"github.com/Sora233/DDBOT/lsp/aliyun"
	"github.com/Sora233/DDBOT/lsp/bilibili"
	localdb "github.com/Sora233/DDBOT/lsp/buntdb"
	"github.com/Sora233/DDBOT/lsp/douyu"
	"github.com/Sora233/DDBOT/lsp/permission"
	"github.com/Sora233/DDBOT/lsp/youtube"
	"github.com/Sora233/DDBOT/proxy_pool"
	"github.com/Sora233/DDBOT/proxy_pool/local_proxy_pool"
	"github.com/Sora233/DDBOT/proxy_pool/py"
	"github.com/Sora233/DDBOT/proxy_pool/zhima"
	zhimaproxypool "github.com/Sora233/zhima-proxy-pool"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/sirupsen/logrus"
	"github.com/tidwall/buntdb"
	"math/rand"
	"os"
	"path"
	"strings"
	"sync"
	"time"
)

const ModuleName = "me.sora233.Lsp"

var logger = utils.GetModuleLogger(ModuleName)

var Debug = false

type Lsp struct {
	bilibiliConcern *bilibili.Concern
	douyuConcern    *douyu.Concern
	youtubeConcern  *youtube.Concern
	pool            image_pool.Pool
	concernNotify   chan concern.Notify
	stop            chan interface{}
	status          *Status

	PermissionStateManager *permission.StateManager
	LspStateManager        *StateManager
}

func (l *Lsp) MiraiGoModule() bot.ModuleInfo {
	return bot.ModuleInfo{
		ID:       ModuleName,
		Instance: Instance,
	}
}

func (l *Lsp) Init() {
	log := logger.WithField("log_level", config.GlobalConfig.GetString("logLevel"))
	lev, err := logrus.ParseLevel(config.GlobalConfig.GetString("logLevel"))
	if err != nil {
		logrus.SetLevel(logrus.DebugLevel)
		log.Warn("unknown log level")
	} else {
		logrus.SetLevel(lev)
		log.Info("set log level")
	}
	if err := localdb.InitBuntDB(); err != nil {
		panic(err)
	}

	bilibili.SetVerify(config.GlobalConfig.GetString("bilibili.SESSDATA"), config.GlobalConfig.GetString("bilibili.bili_jct"))

	keyId := config.GlobalConfig.GetString("aliyun.accessKeyID")
	keySecret := config.GlobalConfig.GetString("aliyun.accessKeySecret")
	if keyId != "" && keySecret != "" {
		aliyun.InitAliyun(keyId, keySecret)
		l.status.AliyunEnable = true
	} else {
		log.Debug("aliyun not init, some feature is not usable")
	}

	l.PermissionStateManager = permission.NewStateManager()
	l.LspStateManager = NewStateManager()

	l.bilibiliConcern = bilibili.NewConcern(l.concernNotify)
	l.douyuConcern = douyu.NewConcern(l.concernNotify)
	l.youtubeConcern = youtube.NewConcern(l.concernNotify)

	imagePoolType := config.GlobalConfig.GetString("imagePool.type")
	log = logger.WithField("image_pool_type", imagePoolType)

	switch imagePoolType {
	case "loliconPool":
		pool, err := lolicon_pool.NewLoliconPool(&lolicon_pool.Config{
			ApiKey:   config.GlobalConfig.GetString("loliconPool.apikey"),
			CacheMin: config.GlobalConfig.GetInt("loliconPool.cacheMin"),
			CacheMax: config.GlobalConfig.GetInt("loliconPool.cacheMax"),
		})
		if err != nil {
			log.Errorf("can not init pool %v", err)
		} else {
			l.pool = pool
			log.Debugf("init image pool")
			l.status.ImagePoolEnable = true
		}
	case "localPool":
		pool, err := local_pool.NewLocalPool(config.GlobalConfig.GetString("localPool.imageDir"))
		if err != nil {
			log.Errorf("can not init pool %v", err)
		} else {
			l.pool = pool
			log.Debugf("init image pool")
			l.status.ImagePoolEnable = true
		}
	case "off":
		log.Debug("image pool turn off")
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
			l.status.ProxyPoolEnable = true
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
		l.status.ProxyPoolEnable = true
	case "localProxyPool":
		overseaProxies := config.GlobalConfig.GetStringSlice("localProxyPool.oversea")
		mainlandProxies := config.GlobalConfig.GetStringSlice("localProxyPool.mainland")
		var proxies []*local_proxy_pool.Proxy
		for _, proxy := range overseaProxies {
			proxies = append(proxies, &local_proxy_pool.Proxy{
				Type:  proxy_pool.PreferOversea,
				Proxy: proxy,
			})
		}
		for _, proxy := range mainlandProxies {
			proxies = append(proxies, &local_proxy_pool.Proxy{
				Type:  proxy_pool.PreferMainland,
				Proxy: proxy,
			})
		}
		pool := local_proxy_pool.NewLocalPool(proxies)
		proxy_pool.Init(pool)
		log.WithField("local_proxy_num", len(proxies)).Debug("debug")
		l.status.ProxyPoolEnable = true
	case "off":
		log.Debug("proxy pool turn off")
	default:
		log.Errorf("unknown proxy type")
	}
}

func (l *Lsp) PostInit() {
}

func (l *Lsp) Serve(bot *bot.Bot) {
	bot.OnGroupInvited(func(qqClient *client.QQClient, request *client.GroupInvitedRequest) {
		log := logger.WithField("group_code", request.GroupCode).
			WithField("group_name", request.GroupName).
			WithField("invitor_uin", request.InvitorUin).
			WithField("invitor_nick", request.InvitorNick)

		if l.PermissionStateManager.CheckBlockList(request.InvitorUin) {
			log.Debug("blocked invited")
			request.Reject(false, "")
			return
		}

		log.Info("new group invited")
		fi := bot.FindFriend(request.InvitorUin)
		if fi == nil {
			request.Reject(false, "未找到阁下的好友信息，请添加好友进行操作")
			log.Errorf("can not find friend info")
			return
		}
		sendingMsg := message.NewSendingMessage()
		sendingMsg.Append(message.NewText(fmt.Sprintf("阁下的群邀请已通过，基于对阁下的信任，阁下已获得本bot在群【%s】的控制权限，相信阁下不会滥用本bot。", request.GroupName)))
		bot.SendPrivateMessage(request.InvitorUin, sendingMsg)
		if err := l.PermissionStateManager.GrantGroupRole(request.GroupCode, request.InvitorUin, permission.GroupAdmin); err != nil {
			log.WithField("target", request.InvitorUin).
				WithField("group_code", request.GroupCode).
				Errorf("grant group admin failed %v", err)
		}
		request.Accept()

		//err := l.LspStateManager.SaveGroupInvitor(request.GroupCode, request.InvitorUin)
		//if err == localdb.ErrKeyExist {
		//	request.Reject(false, "已有其他群友邀请加群，请通知管理员审核")
		//	log.Errorf("invited duplicate")
		//	return
		//} else if err != nil {
		//	request.Reject(false, "未知问题，加群失败")
		//	log.Errorf("invited process failed %v", err)
		//	return
		//} else {
		//	request.Accept()
		//	sendingMsg := message.NewSendingMessage()
		//	sendingMsg.Append(message.NewText(fmt.Sprintf("已接受阁下的群邀请。在成功入群后，基于对阁下的信任，阁下将获得bot在群【%s】的控制权限。", request.GroupName)))
		//	bot.SendPrivateMessage(request.InvitorUin, sendingMsg)
		//}
	})

	bot.OnNewFriendRequest(func(qqClient *client.QQClient, request *client.NewFriendRequest) {
		log := logger.WithField("uin", request.RequesterUin).
			WithField("nickname", request.RequesterNick).
			WithField("message", request.Message)
		if l.PermissionStateManager.CheckBlockList(request.RequesterUin) {
			log.Debug("blocked new friend")
			request.Reject()
			return
		}
		log.Info("friend request")
		request.Accept()
	})

	bot.OnNewFriendAdded(func(qqClient *client.QQClient, event *client.NewFriendEvent) {
		logger.WithField("uin", event.Friend.Uin).
			WithField("nickname", event.Friend.Nickname).
			Info("new friend")
		sendingMsg := message.NewSendingMessage()
		sendingMsg.Append(message.NewText("阁下的好友请求已通过，请使用/help查看帮助，然后在群成员页面邀请bot加群（bot不会主动加群）。"))
		bot.SendPrivateMessage(event.Friend.Uin, sendingMsg)
	})

	bot.OnJoinGroup(func(qqClient *client.QQClient, info *client.GroupInfo) {
		l.FreshIndex()
		log := logger.WithField("group_code", info.Code).
			WithField("member_count", info.MemberCount).
			WithField("group_name", info.Name)
		log.Info("join group")

		minfo := info.FindMember(bot.Uin)
		minfo.EditCard("【bot】")
		//go func() {
		//	// sbtx
		//	// 有一些sb的时候邀请加群会自动同意，这时可能join callback在invited callback之前触发
		//	var (
		//		target int64
		//		err    error
		//	)
		//	localutils.Retry(10, time.Second*3, func() bool {
		//		target, err = l.LspStateManager.GetGroupInvitor(info.Code)
		//		return err == nil
		//	})
		//	if err == buntdb.ErrNotFound {
		//		log.Debug("no invitor found finally, skip grant group admin role")
		//	} else if err != nil {
		//		log.Errorf("get invitor err %v", err)
		//		return
		//	} else {
		//		log = log.WithField("invitor", target)
		//		log.Debug("grant group admin role")
		//		if err := l.PermissionStateManager.GrantGroupRole(info.Code, target, permission.GroupAdmin); err != nil {
		//			log.Errorf("grant group admin role failed %v", err)
		//		}
		//	}
		//}()
	})

	bot.OnLeaveGroup(func(qqClient *client.QQClient, event *client.GroupLeaveEvent) {
		logger.WithField("group_code", event.Group.Code).
			WithField("group_name", event.Group.Name).
			WithField("member_count", event.Group.MemberCount).
			Info("leave group")
		l.RemoveAllByGroup(event.Group.Code)
	})

	bot.OnGroupMessage(func(qqClient *client.QQClient, msg *message.GroupMessage) {
		if len(msg.Elements) <= 0 {
			return
		}
		if err := l.LspStateManager.SaveMessageImageUrl(msg.GroupCode, msg.Id, msg.Elements); err != nil {
			logger.Errorf("SaveMessageImageUrl failed %v", err)
		}
		cmd := NewLspGroupCommand(bot, l, msg)
		if Debug {
			cmd.Debug()
		}
		if !l.LspStateManager.IsMuted(msg.GroupCode, bot.Uin) {
			go cmd.Execute()
		}
	})

	bot.OnSelfGroupMessage(func(qqClient *client.QQClient, msg *message.GroupMessage) {
		if len(msg.Elements) <= 0 {
			return
		}
		if err := l.LspStateManager.SaveMessageImageUrl(msg.GroupCode, msg.Id, msg.Elements); err != nil {
			logger.Errorf("SaveMessageImageUrl failed %v", err)
		}
	})

	bot.OnGroupMuted(func(qqClient *client.QQClient, event *client.GroupMuteEvent) {
		if err := l.LspStateManager.Muted(event.GroupCode, event.TargetUin, event.Time); err != nil {
			logger.Errorf("Muted failed %v", err)
		}
	})

	bot.OnPrivateMessage(func(qqClient *client.QQClient, msg *message.PrivateMessage) {
		// TODO 这个问题已经过已经修复了，再观察一段时间
		if msg.Time < int32(time.Now().Add(time.Minute*-1).Unix()) {
			logger.WithField("Sender", msg.Sender.DisplayName()).
				WithField("time", time.Unix(int64(msg.Time), 0)).
				WithField("MessageID", msg.Id).
				Debug("past private message got, skip.")
			// 有时候消息会再触发一次，应该是tx的问题
			return
		}
		if len(msg.Elements) == 0 {
			return
		}
		cmd := NewLspPrivateCommand(bot, l, msg)
		if Debug {
			cmd.Debug()
		}
		go cmd.Execute()
	})
	bot.OnDisconnected(func(qqClient *client.QQClient, event *client.ClientDisconnectedEvent) {
		go func() {
			time.Sleep(time.Second * 3)
			os.Exit(1)
		}()
	})

	bot.OnLog(func() func(*client.QQClient, *client.LogEvent) {
		// create a new logger for miraigo log
		writer, err := rotatelogs.New(
			path.Join("miraigo-logs", "%Y-%m-%d.log"),
			rotatelogs.WithMaxAge(7*24*time.Hour),
			rotatelogs.WithRotationTime(24*time.Hour),
		)
		if err != nil {
			logrus.WithError(err).Error("unable to write logs")
			return func(qqClient *client.QQClient, event *client.LogEvent) {
				logger.WithField("type", event.Type).Debug(event.Message)
			}
		}
		miraigoLog := logrus.New()
		miraigoLog.SetOutput(writer)
		miraigoLog.SetLevel(logrus.DebugLevel)
		miraigoLog.SetFormatter(&logrus.JSONFormatter{})
		return func(qqClient *client.QQClient, event *client.LogEvent) {
			miraigoLog.WithField("type", event.Type).Debug(event.Message)
		}
	}())
}

func (l *Lsp) PostStart(bot *bot.Bot) {
	l.bilibiliConcern.Start()
	l.douyuConcern.Start()
	l.youtubeConcern.Start()
}

func (l *Lsp) Start(bot *bot.Bot) {
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
	var cacheLabel string
	localdb.RTxCover(func(tx *buntdb.Tx) error {
		key := localdb.ImageCacheKey(string(img.Md5))
		val, err := tx.Get(key)
		if err == nil {
			cacheLabel = val
		}
		return nil
	})
	if len(cacheLabel) != 0 {
		logger.WithField("label", cacheLabel).Debug("detect cache")
		return cacheLabel
	}
	if rand.Int()%2 == 0 {
		logger.Tracef("random skip")
		return ""
	}
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
	label := resp.Data.Results[0].SubResults[0].Label
	localdb.RWTxCover(func(tx *buntdb.Tx) error {
		key := localdb.ImageCacheKey(string(img.Md5))
		_, _, err := tx.Set(key, label, localdb.ExpireOption(time.Hour*72))
		return err
	})
	return label
}

func (l *Lsp) FreshIndex() {
	l.bilibiliConcern.FreshIndex()
	l.douyuConcern.FreshIndex()
	l.youtubeConcern.FreshIndex()
	l.PermissionStateManager.FreshIndex()
	l.LspStateManager.FreshIndex()
}

func (l *Lsp) RemoveAllByGroup(groupCode int64) {
	l.bilibiliConcern.RemoveAllByGroupCode(groupCode)
	l.douyuConcern.RemoveAllByGroupCode(groupCode)
	l.youtubeConcern.RemoveAllByGroupCode(groupCode)
	l.PermissionStateManager.RemoveAllByGroup(groupCode)
}

func (l *Lsp) GetImageFromPool(options ...image_pool.OptionFunc) ([]image_pool.Image, error) {
	if l.pool == nil {
		return nil, image_pool.ErrNotInit
	}
	return l.pool.Get(options...)
}

func (l *Lsp) sendGroupMessage(groupCode int64, msg *message.SendingMessage) *message.GroupMessage {
	if l.LspStateManager.IsMuted(groupCode, bot.Instance.Uin) {
		logger.WithField("groupCode", groupCode).Debug("skip muted group")
		return &message.GroupMessage{Id: -1}
	}
	// don't know why
	// msg.Elements = l.compactTextElements(msg.Elements)
	res := bot.Instance.SendGroupMessage(groupCode, msg)
	if res.Id == -1 {
		logger.WithField("group_code", groupCode).Errorf("send group message failed")
	}
	return res
}

func (l *Lsp) compactTextElements(elements []message.IMessageElement) []message.IMessageElement {
	var compactMsg []message.IMessageElement
	sb := strings.Builder{}
	for _, e := range elements {
		if e.Type() == message.Text {
			sb.WriteString(e.(*message.TextElement).Content)
		} else {
			if sb.Len() != 0 {
				compactMsg = append(compactMsg, message.NewText(sb.String()))
				sb = strings.Builder{}
			}
			compactMsg = append(compactMsg, e)
		}
	}
	if sb.Len() != 0 {
		compactMsg = append(compactMsg, message.NewText(sb.String()))
	}
	return compactMsg
}

var Instance *Lsp

func init() {
	Instance = &Lsp{
		concernNotify: make(chan concern.Notify, 500),
		stop:          make(chan interface{}),
		status:        NewStatus(),
	}
	bot.RegisterModule(Instance)
}
