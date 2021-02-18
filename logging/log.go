package logging

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/Mrs4s/MiraiGo/client"
	"github.com/Mrs4s/MiraiGo/message"

	"github.com/Logiase/MiraiGo-Template/bot"
	"github.com/Logiase/MiraiGo-Template/utils"
)

func init() {
	instance = &logging{}
	bot.RegisterModule(instance)
}

type logging struct {
}

func (m *logging) MiraiGoModule() bot.ModuleInfo {
	return bot.ModuleInfo{
		ID:       "internal.logging",
		Instance: instance,
	}
}

func (m *logging) Init() {
	// 初始化过程
	// 在此处可以进行 Module 的初始化配置
	// 如配置读取
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

var instance *logging

var logger = utils.GetModuleLogger("internal.logging")

func logGroupMessage(msg *message.GroupMessage) {
	logger.
		WithField("from", "GroupMessage").
		WithField("MessageID", msg.Id).
		WithField("MessageIID", msg.InternalId).
		WithField("GroupCode", msg.GroupCode).
		WithField("GroupName", msg.GroupName).
		WithField("SenderID", msg.Sender.Uin).
		WithField("SenderName", msg.Sender.DisplayName()).
		Info(msgToString(msg))
}

func logPrivateMessage(msg *message.PrivateMessage) {
	logger.
		WithField("from", "PrivateMessage").
		WithField("MessageID", msg.Id).
		WithField("MessageIID", msg.InternalId).
		WithField("SenderID", msg.Sender.Uin).
		WithField("SenderName", msg.Sender.DisplayName()).
		WithField("Target", msg.Target).
		Info(msg.ToString())
}

func logFriendMessageRecallEvent(event *client.FriendMessageRecalledEvent) {
	logger.
		WithField("from", "FriendsMessageRecall").
		WithField("MessageID", event.MessageId).
		WithField("SenderID", event.FriendUin).
		Info("friend message recall")
}

func logGroupMessageRecallEvent(event *client.GroupMessageRecalledEvent) {
	logger.
		WithField("from", "GroupMessageRecall").
		WithField("MessageID", event.MessageId).
		WithField("GroupCode", event.GroupCode).
		WithField("SenderID", event.AuthorUin).
		WithField("OperatorID", event.OperatorUin).
		Info("group message recall")
}

func logGroupMuteEvent(event *client.GroupMuteEvent) {
	logger.
		WithField("from", "GroupMute").
		WithField("GroupCode", event.GroupCode).
		WithField("OperatorID", event.OperatorUin).
		WithField("TargetID", event.TargetUin).
		WithField("MuteTime", event.Time).
		Info("group mute")
}

func logDisconnect(event *client.ClientDisconnectedEvent) {
	logger.
		WithField("from", "Disconnected").
		WithField("reason", event.Message).
		Warn("bot disconnected")
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
}

func msgToString(groupMsg *message.GroupMessage) (res string) {
	for _, elem := range groupMsg.Elements {
		switch e := elem.(type) {
		case *message.TextElement:
			res += e.Content
		case *message.ImageElement:
			res += "[Image:" + e.Filename + " " + e.Url + "]"
		case *message.FaceElement:
			res += "[" + e.Name + "]"
		case *message.GroupImageElement:
			res += "[Image: " + e.ImageId + " " + e.Url + "]"
		case *message.GroupFlashImgElement:
			// NOTE: ignore other components
			img, err := bot.Instance.QueryGroupImage(groupMsg.GroupCode, e.Md5, e.Size)
			if err != nil {
				return "[Image (flash img):" + e.Filename + fmt.Sprintf(" {ERROR_URL:%v} ]", err)
			}
			return "[Image (flash img):" + e.Filename + " " + img.Url + "]"
		case *message.AtElement:
			res += e.Display
		case *message.RedBagElement:
			res += "[RedBag:" + e.Title + "]"
		case *message.ReplyElement:
			res += "[Reply:" + strconv.FormatInt(int64(e.ReplySeq), 10) + "]"
		}
	}
	return
}
