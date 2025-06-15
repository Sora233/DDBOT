package lsp

import (
	"errors"
	"fmt"
	"math"
	"os"
	"reflect"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	"github.com/LagrangeDev/LagrangeGo/client"
	"github.com/LagrangeDev/LagrangeGo/client/event"
	"github.com/LagrangeDev/LagrangeGo/message"
	"github.com/fsnotify/fsnotify"
	jsoniter "github.com/json-iterator/go"
	"github.com/robfig/cron/v3"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"github.com/tidwall/buntdb"
	"golang.org/x/sync/semaphore"

	"github.com/Sora233/DDBOT/v2/image_pool"
	"github.com/Sora233/DDBOT/v2/image_pool/local_pool"
	"github.com/Sora233/DDBOT/v2/image_pool/lolicon_pool"
	localdb "github.com/Sora233/DDBOT/v2/lsp/buntdb"
	"github.com/Sora233/DDBOT/v2/lsp/cfg"
	"github.com/Sora233/DDBOT/v2/lsp/concern"
	"github.com/Sora233/DDBOT/v2/lsp/concern_type"
	"github.com/Sora233/DDBOT/v2/lsp/mmsg"
	"github.com/Sora233/DDBOT/v2/lsp/permission"
	"github.com/Sora233/DDBOT/v2/lsp/template"
	"github.com/Sora233/DDBOT/v2/lsp/version"
	"github.com/Sora233/DDBOT/v2/proxy_pool"
	"github.com/Sora233/DDBOT/v2/proxy_pool/local_proxy_pool"
	"github.com/Sora233/DDBOT/v2/proxy_pool/py"
	localutils "github.com/Sora233/DDBOT/v2/utils"
	"github.com/Sora233/DDBOT/v2/utils/msgstringer"
	"github.com/Sora233/MiraiGo-Template/bot"
	"github.com/Sora233/MiraiGo-Template/config"
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

	PermissionStateManager *permission.StateManager[uint32, uint32]
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
		if _, err := version.SetVersion(LspVersionName, LspSupportVersion); err != nil {
			log.Fatalf("警告：初始化LspVersion失败！")
		}
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
	cfg.ReloadCustomCommandPrefix()
	config.GlobalConfig.OnConfigChange(func(in fsnotify.Event) {
		go cfg.ReloadCustomCommandPrefix()
		l.CronjobReload()
	})
}

func (l *Lsp) PostInit() {
}

func (l *Lsp) reject(request *event.GroupInvite, reason string) {

}

func (l *Lsp) accept(request *event.GroupInvite) {

}

func (l *Lsp) Serve(bot *bot.Bot) {
	bot.GroupMemberJoinEvent.Subscribe(func(qqClient *client.QQClient, event *event.GroupMemberIncrease) {
		if err := localdb.Set(localdb.Key("OnGroupMemberJoined", event.GroupUin, event.UserUin, ""), "",
			localdb.SetExpireOpt(time.Minute*2), localdb.SetNoOverWriteOpt()); err != nil {
			return
		}
		groupInfo := lo.FromPtr(qqClient.GetCachedGroupInfo(event.GroupUin))
		memberInfo := lo.FromPtr(qqClient.GetCachedMemberInfo(event.UserUin, event.GroupUin))
		m, _ := template.LoadAndExec("trigger.group.member_in.tmpl", map[string]interface{}{
			"group_code":  event.GroupUin,
			"group_name":  groupInfo.GroupName,
			"member_code": event.UserUin,
			"member_name": memberInfo.DisplayName(),
		})
		if m != nil {
			l.SendMsg(m, mmsg.NewGroupTarget(event.GroupUin))
		}
	})
	bot.GroupMemberLeaveEvent.Subscribe(func(qqClient *client.QQClient, event *event.GroupMemberDecrease) {
		if err := localdb.Set(localdb.Key("OnGroupMemberLeaved", event.GroupUin, event.UserUin, ""), "",
			localdb.SetExpireOpt(time.Minute*2), localdb.SetNoOverWriteOpt()); err != nil {
			return
		}
		groupInfo := lo.FromPtr(qqClient.GetCachedGroupInfo(event.GroupUin))
		memberInfo := lo.FromPtr(qqClient.GetCachedMemberInfo(event.UserUin, event.GroupUin))
		m, _ := template.LoadAndExec("trigger.group.member_out.tmpl", map[string]interface{}{
			"group_code":  event.GroupUin,
			"group_name":  groupInfo.GroupName,
			"member_code": event.UserUin,
			"member_name": memberInfo.DisplayName(),
		})
		if m != nil {
			l.SendMsg(m, mmsg.NewGroupTarget(event.GroupUin))
		}
	})
	bot.GroupInvitedEvent.Subscribe(func(qqClient *client.QQClient, request *event.GroupInvite) {
		log := logger.WithFields(logrus.Fields{
			"GroupCode":   request.GroupUin,
			"GroupName":   request.GroupName,
			"InvitorUin":  request.InvitorUin,
			"InvitorNick": request.InvitorNick,
		})

		if l.PermissionStateManager.CheckBlockList(request.InvitorUin) {
			log.Debug("收到加群邀请，该用户在block列表中，将拒绝加群邀请")
			l.PermissionStateManager.AddBlockList(request.GroupUin, 0)
			l.reject(request, "")
			return
		}

		fi := bot.GetCachedFriendInfo(request.InvitorUin)
		if fi == nil {
			log.Error("收到加群邀请，无法找到好友信息，将拒绝加群邀请")
			l.PermissionStateManager.AddBlockList(request.GroupUin, 0)
			l.reject(request, "未找到阁下的好友信息，请添加好友进行操作")
			return
		}

		if l.PermissionStateManager.CheckAdmin(request.InvitorUin) {
			log.Info("收到管理员的加群邀请，将同意加群邀请")
			l.PermissionStateManager.DeleteBlockList(request.GroupUin)
			l.accept(request)
			return
		}

		switch l.LspStateManager.GetCurrentMode() {
		case PrivateMode:
			log.Info("收到加群邀请，当前BOT处于私有模式，将拒绝加群邀请")
			l.PermissionStateManager.AddBlockList(request.GroupUin, 0)
			l.reject(request, "当前BOT处于私有模式")
		case ProtectMode:
			if err := l.LspStateManager.SaveGroupInvitedRequest(request); err != nil {
				log.Errorf("收到加群邀请，但记录申请失败，将拒绝该申请，请将该问题反馈给开发者 - error %v", err)
				l.reject(request, "内部错误")
			} else {
				log.Info("收到加群邀请，当前BOT处于审核模式，将保留加群邀请")
			}
		case PublicMode:
			l.accept(request)
			l.PermissionStateManager.DeleteBlockList(request.GroupUin)
			log.Info("收到加群邀请，当前BOT处于公开模式，将接受加群邀请")
			m, _ := template.LoadAndExec("trigger.private.group_invited.tmpl", map[string]interface{}{
				"member_code": request.InvitorUin,
				"member_name": request.InvitorNick,
				"group_code":  request.GroupUin,
				"group_name":  request.GroupName,
				"command":     CommandMaps,
			})
			if m != nil {
				l.SendMsg(m, mmsg.NewPrivateTarget(request.InvitorUin))
			}
			if err := l.PermissionStateManager.GrantGroupRole(request.GroupUin, request.InvitorUin, permission.GroupAdmin); err != nil {
				if !errors.Is(err, permission.ErrPermissionExist) {
					log.Errorf("设置群管理员权限失败 - %v", err)
				}
			}
		default:
			// impossible
			log.Errorf("收到加群邀请，当前BOT处于未知模式，将拒绝加群邀请，请将该问题反馈给开发者")
			l.reject(request, "内部错误")
		}
	})

	bot.NewFriendRequestEvent.Subscribe(func(qqClient *client.QQClient, request *event.NewFriendRequest) {
		log := logger.WithFields(logrus.Fields{
			"RequesterUin":  request.SourceUin,
			"RequesterNick": request.SourceNick,
			"Message":       request.Msg,
		})
		if l.PermissionStateManager.CheckBlockList(request.SourceUin) {
			log.Info("收到好友申请，该用户在block列表中，将拒绝好友申请")
			bot.SetFriendRequest(false, request.SourceUID)
			return
		}
		switch l.LspStateManager.GetCurrentMode() {
		case PrivateMode:
			log.Info("收到好友申请，当前BOT处于私有模式，将拒绝好友申请")
			bot.SetFriendRequest(false, request.SourceUID)
		case ProtectMode:
			if err := l.LspStateManager.SaveNewFriendRequest(request); err != nil {
				log.Errorf("收到好友申请，但记录申请失败，将拒绝该申请，请将该问题反馈给开发者 - error %v", err)
				bot.SetFriendRequest(false, request.SourceUID)
			} else {
				log.Info("收到好友申请，当前BOT处于审核模式，将保留好友申请")
			}
		case PublicMode:
			log.Info("收到好友申请，当前BOT处于公开模式，将通过好友申请")
			bot.SetFriendRequest(true, request.SourceUID)
		default:
			// impossible
			log.Errorf("收到好友申请，当前BOT处于未知模式，将拒绝好友申请，请将该问题反馈给开发者")
			bot.SetFriendRequest(false, request.SourceUID)
		}
	})

	bot.NewFriendEvent.Subscribe(func(qqClient *client.QQClient, event *event.NewFriend) {
		log := logger.WithFields(logrus.Fields{
			"Uin":      event.FromUin,
			"Nickname": event.FromNick,
		})
		log.Info("添加新好友")

		l.LspStateManager.RWCover(func() error {
			requests, err := l.LspStateManager.ListNewFriendRequest()
			if err != nil {
				log.Errorf("ListNewFriendRequest error %v", err)
				return err
			}
			for _, req := range requests {
				if req.SourceUin == event.FromUin {
					l.LspStateManager.DeleteNewFriendRequest(req.Source)
				}
			}
			return nil
		})

		m, _ := template.LoadAndExec("trigger.private.new_friend_added.tmpl", map[string]interface{}{
			"member_code": event.FromUin,
			"member_name": event.FromNick,
			"command":     CommandMaps,
		})
		if m != nil {
			l.SendMsg(m, mmsg.NewPrivateTarget(event.FromUin))
		}
	})

	bot.GroupJoinEvent.Subscribe(func(qqClient *client.QQClient, info *event.GroupMemberIncrease) {
		if info.UserUin != bot.Uin {
			return
		}
		l.FreshIndex()
		group := lo.FromPtr(localutils.GetBot().FindGroup(info.GroupUin))
		log := logger.WithFields(logrus.Fields{
			"GroupCode":   info.GroupUin,
			"MemberCount": group.MemberCount,
			"GroupName":   group.GroupName,
			"OwnerUin":    group.GroupOwner,
		})
		log.Info("进入新群聊")

		rename := config.GlobalConfig.GetString("bot.onJoinGroup.rename")
		if len(rename) > 0 {
			if len(rename) > 60 {
				rename = rename[:60]
			}
			qqClient.SetGroupMemberName(info.GroupUin, bot.Uin, rename)
		}

		l.LspStateManager.RWCover(func() error {
			requests, err := l.LspStateManager.ListGroupInvitedRequest()
			if err != nil {
				log.Errorf("ListGroupInvitedRequest error %v", err)
				return err
			}
			for _, req := range requests {
				if req.GroupUin == info.GroupUin {
					if err = l.LspStateManager.DeleteGroupInvitedRequest(req.RequestSeq); err != nil {
						log.WithField("RequestSeq", req.RequestSeq).Errorf("DeleteGroupInvitedRequest error %v", err)
					}
					if err = l.PermissionStateManager.GrantGroupRole(info.GroupUin, req.InvitorUin, permission.GroupAdmin); err != nil {
						if !errors.Is(err, permission.ErrPermissionExist) {
							log.WithField("target", req.InvitorUin).Errorf("设置群管理员权限失败 - %v", err)
						}
					}
				}
			}
			return nil
		})
	})

	bot.GroupLeaveEvent.Subscribe(func(qqClient *client.QQClient, event *event.GroupMemberDecrease) {
		if event.UserUin != bot.Uin {
			return
		}
		groupInfo := lo.FromPtr(localutils.GetBot().FindGroup(event.GroupUin))
		log := logger.WithField("GroupCode", event.GroupUin).
			WithField("GroupName", groupInfo.GroupName).
			WithField("MemberCount", groupInfo.MemberCount)
		for _, c := range concern.ListConcern() {
			_, ids, _, err := c.GetStateManager().ListConcernState(
				func(groupCode uint32, id interface{}, p concern_type.Type) bool {
					return groupCode == event.GroupUin
				})
			if err != nil {
				log = log.WithField(fmt.Sprintf("%v订阅", c.Site()), "查询失败")
			} else {
				log = log.WithField(fmt.Sprintf("%v订阅", c.Site()), len(ids))
			}
		}
		if !event.IsKicked() {
			log.Info("退出群聊")
		} else {
			memberInfo := qqClient.GetCachedMemberInfo(event.OperatorUin, event.GroupUin)
			log.Infof("被 %v 踢出群聊", memberInfo.DisplayName())
		}
		l.RemoveAllByGroup(event.GroupUin)
	})

	bot.GroupNotifyEvent.Subscribe(func(qqClient *client.QQClient, ievent event.INotifyEvent) {
		switch event := ievent.(type) {
		case *event.GroupPokeEvent:
			data := map[string]interface{}{
				"member_code":   event.UserUID,
				"receiver_code": event.Receiver,
				"group_code":    event.GroupUin,
			}
			if gi := localutils.GetBot().FindGroup(event.GroupUin); gi != nil {
				data["group_name"] = gi.GroupName

				if fi := localutils.GetBot().FindGroupMember(gi.GroupUin, event.UserUin); fi != nil {
					data["member_name"] = fi.DisplayName()
				}
				if fi := localutils.GetBot().FindGroupMember(gi.GroupUin, event.Receiver); fi != nil {
					data["receiver_name"] = fi.DisplayName()
				}
			}
			m, _ := template.LoadAndExec("trigger.group.poke.tmpl", data)
			if m != nil {
				l.SendMsg(m, mmsg.NewGroupTarget(event.GroupUin))
			}
		}
	})

	bot.FriendNotifyEvent.Subscribe(func(qqClient *client.QQClient, ievent event.INotifyEvent) {
		switch event := ievent.(type) {
		case *event.FriendPokeEvent:
			if event.Receiver == localutils.GetBot().GetUin() {
				data := map[string]interface{}{
					"member_code": event.Sender,
				}
				if fi := localutils.GetBot().FindFriend(event.Sender); fi != nil {
					data["member_name"] = fi.Nickname
				}
				m, _ := template.LoadAndExec("trigger.private.poke.tmpl", data)
				if m != nil {
					l.SendMsg(m, mmsg.NewPrivateTarget(event.Sender))
				}
			}
		}
	})

	bot.GroupMessageEvent.Subscribe(func(qqClient *client.QQClient, msg *message.GroupMessage) {
		if len(msg.Elements) <= 0 {
			return
		}
		if !l.started.Load() {
			return
		}
		cmd := NewLspGroupCommand(l, msg)
		if Debug {
			cmd.Debug()
		}
		if !l.LspStateManager.IsMuted(msg.GroupUin, bot.Uin) {
			go cmd.Execute()
		}
	})

	bot.GroupMuteEvent.Subscribe(func(qqClient *client.QQClient, event *event.GroupMute) {
		if err := l.LspStateManager.Muted(event.GroupUin, event.UserUin, event.Duration); err != nil {
			logger.Errorf("Muted failed %v", err)
		}
	})

	bot.PrivateMessageEvent.Subscribe(func(qqClient *client.QQClient, msg *message.PrivateMessage) {
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
	bot.DisconnectedEvent.Subscribe(func(qqClient *client.QQClient, event *client.DisconnectedEvent) {
		if config.GlobalConfig.GetString("bot.onDisconnected") != "exit" && config.GlobalConfig.GetString("bot.onDisconnected") != "" {
			logger.Errorf("bot.onDisconnected配置已经不再支持")
		}
		logger.Fatalf("收到OnDisconnected事件 %v，bot将自动退出", event.Message)
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
		m.Textf("DDBOT管理员您好，DDBOT有可用更新版本【%v】，请前往 https://github.com/Sora233/DDBOT/v2/releases 查看详细信息\n\n", newVersion)
		m.Textf("如果您不想接收更新消息，请输入<%v>(不含括号)", l.CommandShowName(NoUpdateCommand))
		for _, admin := range l.PermissionStateManager.ListAdmin() {
			if localdb.Exist(localdb.DDBotNoUpdateKey(admin)) {
				continue
			}
			if localutils.GetBot().FindFriend(admin) == nil {
				continue
			}
			logger.WithField("Target", admin).Infof("new ddbot version notify")
			l.SendMsg(m, mmsg.NewPrivateTarget(admin))
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

func (l *Lsp) RemoveAllByGroup(groupCode uint32) {
	for _, c := range concern.ListConcern() {
		c.GetStateManager().RemoveAllByGroupCode(groupCode)
	}
	l.PermissionStateManager.RemoveAllByGroupCode(groupCode)
}

func (l *Lsp) GetImageFromPool(options ...image_pool.OptionFunc) ([]image_pool.Image, error) {
	if l.pool == nil {
		return nil, image_pool.ErrNotInit
	}
	return l.pool.Get(options...)
}

func (l *Lsp) send(msg *message.SendingMessage, target mmsg.Target) interface{} {
	switch target.TargetType() {
	case mmsg.TargetGroup:
		return l.sendGroupMessage(target.TargetCode(), msg)
	case mmsg.TargetPrivate:
		return l.sendPrivateMessage(target.TargetCode(), msg)
	}
	panic("unknown target type")
}

// SendMsg 总是返回至少一个
func (l *Lsp) SendMsg(m *mmsg.MSG, target mmsg.Target) (res []interface{}) {
	msgs := m.ToMessage(target)
	if len(msgs) == 0 {
		switch target.TargetType() {
		case mmsg.TargetPrivate:
			res = append(res, &message.PrivateMessage{ID: 0})
		case mmsg.TargetGroup:
			res = append(res, &message.GroupMessage{ID: 0})
		}
		return
	}
	for idx, msg := range msgs {
		r := l.send(msg, target)
		res = append(res, r)
		if reflect.ValueOf(r).Elem().FieldByName("ID").Uint() == math.MaxUint32 {
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

func (l *Lsp) sendPrivateMessage(uin uint32, msg *message.SendingMessage) (res *message.PrivateMessage) {
	if !localutils.GetBot().IsOnline() {
		return &message.PrivateMessage{ID: math.MaxUint32, Elements: msg.Elements}
	}

	if msg == nil {
		logger.WithFields(localutils.FriendLogFields(uin)).Debug("send with nil message")
		return &message.PrivateMessage{ID: math.MaxUint32}
	}
	msg.Elements = lo.Compact(msg.Elements)
	if len(msg.Elements) == 0 {
		logger.WithFields(localutils.FriendLogFields(uin)).Debug("send with empty message")
		return &message.PrivateMessage{ID: math.MaxUint32}
	}
	res, err := bot.QQClient.SendPrivateMessage(uin, msg.Elements)
	if err != nil || res == nil || res.ID == math.MaxUint32 {
		logger.WithField("content", msgstringer.MsgToString(msg.Elements)).
			WithError(err).
			WithFields(localutils.GroupLogFields(uin)).
			Errorf("发送消息失败")
	}
	if res == nil {
		res = &message.PrivateMessage{ID: math.MaxUint32, Elements: msg.Elements}
	}
	return res
}

// sendGroupMessage 发送一条消息，返回值总是非nil，Id为-1表示发送失败
// miraigo偶尔发送消息会panic？！
func (l *Lsp) sendGroupMessage(groupCode uint32, msg *message.SendingMessage, recovered ...bool) (res *message.GroupMessage) {
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
				res = &message.GroupMessage{ID: math.MaxUint32, Elements: msg.Elements}
			}
		}
	}()

	if !localutils.GetBot().IsOnline() {
		return &message.GroupMessage{ID: math.MaxUint32, Elements: msg.Elements}
	}
	if l.LspStateManager.IsMuted(groupCode, localutils.GetBot().GetUin()) {
		logger.WithField("content", msgstringer.MsgToString(msg.Elements)).
			WithFields(localutils.GroupLogFields(groupCode)).
			Debug("BOT被禁言无法发送群消息")
		return &message.GroupMessage{ID: math.MaxUint32, Elements: msg.Elements}
	}
	if msg == nil {
		logger.WithFields(localutils.GroupLogFields(groupCode)).Debug("send with nil message")
		return &message.GroupMessage{ID: math.MaxUint32}
	}
	msg.Elements = lo.Compact(msg.Elements)
	if len(msg.Elements) == 0 {
		logger.WithFields(localutils.GroupLogFields(groupCode)).Debug("send with empty message")
		return &message.GroupMessage{ID: math.MaxUint32}
	}
	res, err := bot.QQClient.SendGroupMessage(groupCode, msg.Elements)
	if err != nil || res == nil || res.ID == math.MaxUint32 {
		if lo.CountBy(msg.Elements, func(e message.IMessageElement) bool {
			return e.Type() == message.At && e.(*message.AtElement).TargetUin == 0
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
		res = &message.GroupMessage{ID: math.MaxUint32, Elements: msg.Elements}
	}
	return res
}

var Instance = &Lsp{
	concernNotify:          concern.ReadNotifyChan(),
	stop:                   make(chan interface{}),
	status:                 NewStatus(),
	msgLimit:               semaphore.NewWeighted(3),
	PermissionStateManager: permission.NewStateManager[uint32, uint32](),
	LspStateManager:        NewStateManager(),
	cron:                   cron.New(cron.WithLogger(cron.VerbosePrintfLogger(cronLog))),
}

func init() {
	bot.RegisterModule(Instance)

	template.RegisterExtFunc("currentMode", func() string {
		return string(Instance.LspStateManager.GetCurrentMode())
	})
}
