package logging

import (
	"github.com/Mrs4s/MiraiGo/client"
	"github.com/Mrs4s/MiraiGo/message"
	localutils "github.com/Sora233/DDBOT/utils"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
	"path"
	"sync"
	"time"

	"github.com/Logiase/MiraiGo-Template/bot"
)

const moduleId = "ddbot.logging"

func init() {
	instance = &logging{}
	bot.RegisterModule(instance)
}

type logging struct {
}

var instance *logging

var logger *logrus.Entry

func (m *logging) MiraiGoModule() bot.ModuleInfo {
	return bot.ModuleInfo{
		ID:       moduleId,
		Instance: instance,
	}
}

func (m *logging) Init() {
	// create a new logger for qq log
	writer, err := rotatelogs.New(
		path.Join("qq-logs", "%Y-%m-%d.log"),
		rotatelogs.WithMaxAge(7*24*time.Hour),
		rotatelogs.WithRotationTime(24*time.Hour),
	)
	if err != nil {
		logrus.WithError(err).Error("unable to write logs")
		return
	}
	qqlog := logrus.New()
	qqlog.AddHook(lfshook.NewHook(writer, &logrus.TextFormatter{
		FullTimestamp:    true,
		PadLevelText:     true,
		QuoteEmptyFields: true,
		ForceQuote:       true,
	}))
	qqlog.SetLevel(logrus.DebugLevel)
	logger = qqlog.WithField("module", moduleId)
}

func (m *logging) PostInit() {
	// 第二次初始化
	// 再次过程中可以进行跨Module的动作
	// 如通用数据库等等
}

func (m *logging) Serve(b *bot.Bot) {
	// 注册服务函数部分
	registerLog(b)
}

func (m *logging) Start(b *bot.Bot) {
	// 此函数会新开携程进行调用
	// ```go
	// 		go exampleModule.Start()
	// ```

	// 可以利用此部分进行后台操作
	// 如http服务器等等
}

func (m *logging) Stop(b *bot.Bot, wg *sync.WaitGroup) {
	// 别忘了解锁
	defer wg.Done()
	// 结束部分
	// 一般调用此函数时，程序接收到 os.Interrupt 信号
	// 即将退出
	// 在此处应该释放相应的资源或者对状态进行保存
}

func logGroupMessage(msg *message.GroupMessage) {
	logger.WithFields(localutils.GroupLogFields(msg.GroupCode)).
		WithFields(logrus.Fields{
			"From":       "GroupMessage",
			"MessageID":  msg.Id,
			"MessageIID": msg.InternalId,
			"SenderID":   msg.Sender.Uin,
			"SenderName": msg.Sender.DisplayName(),
		}).Info(localutils.MsgToString(msg.Elements))
}

func logPrivateMessage(msg *message.PrivateMessage) {
	logger.WithFields(logrus.Fields{
		"From":       "PrivateMessage",
		"MessageID":  msg.Id,
		"MessageIID": msg.InternalId,
		"SenderID":   msg.Sender.Uin,
		"SenderName": msg.Sender.DisplayName(),
		"Target":     msg.Target,
	}).Info(localutils.MsgToString(msg.Elements))
}

func logFriendMessageRecallEvent(event *client.FriendMessageRecalledEvent) {
	logger.WithFields(logrus.Fields{
		"From":      "FriendsMessageRecall",
		"MessageID": event.MessageId,
		"SenderID":  event.FriendUin,
	}).Info("好友消息撤回")
}

func logGroupMessageRecallEvent(event *client.GroupMessageRecalledEvent) {
	logger.WithFields(localutils.GroupLogFields(event.GroupCode)).
		WithFields(logrus.Fields{
			"From":       "GroupMessageRecall",
			"MessageID":  event.MessageId,
			"SenderID":   event.AuthorUin,
			"OperatorID": event.OperatorUin,
		}).Info("群消息撤回")
}

func logGroupMuteEvent(event *client.GroupMuteEvent) {
	muteLogger := logger.WithFields(localutils.GroupLogFields(event.GroupCode)).
		WithFields(logrus.Fields{
			"From":        "GroupMute",
			"TargetUin":   event.TargetUin,
			"OperatorUin": event.OperatorUin,
		})
	if event.TargetUin == 0 {
		if event.Time != 0 {
			muteLogger.Debug("开启了全体禁言")
		} else {
			muteLogger.Debug("关闭了全体禁言")
		}
	} else {
		gi := bot.Instance.FindGroup(event.GroupCode)
		var mi *client.GroupMemberInfo
		if gi != nil {
			mi = gi.FindMember(event.TargetUin)
			if mi != nil {
				muteLogger = muteLogger.WithField("TargetName", mi.DisplayName())
			}
			mi = gi.FindMember(event.OperatorUin)
			if mi != nil {
				muteLogger = muteLogger.WithField("OperatorName", mi.DisplayName())
			}
		}
		if event.Time > 0 {
			muteLogger.Debug("用户被禁言")
		} else {
			muteLogger.Debug("用户被取消禁言")
		}
	}
}

func logDisconnect(event *client.ClientDisconnectedEvent) {
	logger.WithFields(logrus.Fields{
		"From":   "Disconnected",
		"Reason": event.Message,
	}).Warn("bot断开链接")
}

func registerLog(b *bot.Bot) {
	b.OnGroupMessageRecalled(func(qqClient *client.QQClient, event *client.GroupMessageRecalledEvent) {
		logGroupMessageRecallEvent(event)
	})

	b.OnGroupMessage(func(qqClient *client.QQClient, groupMessage *message.GroupMessage) {
		logGroupMessage(groupMessage)
	})

	b.OnGroupMuted(func(qqClient *client.QQClient, event *client.GroupMuteEvent) {
		logGroupMuteEvent(event)
	})

	b.OnPrivateMessage(func(qqClient *client.QQClient, privateMessage *message.PrivateMessage) {
		logPrivateMessage(privateMessage)
	})

	b.OnFriendMessageRecalled(func(qqClient *client.QQClient, event *client.FriendMessageRecalledEvent) {
		logFriendMessageRecallEvent(event)
	})

	b.OnDisconnected(func(qqClient *client.QQClient, event *client.ClientDisconnectedEvent) {
		logDisconnect(event)
	})

	b.OnSelfGroupMessage(func(qqClient *client.QQClient, groupMessage *message.GroupMessage) {
		logGroupMessage(groupMessage)
	})

	b.OnSelfPrivateMessage(func(qqClient *client.QQClient, privateMessage *message.PrivateMessage) {
		logPrivateMessage(privateMessage)
	})
}
