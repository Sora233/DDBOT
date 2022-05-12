package lsp

import (
	"fmt"
	"github.com/Mrs4s/MiraiGo/client"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Sora233/DDBOT/image_pool"
	"github.com/Sora233/DDBOT/image_pool/local_pool"
	"github.com/Sora233/DDBOT/image_pool/lolicon_pool"
	localdb "github.com/Sora233/DDBOT/lsp/buntdb"
	"github.com/Sora233/DDBOT/lsp/cfg"
	"github.com/Sora233/DDBOT/lsp/concern"
	"github.com/Sora233/DDBOT/lsp/concern_type"
	"github.com/Sora233/DDBOT/lsp/mmsg"
	"github.com/Sora233/DDBOT/lsp/mmsg/mt"
	"github.com/Sora233/DDBOT/lsp/permission"
	"github.com/Sora233/DDBOT/lsp/template"
	"github.com/Sora233/DDBOT/lsp/version"
	"github.com/Sora233/DDBOT/proxy_pool"
	"github.com/Sora233/DDBOT/proxy_pool/local_proxy_pool"
	"github.com/Sora233/DDBOT/proxy_pool/py"
	localutils "github.com/Sora233/DDBOT/utils"
	"github.com/Sora233/DDBOT/utils/msgstringer"
	"github.com/Sora233/MiraiGo-Template/bot"
	"github.com/Sora233/MiraiGo-Template/config"
	"github.com/fsnotify/fsnotify"
	jsoniter "github.com/json-iterator/go"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	"github.com/tidwall/buntdb"
	"go.uber.org/atomic"
	"golang.org/x/sync/semaphore"
	"os"
	"reflect"
	"runtime/debug"
	"sync"
	"time"
)

const ModuleName = "me.sora233.Lsp"

var logger = logrus.WithField("module", ModuleName)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

var Debug = false

type Lsp struct {
	pool          image_pool.Pool
	concernNotify <-chan concern.Notify
	stop          chan interface{}
	wg            sync.WaitGroup
	status        *Status
	notifyWg      sync.WaitGroup
	msgLimit      *semaphore.Weighted
	cron          *cron.Cron

	PermissionStateManager *permission.StateManager
	LspStateManager        *StateManager
	started                atomic.Bool
}

func (l *Lsp) CommandShowName(command string) string {
	return cfg.GetCommandPrefix(command) + command
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
		log.Warn("无法识别logLevel，将使用Debug级别")
	} else {
		logrus.SetLevel(lev)
		log.Infof("设置logLevel为%v", lev.String())
	}

	l.msgLimit = semaphore.NewWeighted(int64(cfg.GetNotifyParallel()))

	if Tags != "UNKNOWN" {
		logger.Infof("DDBOT版本：Release版本【%v】", Tags)
	} else {
		if CommitId == "UNKNOWN" {
			logger.Infof("DDBOT版本：编译版本未知")
		} else {
			logger.Infof("DDBOT版本：编译版本【%v-%v】", BuildTime, CommitId)
		}
	}

	db := localdb.MustGetClient()
	var count int
	err = db.View(func(tx *buntdb.Tx) error {
		return tx.Ascend("", func(key, value string) bool {
			count++
			return true
		})
	})
	if err == nil && count == 0 {
		version.SetVersion(LspVersionName, LspSupportVersion)
	} else {
		curVersion := version.GetCurrentVersion(LspVersionName)
		if curVersion < 0 {
			log.Errorf("警告：无法检查数据库兼容性，程序可能无法正常工作")
		} else if curVersion > LspSupportVersion {
			log.Fatalf("警告：检查数据库兼容性失败！最高支持版本：%v，当前版本：%v", LspSupportVersion, curVersion)
		} else if curVersion < LspSupportVersion {
			// 应该更新下
			backupFileName := fmt.Sprintf("%v-%v", localdb.LSPDB, time.Now().Unix())
			log.Warnf(
				`警告：数据库兼容性检查完毕，当前需要从<%v>更新至<%v>，将备份当前数据库文件到"%v"`,
				curVersion, LspSupportVersion, backupFileName)
			f, err := os.Create(backupFileName)
			if err != nil {
				log.Fatalf(`无法创建备份文件<%v>：%v`, backupFileName, err)
			}
			err = db.Save(f)
			if err != nil {
				log.Fatalf(`无法备份数据库到<%v>：%v`, backupFileName, err)
			}
			log.Infof(`备份完成，已备份数据库到<%v>"`, backupFileName)
			log.Info("五秒后将开始更新数据库，如需取消请按Ctrl+C")
			time.Sleep(time.Second * 5)
			err = version.DoMigration(LspVersionName, lspMigrationMap)
			if err != nil {
				log.Fatalf("更新数据库失败：%v", err)
			}
		} else {
			log.Debugf("数据库兼容性检查完毕，当前已为最新模式：%v", curVersion)
		}
	}

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
			log.Infof("初始化%v图片池", imagePoolType)
			l.status.ImagePoolEnable = true
		}
	case "localPool":
		pool, err := local_pool.NewLocalPool(config.GlobalConfig.GetString("localPool.imageDir"))
		if err != nil {
			log.Errorf("初始化%v图片池失败 %v", imagePoolType, err)
		} else {
			l.pool = pool
			log.Infof("初始化%v图片池", imagePoolType)
			l.status.ImagePoolEnable = true
		}
	case "off":
		log.Debug("关闭图片池")
	default:
		log.Errorf("未知的图片池")
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
	if cfg.GetTemplateEnabled() {
		log.Infof("已启用模板")
		template.InitTemplateLoader()
	}
	config.GlobalConfig.OnConfigChange(func(in fsnotify.Event) {
		l.CronjobReload()
	})
}

func (l *Lsp) PostInit() {
}

func (l *Lsp) Serve(bot *bot.Bot) {
	bot.OnGroupMemberJoined(func(qqClient *client.QQClient, event *client.MemberJoinGroupEvent) {
		if err := localdb.Set(localdb.Key("OnGroupMemberJoined", event.Group.Code, event.Member.Uin, event.Member.JoinTime), "",
			localdb.SetExpireOpt(time.Minute*2), localdb.SetNoOverWriteOpt()); err != nil {
			return
		}
		m, _ := template.LoadAndExec("trigger.group.member_in.tmpl", map[string]interface{}{
			"group_code":  event.Group.Code,
			"group_name":  event.Group.Name,
			"member_code": event.Member.Uin,
			"member_name": event.Member.DisplayName(),
		})
		if m != nil {
			l.SendMsg(m, mt.NewGroupTarget(event.Group.Code))
		}
	})
	bot.OnGroupMemberLeaved(func(qqClient *client.QQClient, event *client.MemberLeaveGroupEvent) {
		if err := localdb.Set(localdb.Key("OnGroupMemberLeaved", event.Group.Code, event.Member.Uin, event.Member.JoinTime), "",
			localdb.SetExpireOpt(time.Minute*2), localdb.SetNoOverWriteOpt()); err != nil {
			return
		}
		m, _ := template.LoadAndExec("trigger.group.member_out.tmpl", map[string]interface{}{
			"group_code":  event.Group.Code,
			"group_name":  event.Group.Name,
			"member_code": event.Member.Uin,
			"member_name": event.Member.DisplayName(),
		})
		if m != nil {
			l.SendMsg(m, mt.NewGroupTarget(event.Group.Code))
		}
	})
	bot.OnGroupInvited(func(qqClient *client.QQClient, request *client.GroupInvitedRequest) {
		log := logger.WithFields(logrus.Fields{
			"GroupCode":   request.GroupCode,
			"GroupName":   request.GroupName,
			"InvitorUin":  request.InvitorUin,
			"InvitorNick": request.InvitorNick,
		})

		if l.PermissionStateManager.CheckBlockList(request.InvitorUin) {
			log.Debug("收到加群邀请，该用户在block列表中，将拒绝加群邀请")
			l.PermissionStateManager.AddBlockList(request.GroupCode, 0)
			request.Reject(false, "")
			return
		}

		fi := bot.FindFriend(request.InvitorUin)
		if fi == nil {
			log.Error("收到加群邀请，无法找到好友信息，将拒绝加群邀请")
			l.PermissionStateManager.AddBlockList(request.GroupCode, 0)
			request.Reject(false, "未找到阁下的好友信息，请添加好友进行操作")
			return
		}

		if l.PermissionStateManager.CheckAdmin(request.InvitorUin) {
			log.Info("收到管理员的加群邀请，将同意加群邀请")
			l.PermissionStateManager.DeleteBlockList(request.GroupCode)
			request.Accept()
			return
		}

		switch l.LspStateManager.GetCurrentMode() {
		case PrivateMode:
			log.Info("收到加群邀请，当前BOT处于私有模式，将拒绝加群邀请")
			l.PermissionStateManager.AddBlockList(request.GroupCode, 0)
			request.Reject(false, "当前BOT处于私有模式")
		case ProtectMode:
			if err := l.LspStateManager.SaveGroupInvitedRequest(request); err != nil {
				log.Errorf("收到加群邀请，但记录申请失败，将拒绝该申请，请将该问题反馈给开发者 - error %v", err)
				request.Reject(false, "内部错误")
			} else {
				log.Info("收到加群邀请，当前BOT处于审核模式，将保留加群邀请")
			}
		case PublicMode:
			request.Accept()
			l.PermissionStateManager.DeleteBlockList(request.GroupCode)
			log.Info("收到加群邀请，当前BOT处于公开模式，将接受加群邀请")
			m, _ := template.LoadAndExec("trigger.private.group_invited.tmpl", map[string]interface{}{
				"member_code": request.InvitorUin,
				"member_name": request.InvitorNick,
				"group_code":  request.GroupCode,
				"group_name":  request.GroupName,
				"command":     CommandMaps,
			})
			if m != nil {
				l.SendMsg(m, mt.NewPrivateTarget(request.InvitorUin))
			}
			if err := l.PermissionStateManager.GrantTargetRole(mt.NewGroupTarget(request.GroupCode), request.InvitorUin, permission.TargetAdmin); err != nil {
				if err != permission.ErrPermissionExist {
					log.Errorf("设置群管理员权限失败 - %v", err)
				}
			}
		default:
			// impossible
			log.Errorf("收到加群邀请，当前BOT处于未知模式，将拒绝加群邀请，请将该问题反馈给开发者")
			request.Reject(false, "内部错误")
		}
	})

	bot.OnNewFriendRequest(func(qqClient *client.QQClient, request *client.NewFriendRequest) {
		log := logger.WithFields(logrus.Fields{
			"RequesterUin":  request.RequesterUin,
			"RequesterNick": request.RequesterNick,
			"Message":       request.Message,
		})
		if l.PermissionStateManager.CheckBlockList(request.RequesterUin) {
			log.Info("收到好友申请，该用户在block列表中，将拒绝好友申请")
			request.Reject()
			return
		}
		switch l.LspStateManager.GetCurrentMode() {
		case PrivateMode:
			log.Info("收到好友申请，当前BOT处于私有模式，将拒绝好友申请")
			request.Reject()
		case ProtectMode:
			if err := l.LspStateManager.SaveNewFriendRequest(request); err != nil {
				log.Errorf("收到好友申请，但记录申请失败，将拒绝该申请，请将该问题反馈给开发者 - error %v", err)
				request.Reject()
			} else {
				log.Info("收到好友申请，当前BOT处于审核模式，将保留好友申请")
			}
		case PublicMode:
			log.Info("收到好友申请，当前BOT处于公开模式，将通过好友申请")
			request.Accept()
		default:
			// impossible
			log.Errorf("收到好友申请，当前BOT处于未知模式，将拒绝好友申请，请将该问题反馈给开发者")
			request.Reject()
		}
	})

	bot.OnNewFriendAdded(func(qqClient *client.QQClient, event *client.NewFriendEvent) {
		log := logger.WithFields(logrus.Fields{
			"Uin":      event.Friend.Uin,
			"Nickname": event.Friend.Nickname,
		})
		log.Info("添加新好友")

		l.LspStateManager.RWCover(func() error {
			requests, err := l.LspStateManager.ListNewFriendRequest()
			if err != nil {
				log.Errorf("ListNewFriendRequest error %v", err)
				return err
			}
			for _, req := range requests {
				if req.RequesterUin == event.Friend.Uin {
					l.LspStateManager.DeleteNewFriendRequest(req.RequestId)
				}
			}
			return nil
		})

		m, _ := template.LoadAndExec("trigger.private.new_friend_added.tmpl", map[string]interface{}{
			"member_code": event.Friend.Uin,
			"member_name": event.Friend.Nickname,
			"command":     CommandMaps,
		})
		if m != nil {
			l.SendMsg(m, mt.NewPrivateTarget(event.Friend.Uin))
		}
	})

	bot.OnJoinGroup(func(qqClient *client.QQClient, info *client.GroupInfo) {
		l.FreshIndex()
		log := logger.WithFields(logrus.Fields{
			"GroupCode":   info.Code,
			"MemberCount": info.MemberCount,
			"GroupName":   info.Name,
			"OwnerUin":    info.OwnerUin,
			"Memo":        info.Memo,
		})
		log.Info("进入新群聊")

		rename := config.GlobalConfig.GetString("bot.onJoinGroup.rename")
		if len(rename) > 0 {
			if len(rename) > 60 {
				rename = rename[:60]
			}
			minfo := info.FindMember(bot.Uin)
			if minfo != nil {
				minfo.EditCard(rename)
			}
		}

		l.LspStateManager.RWCover(func() error {
			requests, err := l.LspStateManager.ListGroupInvitedRequest()
			if err != nil {
				log.Errorf("ListGroupInvitedRequest error %v", err)
				return err
			}
			for _, req := range requests {
				if req.GroupCode == info.Code {
					if err = l.LspStateManager.DeleteGroupInvitedRequest(req.RequestId); err != nil {
						log.WithField("RequestId", req.RequestId).Errorf("DeleteGroupInvitedRequest error %v", err)
					}
					if err = l.PermissionStateManager.GrantTargetRole(mt.NewGroupTarget(info.Code), req.InvitorUin, permission.TargetAdmin); err != nil {
						if err != permission.ErrPermissionExist {
							log.WithField("target", req.InvitorUin).Errorf("设置群管理员权限失败 - %v", err)
						}
					}
				}
			}
			return nil
		})
	})

	bot.OnLeaveGroup(func(qqClient *client.QQClient, event *client.GroupLeaveEvent) {
		log := logger.WithField("GroupCode", event.Group.Code).
			WithField("GroupName", event.Group.Name).
			WithField("MemberCount", event.Group.MemberCount)
		groupTarge := mt.NewGroupTarget(event.Group.Code)
		for _, c := range concern.ListConcern() {
			_, ids, _, err := c.GetStateManager().ListConcernState(
				func(target mt.Target, id interface{}, p concern_type.Type) bool {
					return target.Equal(groupTarge)
				})
			if err != nil {
				log = log.WithField(fmt.Sprintf("%v订阅", c.Site()), "查询失败")
			} else {
				log = log.WithField(fmt.Sprintf("%v订阅", c.Site()), len(ids))
			}
		}
		if event.Operator == nil {
			log.Info("退出群聊")
		} else {
			log.Infof("被 %v 踢出群聊", event.Operator.DisplayName())
		}
		l.RemoveAllByTarget(mt.NewGroupTarget(event.Group.Code))
	})

	bot.OnGroupMessage(func(qqClient *client.QQClient, msg *message.GroupMessage) {
		if len(msg.Elements) <= 0 {
			return
		}
		if err := l.LspStateManager.SaveMessageImageUrl(msg.GroupCode, msg.Id, msg.Elements); err != nil {
			logger.Errorf("SaveMessageImageUrl failed %v", err)
		}
		if !l.started.Load() {
			return
		}
		cmd := NewLspGroupCommand(l, msg)
		if Debug {
			cmd.Debug()
		}
		if !l.LspStateManager.IsMuted(mt.NewGroupTarget(msg.GroupCode), bot.Uin) {
			go cmd.Execute()
		}
	})

	bot.GuildService.OnGuildChannelMessage(func(qqClient *client.QQClient, msg *message.GuildChannelMessage) {
		if len(msg.Elements) <= 0 {
			return
		}
		if !l.started.Load() {
			return
		}
		cmd := NewLspGuildChannelCommand(l, msg)
		go cmd.Execute()
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
		if err := l.LspStateManager.Muted(mt.NewGroupTarget(event.GroupCode), event.TargetUin, event.Time); err != nil {
			logger.Errorf("Muted failed %v", err)
		}
	})

	bot.OnPrivateMessage(func(qqClient *client.QQClient, msg *message.PrivateMessage) {
		if !l.started.Load() {
			return
		}
		if len(msg.Elements) == 0 {
			return
		}
		cmd := NewLspPrivateCommand(l, msg)
		if Debug {
			cmd.Debug()
		}
		go cmd.Execute()
	})
	bot.OnDisconnected(func(qqClient *client.QQClient, event *client.ClientDisconnectedEvent) {
		logger.Errorf("收到OnDisconnected事件 %v", event.Message)
		if config.GlobalConfig.GetString("bot.onDisconnected") == "exit" {
			logger.Fatalf("onDisconnected设置为exit，bot将自动退出")
		}
		if err := bot.ReLogin(event); err != nil {
			logger.Fatalf("重连时发生错误%v，bot将自动退出", err)
		}
	})

}

func (l *Lsp) PostStart(bot *bot.Bot) {
	l.FreshIndex()
	go func() {
		for range time.Tick(time.Second * 30) {
			l.FreshIndex()
		}
	}()
	l.CronjobReload()
	l.CronStart()
	concern.StartAll()
	l.started.Store(true)

	var newVersionChan = make(chan string, 1)
	go func() {
		newVersionChan <- CheckUpdate()
		for range time.Tick(time.Hour * 24) {
			newVersionChan <- CheckUpdate()
		}
	}()
	go l.NewVersionNotify(newVersionChan)

	logger.Infof("DDBOT启动完成")
	logger.Infof("D宝，一款真正人性化的单推BOT")
	if len(l.PermissionStateManager.ListAdmin()) == 0 {
		logger.Infof("您似乎正在部署全新的BOT，请通过qq对bot私聊发送<%v>(不含括号)获取管理员权限，然后私聊发送<%v>(不含括号)开始使用您的bot",
			l.CommandShowName(WhosyourdaddyCommand), l.CommandShowName(HelpCommand))
	}

}

func (l *Lsp) Start(bot *bot.Bot) {
	go l.ConcernNotify()
}

func (l *Lsp) Stop(bot *bot.Bot, wg *sync.WaitGroup) {
	defer wg.Done()
	if l.stop != nil {
		close(l.stop)
	}
	l.CronStop()
	concern.StopAll()

	l.wg.Wait()
	logger.Debug("等待所有推送发送完毕")
	l.notifyWg.Wait()
	logger.Debug("推送发送完毕")

	proxy_pool.Stop()
}

func (l *Lsp) NewVersionNotify(newVersionChan <-chan string) {
	defer func() {
		if err := recover(); err != nil {
			logger.WithField("stack", string(debug.Stack())).
				Errorf("new version notify recoverd %v", err)
			go l.NewVersionNotify(newVersionChan)
		}
	}()
	for newVersion := range newVersionChan {
		if newVersion == "" {
			continue
		}
		var newVersionNotify bool
		err := localdb.RWCover(func() error {
			key := localdb.DDBotReleaseKey()
			releaseVersion, err := localdb.Get(key, localdb.IgnoreNotFoundOpt())
			if err != nil {
				return err
			}
			if releaseVersion != newVersion {
				newVersionNotify = true
			}
			return localdb.Set(key, newVersion)
		})
		if err != nil {
			logger.Errorf("NewVersionNotify error %v", err)
			continue
		}
		if !newVersionNotify {
			continue
		}
		m := mmsg.NewMSG()
		m.Textf("DDBOT管理员您好，DDBOT有可用更新版本【%v】，请前往 https://github.com/Sora233/DDBOT/releases 查看详细信息\n\n", newVersion)
		m.Textf("如果您不想接收更新消息，请输入<%v>(不含括号)", l.CommandShowName(NoUpdateCommand))
		for _, admin := range l.PermissionStateManager.ListAdmin() {
			if localdb.Exist(localdb.DDBotNoUpdateKey(admin)) {
				continue
			}
			if localutils.GetBot().FindFriend(admin) == nil {
				continue
			}
			logger.WithField("Target", admin).Infof("new ddbot version notify")
			l.SendMsg(m, mt.NewPrivateTarget(admin))
		}
	}
}

func (l *Lsp) FreshIndex() {
	for _, c := range concern.ListConcern() {
		c.FreshIndex()
	}
	l.PermissionStateManager.FreshIndex()
	l.LspStateManager.FreshIndex()
}

func (l *Lsp) RemoveAllByTarget(target mt.Target) {
	for _, c := range concern.ListConcern() {
		c.GetStateManager().RemoveAllByTarget(target)
	}
	l.PermissionStateManager.RemoveAllByTarget(target)
}

func (l *Lsp) GetImageFromPool(options ...image_pool.OptionFunc) ([]image_pool.Image, error) {
	if l.pool == nil {
		return nil, image_pool.ErrNotInit
	}
	return l.pool.Get(options...)
}

func (l *Lsp) send(msg *message.SendingMessage, target mt.Target) interface{} {
	switch target.GetTargetType() {
	case mt.TargetGroup:
		return l.sendGroupMessage(target.(*mt.GroupTarget).TargetCode(), msg)
	case mt.TargetPrivate:
		return l.sendPrivateMessage(target.(*mt.PrivateTarget).TargetCode(), msg)
	case mt.TargetGuild:
		guildTarget := target.(*mt.GuildTarget)
		return l.sendGuildMessage(guildTarget.GuildId, guildTarget.ChannelId, msg)
	}
	panic("unknown target type")
}

// SendMsg 总是返回至少一个
func (l *Lsp) SendMsg(m *mmsg.MSG, target mt.Target) (res []interface{}) {
	msgs := m.ToMessage(target)
	if len(msgs) == 0 {
		switch target.GetTargetType() {
		case mt.TargetPrivate:
			res = append(res, &message.PrivateMessage{Id: -1})
		case mt.TargetGroup:
			res = append(res, &message.GroupMessage{Id: -1})
		case mt.TargetGuild:
			res = append(res, &message.GuildChannelMessage{Id: 0})
		}
		return
	}
	for idx, msg := range msgs {
		r := l.send(msg, target)
		res = append(res, r)
		var v int64
		if target.GetTargetType() == mt.TargetGuild {
			v = int64(reflect.ValueOf(r).Elem().FieldByName("Id").Uint())
		} else {
			v = reflect.ValueOf(r).Elem().FieldByName("Id").Int()
		}
		if v == -1 || v == 0 {
			break
		}
		if idx > 1 {
			time.Sleep(time.Millisecond * 300)
		}
	}
	return res
}

func (l *Lsp) GM(res []interface{}) []*message.GroupMessage {
	var result []*message.GroupMessage
	for _, r := range res {
		result = append(result, r.(*message.GroupMessage))
	}
	return result
}

func (l *Lsp) PM(res []interface{}) []*message.PrivateMessage {
	var result []*message.PrivateMessage
	for _, r := range res {
		result = append(result, r.(*message.PrivateMessage))
	}
	return result
}

func (l *Lsp) GCM(res []interface{}) []*message.GuildChannelMessage {
	var result []*message.GuildChannelMessage
	for _, r := range res {
		result = append(result, r.(*message.GuildChannelMessage))
	}
	return result
}

func (l *Lsp) sendGuildMessage(guildId uint64, channelId uint64, msg *message.SendingMessage) (res *message.GuildChannelMessage) {
	if bot.Instance == nil || !bot.Instance.Online.Load() {
		return &message.GuildChannelMessage{Id: 0, Elements: msg.Elements}
	}
	if msg == nil {
		logger.WithFields(localutils.GuildChannelLogFields(guildId, channelId)).Debug("send with nil message")
		return &message.GuildChannelMessage{Id: 0}
	}
	msg.Elements = localutils.MessageFilter(msg.Elements, func(element message.IMessageElement) bool {
		return element != nil
	})
	if len(msg.Elements) == 0 {
		logger.WithFields(localutils.GuildChannelLogFields(guildId, channelId)).Debug("send with empty message")
		return &message.GuildChannelMessage{Id: 0}
	}
	res, _ = bot.Instance.GuildService.SendGuildChannelMessage(guildId, channelId, msg)
	if res == nil || res.Id == 0 {
		logger.WithField("content", msgstringer.MsgToString(msg.Elements)).
			WithFields(localutils.GuildChannelLogFields(guildId, channelId)).
			Errorf("发送消息失败")
	}
	if res == nil {
		res = &message.GuildChannelMessage{Id: 0, Elements: msg.Elements}
	}
	if res == nil || res.Id == 0 {
		if msg.Count(func(e message.IMessageElement) bool {
			return e.Type() == message.At && e.(*message.AtElement).Target == 0
		}) > 0 {
			logger.WithField("content", msgstringer.MsgToString(msg.Elements)).
				WithFields(localutils.GuildChannelLogFields(guildId, channelId)).
				Errorf("发送群消息失败，可能是@全员次数用尽")
		} else {
			logger.WithField("content", msgstringer.MsgToString(msg.Elements)).
				WithFields(localutils.GuildChannelLogFields(guildId, channelId)).
				Errorf("发送群消息失败，可能是被禁言或者账号被风控")
		}
	}
	return res
}

func (l *Lsp) sendPrivateMessage(uin int64, msg *message.SendingMessage) (res *message.PrivateMessage) {
	if bot.Instance == nil || !bot.Instance.Online.Load() {
		return &message.PrivateMessage{Id: -1, Elements: msg.Elements}
	}
	if msg == nil {
		logger.WithFields(localutils.FriendLogFields(uin)).Debug("send with nil message")
		return &message.PrivateMessage{Id: -1}
	}
	msg.Elements = localutils.MessageFilter(msg.Elements, func(element message.IMessageElement) bool {
		return element != nil
	})
	if len(msg.Elements) == 0 {
		logger.WithFields(localutils.FriendLogFields(uin)).Debug("send with empty message")
		return &message.PrivateMessage{Id: -1}
	}
	res = bot.Instance.SendPrivateMessage(uin, msg)
	if res == nil || res.Id == -1 {
		logger.WithField("content", msgstringer.MsgToString(msg.Elements)).
			WithFields(localutils.GroupLogFields(uin)).
			Errorf("发送消息失败")
	}
	if res == nil {
		res = &message.PrivateMessage{Id: -1, Elements: msg.Elements}
	}
	return res
}

// sendGroupMessage 发送一条消息，返回值总是非nil，Id为-1表示发送失败
// miraigo偶尔发送消息会panic？！
func (l *Lsp) sendGroupMessage(groupCode int64, msg *message.SendingMessage, recovered ...bool) (res *message.GroupMessage) {
	defer func() {
		if e := recover(); e != nil {
			if len(recovered) == 0 {
				logger.WithField("content", msgstringer.MsgToString(msg.Elements)).
					WithField("stack", string(debug.Stack())).
					Errorf("sendGroupMessage panic recovered")
				res = l.sendGroupMessage(groupCode, msg, true)
			} else {
				logger.WithField("content", msgstringer.MsgToString(msg.Elements)).
					WithField("stack", string(debug.Stack())).
					Errorf("sendGroupMessage panic recovered but panic again %v", e)
				res = &message.GroupMessage{Id: -1, Elements: msg.Elements}
			}
		}
	}()

	if bot.Instance == nil || !bot.Instance.Online.Load() {
		return &message.GroupMessage{Id: -1, Elements: msg.Elements}
	}
	if l.LspStateManager.IsMuted(mt.NewGroupTarget(groupCode), bot.Instance.Uin) {
		logger.WithField("content", msgstringer.MsgToString(msg.Elements)).
			WithFields(localutils.GroupLogFields(groupCode)).
			Debug("BOT被禁言无法发送群消息")
		return &message.GroupMessage{Id: -1, Elements: msg.Elements}
	}
	if msg == nil {
		logger.WithFields(localutils.GroupLogFields(groupCode)).Debug("send with nil message")
		return &message.GroupMessage{Id: -1}
	}
	msg.Elements = localutils.MessageFilter(msg.Elements, func(element message.IMessageElement) bool {
		return element != nil
	})
	if len(msg.Elements) == 0 {
		logger.WithFields(localutils.GroupLogFields(groupCode)).Debug("send with empty message")
		return &message.GroupMessage{Id: -1}
	}
	res = bot.Instance.SendGroupMessage(groupCode, msg, cfg.GetFramMessage())
	if res == nil || res.Id == -1 {
		if msg.Count(func(e message.IMessageElement) bool {
			return e.Type() == message.At && e.(*message.AtElement).Target == 0
		}) > 0 {
			logger.WithField("content", msgstringer.MsgToString(msg.Elements)).
				WithFields(localutils.GroupLogFields(groupCode)).
				Errorf("发送群消息失败，可能是@全员次数用尽")
		} else {
			logger.WithField("content", msgstringer.MsgToString(msg.Elements)).
				WithFields(localutils.GroupLogFields(groupCode)).
				Errorf("发送群消息失败，可能是被禁言或者账号被风控")
		}
	}
	if res == nil {
		res = &message.GroupMessage{Id: -1, Elements: msg.Elements}
	}
	return res
}

var Instance = &Lsp{
	concernNotify:          concern.ReadNotifyChan(),
	stop:                   make(chan interface{}),
	status:                 NewStatus(),
	msgLimit:               semaphore.NewWeighted(3),
	PermissionStateManager: permission.NewStateManager(),
	LspStateManager:        NewStateManager(),
	cron:                   cron.New(cron.WithLogger(cron.VerbosePrintfLogger(cronLog))),
}

func init() {
	bot.RegisterModule(Instance)

	template.RegisterExtFunc("currentMode", func() string {
		return string(Instance.LspStateManager.GetCurrentMode())
	})
}
