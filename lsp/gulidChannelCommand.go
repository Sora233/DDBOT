package lsp

import (
	"fmt"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Sora233/DDBOT/lsp/mmsg"
	"github.com/Sora233/DDBOT/lsp/mmsg/mt"
	"github.com/Sora233/DDBOT/lsp/template"
	"github.com/Sora233/DDBOT/utils"
	"github.com/alecthomas/kong"
	"github.com/sirupsen/logrus"
	"runtime/debug"
)

type LspGulidChannelCommand struct {
	msg *message.GuildChannelMessage

	*Runtime
}

func NewLspGulidChannelCommand(l *Lsp, msg *message.GuildChannelMessage) *LspGulidChannelCommand {
	c := &LspGulidChannelCommand{
		msg:     msg,
		Runtime: NewRuntime(l),
	}
	c.Parse(c.msg.Elements)
	return c
}

func (gc *LspGulidChannelCommand) Execute() {
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
	case HelpCommand:
		gc.HelpCommand()
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

func (gc *LspGulidChannelCommand) HelpCommand() {
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
	gc.sendChain(gc.templateMsg("command.gulid.help.tmpl", nil))
}

func (gc *LspGulidChannelCommand) DefaultLogger() *logrus.Entry {
	return logger.WithFields(logrus.Fields{
		"Uin":  gc.uin(),
		"Name": gc.name(),
	}).WithFields(utils.GulidChannelLogFields(gc.gulidChannelID()))
}

func (gc *LspGulidChannelCommand) DefaultLoggerWithCommand(command string) *logrus.Entry {
	return gc.DefaultLogger().WithField("Command", command)
}

func (gc *LspGulidChannelCommand) commonTemplateData() map[string]interface{} {
	return map[string]interface{}{
		"msg":          gc.msg,
		"member_code":  gc.uin(),
		"member_name":  gc.name(),
		"channel_id":   gc.channelID(),
		"gulid_id":     gc.gulidID(),
		"channel_name": utils.GetBot().FindChannelName(gc.gulidChannelID()),
		"gulid_name":   utils.GetBot().FindGulidName(gc.gulidID()),
		"command":      CommandMaps,
	}
}

func (gc *LspGulidChannelCommand) templateMsg(name string, data map[string]interface{}) *mmsg.MSG {
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

func (gc *LspGulidChannelCommand) gulidID() uint64 {
	return gc.msg.GuildId
}

func (gc *LspGulidChannelCommand) channelID() uint64 {
	return gc.msg.ChannelId
}

func (gc *LspGulidChannelCommand) gulidChannelID() (uint64, uint64) {
	return gc.gulidID(), gc.channelID()
}

func (gc *LspGulidChannelCommand) sender() *message.GuildSender {
	return gc.msg.Sender
}
func (gc *LspGulidChannelCommand) uin() uint64 {
	return gc.sender().TinyId
}

func (gc *LspGulidChannelCommand) name() string {
	return gc.sender().Nickname
}

func (gc *LspGulidChannelCommand) textSend(text string) *message.GuildChannelMessage {
	return gc.send(mmsg.NewText(text))
}

func (gc *LspGulidChannelCommand) textReply(text string) *message.GuildChannelMessage {
	return gc.send(mmsg.NewText(text))
}

func (gc *LspGulidChannelCommand) textReplyF(format string, args ...interface{}) *message.GuildChannelMessage {
	return gc.send(mmsg.NewTextf(format, args...))
}

func (gc *LspGulidChannelCommand) send(msg *mmsg.MSG) *message.GuildChannelMessage {
	return gc.l.GCM(gc.l.SendMsg(msg, mt.NewGulidTarget(gc.gulidChannelID())))[0]
}

func (gc *LspGulidChannelCommand) sendChain(msg *mmsg.MSG) []*message.GuildChannelMessage {
	return gc.l.GCM(gc.l.SendMsg(msg, mt.NewGulidTarget(gc.gulidChannelID())))
}
