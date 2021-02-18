package lsp

import (
	miraiBot "github.com/Logiase/MiraiGo-Template/bot"
	"github.com/Logiase/MiraiGo-Template/config"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Sora233/sliceutil"
	"github.com/alecthomas/kong"
	"runtime/debug"
	"strings"
)

type LspPrivateCommand struct {
	msg *message.PrivateMessage

	*Runtime
}

func NewLspPrivateCommand(bot *miraiBot.Bot, l *Lsp, msg *message.PrivateMessage) *LspPrivateCommand {
	c := &LspPrivateCommand{
		msg:     msg,
		Runtime: NewRuntime(bot, l),
	}
	c.Parse(c.msg.Elements)
	return c
}

func (c *LspPrivateCommand) Execute() {
	defer func() {
		if err := recover(); err != nil {
			logger.WithField("stack", string(debug.Stack())).
				Errorf("panic recovered: %v", err)
			c.textSend("エラー発生")
		}
	}()
	if !c.DebugCheck() {
		return
	}
	if c.GetCmd() != "" && !strings.HasPrefix(c.GetCmd(), "/") {
		return
	}
	switch c.GetCmd() {
	case "/ping":
		c.PingCommand()
	case "/help":
		c.HelpCommand()
	}
}

func (c *LspPrivateCommand) PingCommand() {
	log := logger.WithField("uin", c.uin())
	log.Info("run ping command")
	defer log.Info("ping command end")

	output := c.parseCommandSyntax(&struct{}{}, PingCommand, kong.Description("reply a pong"), kong.UsageOnError())
	if output != "" {
		c.textSend(output)
	}
	if c.exit {
		return
	}
	c.textSend("pong")
}

func (c *LspPrivateCommand) HelpCommand() {
	log := logger.WithField("uin", c.uin())
	log.Info("run help command")
	defer log.Info("help command end")

	output := c.parseCommandSyntax(&struct{}{}, HelpCommand, kong.Description("print help message"))
	if output != "" {
		c.textSend(output)
	}
	if c.exit {
		return
	}

	help := "部分指令：\n" +
		"/watch 用于订阅推送，例如：\n" +
		"订阅b站uid为2的用户（https://space.bilibili.com/2）的直播信息：/watch -s bilibili -t live 2\n" +
		"订阅b站uid为2的用户的动态信息：/watch -s bilibili -t news 2\n" +
		"uid即b站用户空间末尾的数字\n" +
		"订阅斗鱼6655号直播间（https://www.douyu.com/6655）：/watch -s douyu -t live 6655\n" +
		"订阅油管karory（https://www.youtube.com/channel/UCGXspjV3G7ZSunbikIdp3EA）直播和视频：/watch -s youtube -t live UCGXspjV3G7ZSunbikIdp3EA\n" +
		"可以用相应的/unwatch命令取消订阅\n" +
		"取消订阅斗鱼6655直播间：/unwatch -s douyu -t live 6655\n" +
		"该系列命令默认情况下仅管理员可用\n" +
		"/list 用于查看当前订阅，例如：\n" +
		"查看当前b站订阅列表中正在直播的：/list -s bilibili -t live\n" +
		"/grant 用于管理员给其他成员设置权限，例如：\n" +
		"/grant -c watch 1234567 给qq号为1234567的用户使用watch命令的权限\n" +
		"设置的权限可以使用-d参数取消：\n" +
		"/grant -d -c watch 1234567 取消qq号为1234567的用户的watch命令权限\n" +
		"/enable和/disable 用于开启与禁用命令，例如：\n" +
		"/enable watch 将开启watch命令\n" +
		"/disable watch 将禁用watch命令，调用watch命令将不再有任何反应\n" +
		"其他使用问题请在此提出：https://github.com/Sora233/Sora233-MiraiGo/discussions"
	c.textSend(help)
}

func (c *LspPrivateCommand) DebugCheck() bool {
	var ok bool
	if c.debug {
		if sliceutil.Contains(config.GlobalConfig.GetStringSlice("debug.uin"), c.msg.Sender) {
			ok = true
		}
	} else {
		ok = true
	}
	return ok
}

func (c *LspPrivateCommand) textSend(text string) *message.PrivateMessage {
	sendingMsg := message.NewSendingMessage()
	sendingMsg.Append(message.NewText(text))
	return c.send(sendingMsg)
}
func (c *LspPrivateCommand) send(msg *message.SendingMessage) *message.PrivateMessage {
	return c.bot.SendPrivateMessage(c.uin(), msg)
}
func (c *LspPrivateCommand) sender() *message.Sender {
	return c.msg.Sender
}
func (c *LspPrivateCommand) uin() int64 {
	return c.sender().Uin
}
