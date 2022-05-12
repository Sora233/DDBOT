package lsp

import (
	"fmt"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Sora233/DDBOT/lsp/concern"
	"github.com/Sora233/DDBOT/lsp/concern_type"
	"github.com/Sora233/DDBOT/lsp/mmsg"
	"github.com/Sora233/DDBOT/lsp/mmsg/mt"
	"github.com/Sora233/DDBOT/lsp/template"
	"github.com/Sora233/DDBOT/utils"
	"github.com/alecthomas/kong"
	"github.com/sirupsen/logrus"
	"runtime/debug"
	"strings"
)

type LspGuildChannelCommand struct {
	msg *message.GuildChannelMessage

	*Runtime
}

func NewLspGuildChannelCommand(l *Lsp, msg *message.GuildChannelMessage) *LspGuildChannelCommand {
	c := &LspGuildChannelCommand{
		msg:     msg,
		Runtime: NewRuntime(l),
	}
	c.Parse(c.msg.Elements)
	return c
}

func (gc *LspGuildChannelCommand) Execute() {
	defer func() {
		if err := recover(); err != nil {
			logger.WithField("stack", string(debug.Stack())).
				Errorf("panic recovered: %v", err)
			gc.textSend("エラー発生：看到该信息表示BOT出了一些问题，该问题已记录")
		}
	}()

	if len(gc.CommandName()) == 0 {
		return
	}

	log := gc.DefaultLogger().WithField("cmd", gc.GetCmdArgs())

	log.Debug("execute command")

	switch gc.CommandName() {
	case LspCommand:
		if gc.requireNotDisable(LspCommand) {
			gc.LspCommand()
		}
	case HelpCommand:
		if gc.requireNotDisable(HelpCommand) {
			gc.HelpCommand()
		}
	case WatchCommand:
		if gc.requireNotDisable(WatchCommand) {
			gc.WatchCommand(false)
		}
	case UnwatchCommand:
		if gc.requireNotDisable(WatchCommand) {
			gc.WatchCommand(true)
		}
	case ListCommand:
		if gc.requireNotDisable(ListCommand) {
			gc.ListCommand()
		}
	default:
		if CheckCustomGroupCommand(gc.CommandName()) {
			func() {
				log := gc.DefaultLoggerWithCommand(gc.CommandName()).WithField("CustomCommand", true)
				log.Infof("run %v command", gc.CommandName())
				defer func() { log.Infof("%v command end", gc.CommandName()) }()
				gc.sendChain(
					gc.templateMsg(fmt.Sprintf("custom.command.group.%s.tmpl", gc.CommandName()), map[string]interface{}{
						"cmd":  gc.CommandName(),
						"args": gc.GetArgs(),
					}),
				)
			}()
		} else {
			log.Debug("no command matched")
		}
	}
}

func (gc *LspGuildChannelCommand) WatchCommand(remove bool) {
	var (
		site      string
		watchType = concern_type.Type("live")
		err       error
	)

	log := gc.DefaultLoggerWithCommand(gc.CommandName())
	log.Infof("run %v command", gc.CommandName())
	defer func() { log.Infof("%v command end", gc.CommandName()) }()

	var watchCmd struct {
		Site string `optional:"" short:"s" default:"bilibili" help:"网站参数"`
		Type string `optional:"" short:"t" default:"" help:"类型参数"`
		Id   string `arg:""`
	}

	_, output := gc.parseCommandSyntax(&watchCmd, gc.CommandName(), kong.Description(
		fmt.Sprintf("当前支持的网站：%v", strings.Join(concern.ListSite(), "/"))),
	)
	if output != "" {
		gc.textReply(output)
	}
	if gc.exit {
		return
	}

	site, watchType, err = gc.ParseRawSiteAndType(watchCmd.Site, watchCmd.Type)
	if err != nil {
		log = log.WithField("args", gc.GetArgs())
		log.Errorf("ParseRawSiteAndType failed %v", err)
		gc.textReply(fmt.Sprintf("参数错误 - %v", err))
		return
	}
	log = log.WithField("site", site).WithField("type", watchType)

	id := watchCmd.Id

	IWatch(gc.NewMessageContext(log), mt.NewGuildTarget(gc.guildChannelID()), id, site, watchType, remove)
}

func (gc *LspGuildChannelCommand) LspCommand() {
	log := gc.DefaultLoggerWithCommand(gc.CommandName())
	log.Infof("run %v command", gc.CommandName())
	defer func() { log.Infof("%v command end", gc.CommandName()) }()

	var lspCmd struct{}
	_, output := gc.parseCommandSyntax(&lspCmd, gc.CommandName())
	if output != "" {
		gc.textReply(output)
	}
	if gc.exit {
		return
	}
	gc.sendChain(gc.templateMsg("command.guild.lsp.tmpl", nil))
}

func (gc *LspGuildChannelCommand) ListCommand() {
	log := gc.DefaultLoggerWithCommand(gc.CommandName())
	log.Infof("run %v command", gc.CommandName())
	defer func() { log.Infof("%v command end", gc.CommandName()) }()

	var listCmd struct {
		Site string `optional:"" short:"s" help:"网站参数"`
	}
	_, output := gc.parseCommandSyntax(&listCmd, gc.CommandName())
	if output != "" {
		gc.textReply(output)
	}
	if gc.exit {
		return
	}

	IList(gc.NewMessageContext(log), mt.NewGuildTarget(gc.guildChannelID()), listCmd.Site)
}

func (gc *LspGuildChannelCommand) HelpCommand() {
	log := gc.DefaultLoggerWithCommand(gc.CommandName())
	log.Infof("run %v command", gc.CommandName())
	defer func() { log.Infof("%v command end", gc.CommandName()) }()

	_, output := gc.parseCommandSyntax(&struct{}{}, gc.CommandName(), kong.Description("显示帮助信息"))
	if output != "" {
		gc.textReply(output)
	}
	if gc.exit {
		return
	}
	gc.sendChain(gc.templateMsg("command.guild.help.tmpl", nil))
}

func (gc *LspGuildChannelCommand) DefaultLogger() *logrus.Entry {
	return logger.WithFields(logrus.Fields{
		"Uin":  gc.uin(),
		"Name": gc.name(),
	}).WithFields(utils.GuildChannelLogFields(gc.guildChannelID()))
}

func (gc *LspGuildChannelCommand) DefaultLoggerWithCommand(command string) *logrus.Entry {
	return gc.DefaultLogger().WithField("Command", command)
}

// explicit defined and enabled
func (gc *LspGuildChannelCommand) groupEnabled(command string) bool {
	return gc.Runtime.groupEnabled(mt.NewGuildTarget(gc.guildChannelID()), command)
}

// explicit defined and disabled
func (gc *LspGuildChannelCommand) groupDisabled(command string) bool {
	return gc.Runtime.groupDisabled(mt.NewGuildTarget(gc.guildChannelID()), command)
}

func (gc *LspGuildChannelCommand) requireEnable(command string) bool {
	if !gc.groupEnabled(command) {
		gc.DefaultLoggerWithCommand(command).Debug("not enable")
		return false
	}
	return true
}

func (gc *LspGuildChannelCommand) requireNotDisable(command string) bool {
	if gc.groupDisabled(command) {
		gc.DefaultLoggerWithCommand(command).Debug("disabled")
		return false
	}
	return true
}

func (gc *LspGuildChannelCommand) commonTemplateData() map[string]interface{} {
	return map[string]interface{}{
		"msg":          gc.msg,
		"member_code":  gc.uin(),
		"member_name":  gc.name(),
		"channel_id":   int64(gc.channelID()),
		"guild_id":     int64(gc.guildID()),
		"channel_name": utils.GetBot().FindChannelName(gc.guildChannelID()),
		"guild_name":   utils.GetBot().FindGuildName(gc.guildID()),
		"command":      CommandMaps,
	}
}

func (gc *LspGuildChannelCommand) templateMsg(name string, data map[string]interface{}) *mmsg.MSG {
	commonData := gc.commonTemplateData()
	for k, v := range data {
		commonData[k] = v
	}
	m, err := template.LoadAndExec(name, commonData)
	if err != nil {
		logger.Errorf("LoadAndExec error %v", err)
		gc.textReply(fmt.Sprintf("错误 - %v", err))
		return nil
	}
	return m
}

func (gc *LspGuildChannelCommand) guildID() uint64 {
	return gc.msg.GuildId
}

func (gc *LspGuildChannelCommand) channelID() uint64 {
	return gc.msg.ChannelId
}

func (gc *LspGuildChannelCommand) guildChannelID() (uint64, uint64) {
	return gc.guildID(), gc.channelID()
}

func (gc *LspGuildChannelCommand) sender() *message.GuildSender {
	return gc.msg.Sender
}
func (gc *LspGuildChannelCommand) uin() int64 {
	return int64(gc.sender().TinyId)
}

func (gc *LspGuildChannelCommand) name() string {
	return gc.sender().Nickname
}

func (gc *LspGuildChannelCommand) textSend(text string) *message.GuildChannelMessage {
	return gc.send(mmsg.NewText(text))
}

func (gc *LspGuildChannelCommand) textReply(text string) *message.GuildChannelMessage {
	return gc.reply(mmsg.NewText(text))
}

func (gc *LspGuildChannelCommand) textReplyF(format string, args ...interface{}) *message.GuildChannelMessage {
	return gc.send(mmsg.NewTextf(format, args...))
}

func (gc *LspGuildChannelCommand) reply(msg *mmsg.MSG) *message.GuildChannelMessage {
	m := mmsg.NewMSG()
	m.Append(utils.NewGuildChannelReply(gc.msg))
	m.Append(msg.Elements()...)
	return gc.send(m)
}

func (gc *LspGuildChannelCommand) send(msg *mmsg.MSG) *message.GuildChannelMessage {
	return gc.l.GCM(gc.l.SendMsg(msg, mt.NewGuildTarget(gc.guildChannelID())))[0]
}

func (gc *LspGuildChannelCommand) sendChain(msg *mmsg.MSG) []*message.GuildChannelMessage {
	return gc.l.GCM(gc.l.SendMsg(msg, mt.NewGuildTarget(gc.guildChannelID())))
}

func (gc *LspGuildChannelCommand) noPermissionReply() *message.GuildChannelMessage {
	return gc.textReply("权限不够")
}

func (gc *LspGuildChannelCommand) globalDisabledReply() *message.GuildChannelMessage {
	return gc.textReply("无法操作该命令，该命令已被管理员禁用")
}

func (gc *LspGuildChannelCommand) NewMessageContext(log *logrus.Entry) *MessageContext {
	ctx := NewMessageContext()
	ctx.Source = mt.TargetGroup
	ctx.Lsp = gc.l
	ctx.Log = log
	ctx.SendFunc = func(m *mmsg.MSG) interface{} {
		return gc.send(m)
	}
	ctx.ReplyFunc = func(m *mmsg.MSG) interface{} {
		return gc.send(m)
	}
	ctx.NoPermissionReplyFunc = func() interface{} {
		ctx.Log.Debugf("no permission")
		if !gc.l.PermissionStateManager.CheckTargetSilence(mt.NewGuildTarget(gc.guildChannelID())) {
			return gc.noPermissionReply()
		}
		return nil
	}
	ctx.NotImplReplyFunc = func() interface{} {
		ctx.Log.Errorf("not impl")
		gc.textReply("暂未实现，你可以催作者GKD")
		return nil
	}
	ctx.DisabledReply = func() interface{} {
		ctx.Log.Debugf("disabled")
		return nil
	}
	ctx.GlobalDisabledReply = func() interface{} {
		ctx.Log.Debugf("global disabled")
		if !gc.l.PermissionStateManager.CheckTargetSilence(mt.NewGuildTarget(gc.guildChannelID())) {
			return gc.globalDisabledReply()
		}
		return nil
	}
	ctx.Sender = NewMessageSender(int64(gc.msg.Sender.TinyId), gc.msg.Sender.Nickname)
	return ctx
}
