package lsp

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	miraiBot "github.com/Logiase/MiraiGo-Template/bot"
	"github.com/Logiase/MiraiGo-Template/config"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Sora233/DDBOT/concern"
	"github.com/Sora233/DDBOT/lsp/bilibili"
	localdb "github.com/Sora233/DDBOT/lsp/buntdb"
	"github.com/Sora233/DDBOT/lsp/permission"
	"github.com/Sora233/sliceutil"
	"github.com/alecthomas/kong"
	"github.com/sirupsen/logrus"
	"github.com/tidwall/buntdb"
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
	if !strings.HasPrefix(c.GetCmd(), "/") {
		return
	}

	log := c.DefaultLogger().WithField("cmd", c.GetCmd()).WithField("args", c.GetArgs())

	if c.l.PermissionStateManager.CheckBlockList(c.uin()) {
		log.Debug("blocked")
		return
	}

	if !c.DebugCheck() {
		log.Debugf("debug mode, skip execute.")
		return
	}

	log.Debug("execute command")

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
	case "/watch":
		c.WatchCommand(false)
	case "/unwatch":
		c.WatchCommand(true)
	case "/enable":
		c.EnableCommand(false)
	case "/disable":
		c.EnableCommand(true)
	case "/log":
		if !c.l.PermissionStateManager.RequireAny(
			permission.AdminRoleRequireOption(c.uin()),
		) {
			c.noPermission()
			return
		}
		c.LogCommand()
	case "/list":
		c.ListCommand()
	default:
		c.textReply("阁下似乎输入了一个无法识别的命令，请注意BOT的大多数命令只能在群聊内生效。")
		log.Debug("no command matched")
	}
}

func (c *LspPrivateCommand) ListCommand() {
	log := c.DefaultLoggerWithCommand(ListCommand)
	log.Info("run list command")
	defer func() { log.Info("list command end") }()

	var listCmd struct {
		Site  string `optional:"" short:"s" default:"bilibili" help:"bilibili / douyu / youtube"`
		Type  string `optional:"" short:"t" default:"live" help:"news / live"`
		Group int64  `optional:"" short:"g" help:"要操作的QQ群号码"`
	}
	output := c.parseCommandSyntax(&listCmd, ListCommand)
	if output != "" {
		c.textReply(output)
	}
	if c.exit {
		return
	}
	site, watchType, err := c.ParseRawSiteAndType(listCmd.Site, listCmd.Type)
	if err != nil {
		log = log.WithField("args", c.GetArgs())
		log.Errorf("parse raw concern failed %v", err)
		c.textReply(fmt.Sprintf("参数错误 - %v", err))
		return
	}
	log = log.WithField("site", site).WithField("type", watchType)

	groupCode := listCmd.Group
	if err := c.checkGroupCode(groupCode); err != nil {
		c.textReply(err.Error())
		return
	}
	log = log.WithField("groupCode", groupCode)

	if c.l.PermissionStateManager.CheckGroupCommandDisabled(groupCode, ListCommand) {
		c.disabledReply()
		return
	}

	IList(c.NewCommandContext(log), groupCode, site, watchType)

}

func (c *LspPrivateCommand) WatchCommand(remove bool) {
	log := c.DefaultLoggerWithCommand(WatchCommand).WithField("unwatch", remove)
	log.Info("run watch command")
	defer func() { log.Info("watch command end") }()

	var (
		site      = bilibili.Site
		watchType = concern.BibiliLive
		err       error
	)

	var name string
	var Command string
	if remove {
		name = "unwatch"
		Command = UnwatchCommand
	} else {
		name = "watch"
		Command = WatchCommand
	}

	var watchCmd struct {
		Site  string `optional:"" short:"s" default:"bilibili" help:"bilibili / douyu / youtube / huya"`
		Type  string `optional:"" short:"t" default:"live" help:"news / live"`
		Group int64  `optional:"" short:"g" help:"要操作的QQ群号码"`
		Id    string `arg:""`
	}

	output := c.parseCommandSyntax(&watchCmd, name)
	if output != "" {
		c.textReply(output)
	}
	if c.exit {
		return
	}

	site, watchType, err = c.ParseRawSiteAndType(watchCmd.Site, watchCmd.Type)
	if err != nil {
		log = log.WithField("args", c.GetArgs())
		log.Errorf("parse raw concern failed %v", err)
		c.textReply(fmt.Sprintf("参数错误 - %v", err))
		return
	}
	log = log.WithField("site", site).WithField("type", watchType)

	id := watchCmd.Id
	groupCode := watchCmd.Group

	if err := c.checkGroupCode(groupCode); err != nil {
		c.textReply(err.Error())
		return
	}

	log = log.WithField("groupCode", groupCode)

	if c.l.PermissionStateManager.CheckGroupCommandDisabled(groupCode, Command) {
		c.disabledReply()
		return
	}

	if !c.l.PermissionStateManager.RequireAny(
		permission.AdminRoleRequireOption(c.uin()),
		permission.GroupAdminRoleRequireOption(groupCode, c.uin()),
		permission.QQAdminRequireOption(groupCode, c.uin()),
		permission.GroupCommandRequireOption(groupCode, c.uin(), WatchCommand),
		permission.GroupCommandRequireOption(groupCode, c.uin(), UnwatchCommand),
	) {
		c.noPermission()
		return
	}

	IWatch(c.NewCommandContext(log), groupCode, id, site, watchType, remove)
}

func (c *LspPrivateCommand) EnableCommand(disable bool) {
	log := c.DefaultLoggerWithCommand(EnableCommand).WithField("disable", disable)
	log.Info("run enable command")
	defer func() { log.Info("enable command end") }()

	var name string
	if disable {
		name = "disable"
	} else {
		name = "enable"
	}

	var enableCmd struct {
		Group   int64  `optional:"" short:"g" help:"要操作的QQ群号码"`
		Command string `arg:"" help:"command name"`
	}

	output := c.parseCommandSyntax(&enableCmd, name)
	if output != "" {
		c.textReply(output)
	}
	if c.exit {
		return
	}

	groupCode := enableCmd.Group

	if err := c.checkGroupCode(groupCode); err != nil {
		c.textReply(err.Error())
		return
	}

	log = log.WithField("groupCode", groupCode)

	if !c.l.PermissionStateManager.RequireAny(
		permission.AdminRoleRequireOption(c.uin()),
		permission.GroupAdminRoleRequireOption(groupCode, c.uin()),
	) {
		c.noPermission()
		return
	}

	IEnable(c.NewCommandContext(log), groupCode, enableCmd.Command, disable)
}

func (c *LspPrivateCommand) BlockCommand() {
	log := c.DefaultLoggerWithCommand(BlockCommand)
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
		} else if err == localdb.ErrKeyExist {
			log.Errorf("block failed - duplicate")
			c.textReply("失败 - 已经block过了")
		} else {
			log.Errorf("block failed err %v", err)
			c.textReply("失败")
		}
	} else {
		if err := c.l.PermissionStateManager.DeleteBlockList(blockCmd.Uin); err == nil {
			log.Info("unblocked")
			c.textReply("成功")
		} else if err == buntdb.ErrNotFound {
			log.Errorf("unblock failed - not exist")
			c.textReply("失败 - 该id未被block")
		} else {
			log.Errorf("unblock failed err %v", err)
			c.textReply("失败")
		}
	}
}

func (c *LspPrivateCommand) LogCommand() {
	log := c.DefaultLoggerWithCommand(LogCommand)
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
	log := c.DefaultLoggerWithCommand(PingCommand)
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
	log := c.DefaultLoggerWithCommand(HelpCommand)
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
		"订阅斗鱼乐言直播间（https://www.huya.com/xiaoleyan）：/watch -s huya -t live xiaoleyan\n" +
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
		"其他使用问题请在此提出：https://github.com/Sora233/DDBOT/discussions"
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

func (c *LspPrivateCommand) DefaultLoggerWithCommand(command string) *logrus.Entry {
	return c.DefaultLogger().WithField("command", command)
}

func (c *LspPrivateCommand) noPermission() *message.PrivateMessage {
	return c.textReply("权限不够")
}

func (c *LspPrivateCommand) disabledReply() *message.PrivateMessage {
	return c.textSend("该命令已被设置为disable，请设置enable后重试")
}

func (c *LspPrivateCommand) textSend(text string) *message.PrivateMessage {
	sendingMsg := message.NewSendingMessage()
	sendingMsg.Append(message.NewText(text))
	return c.send(sendingMsg)
}

func (c *LspPrivateCommand) textReply(text string) *message.PrivateMessage {
	sendingMsg := message.NewSendingMessage()
	// reply work bad
	//sendingMsg.Append(message.NewPrivateReply(c.msg))
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

func (c *LspPrivateCommand) NewCommandContext(log *logrus.Entry) *CommandContext {
	ctx := NewCommandContext()
	ctx.Lsp = c.l
	ctx.Log = log
	ctx.TextReply = func(text string) interface{} {
		return c.textReply(text)
	}
	ctx.Send = func(msg *message.SendingMessage) interface{} {
		return c.send(msg)
	}
	ctx.Reply = ctx.Send
	return ctx
}

func (c *LspPrivateCommand) checkGroupCode(groupCode int64) error {
	if groupCode == 0 {
		return fmt.Errorf("没有指定QQ群号码，请使用-g参数指定QQ群，例如对QQ群123456进行操作：%v %v %v", c.GetCmd(), "-g 123456", strings.Join(c.GetArgs(), " "))
	}
	group := c.bot.FindGroup(groupCode)
	if group == nil {
		return errors.New("没有找到该QQ群，请确认bot是否在群内")
	}
	member := group.FindMember(c.uin())
	if member == nil {
		return errors.New("没有在该群内找到您，请确认您是否在群内")
	}
	return nil
}
