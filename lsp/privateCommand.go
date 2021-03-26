package lsp

import (
	"bufio"
	"bytes"
	"fmt"
	miraiBot "github.com/Logiase/MiraiGo-Template/bot"
	"github.com/Logiase/MiraiGo-Template/config"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Sora233/Sora233-MiraiGo/lsp/permission"
	"github.com/Sora233/sliceutil"
	"github.com/alecthomas/kong"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"runtime/debug"
	"strings"
	"time"
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
	case "/block":
		if !c.l.PermissionStateManager.RequireAny(
			permission.AdminRoleRequireOption(c.uin()),
		) {
			c.noPermission()
			return
		}
		c.BlockCommand()
	case "/log":
		if !c.l.PermissionStateManager.RequireAny(
			permission.AdminRoleRequireOption(c.uin()),
		) {
			c.noPermission()
			return
		}
		c.LogCommand()
	}
}

func (c *LspPrivateCommand) BlockCommand() {
	log := c.DefaultLogger()
	log.Info("run block command")
	defer func() { log.Info("block command end") }()

	var blockCmd struct {
		Uin    int64 `arg:"" required:"" help:"the uin to block"`
		Days   int   `optional:""`
		Delete bool  `optional:"" short:"d"`
	}

	output := c.parseCommandSyntax(&blockCmd, BlockCommand)
	if output != "" {
		c.textReply(output)
	}
	if c.exit {
		return
	}

	if blockCmd.Uin == c.uin() {
		log.Errorf("can not block yourself")
		c.textReply("失败 - 不能block自己")
		return
	}

	if blockCmd.Days == 0 {
		blockCmd.Days = 7
	}

	log = log.WithField("target", blockCmd.Uin).WithField("days", blockCmd.Days).WithField("delete", blockCmd.Delete)

	if !blockCmd.Delete {
		if err := c.l.PermissionStateManager.AddBlockList(blockCmd.Uin, time.Duration(blockCmd.Days)*time.Hour*24); err == nil {
			log.Info("blocked")
			c.textReply("成功")
		} else {
			log.Errorf("block failed err %v", err)
			c.textReply("失败")
		}
	} else {
		if err := c.l.PermissionStateManager.DeleteBlockList(blockCmd.Uin); err == nil {
			log.Info("blocked")
			c.textReply("成功")
		} else {
			log.Errorf("unblock failed err %v", err)
			c.textReply("失败")
		}
	}
}

func (c *LspPrivateCommand) LogCommand() {
	log := c.DefaultLogger()
	log.Info("run log command")
	defer func() { log.Info("log command end") }()

	var logCmd struct {
		N       int       `arg:"" optional:"" help:"the number of lines from tail"`
		Date    time.Time `optional:"" short:"d" format:"2006-01-02"`
		Keyword string    `optional:"" short:"k" help:"the lines contains at lease one keyword"`
	}

	output := c.parseCommandSyntax(&logCmd, LogCommand)
	if output != "" {
		c.textSend(output)
	}
	if c.exit {
		return
	}
	if logCmd.N == 0 {
		logCmd.N = 10
	}
	if logCmd.Date.IsZero() {
		logCmd.Date = time.Now()
	}
	logName := fmt.Sprintf("%v.log", logCmd.Date.Format("2006-01-02"))
	b, err := ioutil.ReadFile("logs/" + logName)
	if err != nil {
		c.textSend(fmt.Sprintf("失败 - %v", err))
		return
	}
	var lines []string
	sc := bufio.NewScanner(bytes.NewReader(b))
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}
	if logCmd.N > len(lines) {
		logCmd.N = len(lines)
	}
	lines = lines[len(lines)-logCmd.N:]
	var filteredLines []string
	if len(logCmd.Keyword) != 0 {
		for _, line := range lines {
			if strings.Contains(line, logCmd.Keyword) {
				filteredLines = append(filteredLines, line)
			}
		}
	} else {
		filteredLines = lines[:]
	}
	if len(filteredLines) == 0 {
		c.textSend("无结果")
	} else {
		c.textSend(strings.Join(filteredLines, "\n"))
	}
}

func (c *LspPrivateCommand) PingCommand() {
	log := c.DefaultLogger()
	log.Info("run ping command")
	defer func() { log.Info("ping command end") }()

	output := c.parseCommandSyntax(&struct{}{}, PingCommand, kong.Description("reply a pong"), kong.UsageOnError())
	if output != "" {
		c.textReply(output)
	}
	if c.exit {
		return
	}
	c.textReply("pong")
}

func (c *LspPrivateCommand) HelpCommand() {
	log := c.DefaultLogger()
	log.Info("run help command")
	defer func() { log.Info("help command end") }()

	output := c.parseCommandSyntax(&struct{}{}, HelpCommand, kong.Description("print help message"))
	if output != "" {
		c.textReply(output)
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
		"订阅油管karory（https://www.youtube.com/channel/UCGXspjV3G7ZSunbikIdp3EA）直播和预约直播：/watch -s youtube -t live UCGXspjV3G7ZSunbikIdp3EA\n" +
		"可以用相应的/unwatch命令取消订阅\n" +
		"取消订阅斗鱼6655直播间：/unwatch -s douyu -t live 6655\n" +
		"该系列命令默认情况下仅管理员可用\n" +
		"/list 用于查看当前订阅，例如：\n" +
		"查看当前b站订阅列表中正在直播的：/list -s bilibili -t live\n" +
		"/grant 用于管理员给其他成员设置权限，例如：\n" +
		"给qq号为1234567的用户使用watch命令的权限：/grant -c watch 1234567\n" +
		"设置的权限可以使用-d参数取消：\n" +
		"取消qq号为1234567的用户的watch命令权限：/grant -d -c watch 1234567\n" +
		"/enable和/disable 用于开启与禁用命令，例如：\n" +
		"开启watch命令：/enable watch\n" +
		"禁用watch命令，调用watch命令将不再有任何反应：/disable watch\n" +
		"注意，bot只会在群聊内工作，私聊无法生效\n" +
		"详细使用介绍及样例，请查看https://github.com/Sora233/DDBOT/blob/master/EXAMPLE.md\n" +
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

func (c *LspPrivateCommand) DefaultLogger() *logrus.Entry {
	return logger.WithField("Uin", c.uin()).WithField("Name", c.name())
}

func (c *LspPrivateCommand) noPermission() *message.PrivateMessage {
	return c.textReply("权限不够")
}

func (c *LspPrivateCommand) textSend(text string) *message.PrivateMessage {
	sendingMsg := message.NewSendingMessage()
	sendingMsg.Append(message.NewText(text))
	return c.send(sendingMsg)
}

func (c *LspPrivateCommand) textReply(text string) *message.PrivateMessage {
	sendingMsg := message.NewSendingMessage()
	sendingMsg.Append(&message.ReplyElement{
		ReplySeq: c.msg.Id,
		Sender:   c.uin(),
		Time:     c.msg.Time,
		Elements: c.msg.Elements,
	})
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

func (c *LspPrivateCommand) name() string {
	return c.sender().DisplayName()
}
