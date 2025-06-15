package logging

import (
	"io"
	"path"
	"sync"
	"time"

	"github.com/LagrangeDev/LagrangeGo/client"
	"github.com/LagrangeDev/LagrangeGo/client/event"
	"github.com/LagrangeDev/LagrangeGo/message"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/rifflock/lfshook"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"

	localutils "github.com/Sora233/DDBOT/v2/utils"
	"github.com/Sora233/DDBOT/v2/utils/msgstringer"
	"github.com/Sora233/MiraiGo-Template/config"

	"github.com/Sora233/MiraiGo-Template/bot"
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
	if !config.GlobalConfig.GetBool("qq-logs.enabled") && !config.GlobalConfig.GetBool("qq-logs.enable") {
		qqlog.Out = io.Discard
	}
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
	logger.
		WithField("from", "GroupMessage").
		WithField("MessageID", msg.ID).
		WithField("MessageIID", msg.InternalID).
		WithField("GroupCode", msg.GroupUin).
		WithField("SenderUin", msg.Sender.Uin).
		WithField("SenderName", lo.CoalesceOrEmpty(msg.Sender.CardName, msg.Sender.Nickname)).
		Info(msg.ToString())
}

func logPrivateMessage(msg *message.PrivateMessage) {
	logger.WithFields(logrus.Fields{
		"From":       "PrivateMessage",
		"MessageID":  msg.ID,
		"MessageIID": msg.InternalID,
		"SenderID":   msg.Sender.Uin,
		"SenderName": lo.CoalesceOrEmpty(msg.Sender.CardName, msg.Sender.Nickname),
		"Target":     msg.Target,
	}).Info(msgstringer.MsgToString(msg.Elements))
}

func logFriendMessageRecallEvent(event *event.FriendRecall) {
	l := logger.WithFields(logrus.Fields{
		"From":      "FriendsMessageRecall",
		"MessageID": event.Sequence,
		"SenderID":  event.FromUin,
	})
	if fi := localutils.GetBot().FindFriend(event.FromUin); fi != nil {
		l = l.WithField("SenderName", fi.Nickname)
	}
	l.Info("好友消息撤回")
}

func logGroupMessageRecallEvent(event *event.GroupRecall) {
	l := logger.WithFields(localutils.GroupLogFields(event.GroupUin)).
		WithFields(logrus.Fields{
			"From":       "GroupMessageRecall",
			"MessageID":  event.Sequence,
			"SenderID":   event.UserUin,
			"OperatorID": event.OperatorUin,
		})
	if fi := localutils.GetBot().FindGroupMember(event.GroupUin, event.UserUin); fi != nil {
		l = l.WithField("SenderName", fi.Nickname)
	}
	if fi := localutils.GetBot().FindGroupMember(event.GroupUin, event.OperatorUin); fi != nil {
		l = l.WithField("OperatorName", fi.Nickname)
	}
	l.Info("群消息撤回")
}

func logGroupMuteEvent(event *event.GroupMute) {
	muteLogger := logger.WithFields(localutils.GroupLogFields(event.GroupUin)).
		WithFields(logrus.Fields{
			"From":        "GroupMute",
			"TargetUin":   event.UserUin,
			"OperatorUin": event.OperatorUin,
		})
	if event.UserUin != 0 {
		if fi := localutils.GetBot().FindGroupMember(event.GroupUin, event.UserUin); fi != nil {
			muteLogger = muteLogger.WithField("TargetName", fi.Nickname)
		}
	}
	if event.OperatorUin != 0 {
		if fi := localutils.GetBot().FindGroupMember(event.GroupUin, event.OperatorUin); fi != nil {
			muteLogger = muteLogger.WithField("OperatorName", fi.Nickname)
		}
	}
	if event.UserUin == 0 {
		if event.Duration != 0 {
			muteLogger.Debug("开启了全体禁言")
		} else {
			muteLogger.Debug("关闭了全体禁言")
		}
	} else {
		if event.Duration > 0 {
			muteLogger.Debug("用户被禁言")
		} else {
			muteLogger.Debug("用户被取消禁言")
		}
	}
}

func logDisconnect(event *client.DisconnectedEvent) {
	logger.WithFields(logrus.Fields{
		"From":   "Disconnected",
		"Reason": event.Message,
	}).Warn("bot断开链接")
}

func registerLog(b *bot.Bot) {
	b.GroupRecallEvent.Subscribe(func(qqClient *client.QQClient, event *event.GroupRecall) {
		logGroupMessageRecallEvent(event)
	})

	b.GroupMessageEvent.Subscribe(func(qqClient *client.QQClient, groupMessage *message.GroupMessage) {
		logGroupMessage(groupMessage)
	})

	b.GroupMuteEvent.Subscribe(func(qqClient *client.QQClient, event *event.GroupMute) {
		logGroupMuteEvent(event)
	})

	b.PrivateMessageEvent.Subscribe(func(qqClient *client.QQClient, privateMessage *message.PrivateMessage) {
		logPrivateMessage(privateMessage)
	})

	b.FriendRecallEvent.Subscribe(func(qqClient *client.QQClient, event *event.FriendRecall) {
		logFriendMessageRecallEvent(event)
	})

	b.DisconnectedEvent.Subscribe(func(qqClient *client.QQClient, event *client.DisconnectedEvent) {
		logDisconnect(event)
	})

	b.SelfGroupMessageEvent.Subscribe(func(qqClient *client.QQClient, groupMessage *message.GroupMessage) {
		logGroupMessage(groupMessage)
	})

	b.SelfPrivateMessageEvent.Subscribe(func(qqClient *client.QQClient, privateMessage *message.PrivateMessage) {
		logPrivateMessage(privateMessage)
	})
}
