package lsp

import (
	"github.com/Logiase/MiraiGo-Template/bot"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Sora233/DDBOT/concern"
	"github.com/Sora233/DDBOT/utils"
	"runtime/debug"
)

func (l *Lsp) ConcernNotify(bot *bot.Bot) {
	defer func() {
		if err := recover(); err != nil {
			logger.WithField("stack", string(debug.Stack())).Errorf("concern notify recoverd %v", err)
			go l.ConcernNotify(bot)
		}
	}()
	l.wg.Add(1)
	defer l.wg.Done()
	for {
		var chainMsg []*message.SendingMessage
		select {
		case inotify, ok := <-l.concernNotify:
			if !ok {
				return
			}
			if inotify == nil {
				continue
			}
			nLogger := inotify.Logger()

			if l.LspStateManager.IsMuted(inotify.GetGroupCode(), bot.Uin) {
				nLogger.Info("bot is muted in group, skip notify")
				continue
			}

			innertState := l.getInnerState(inotify.Type())
			cfg := l.getConcernConfig(inotify.GetGroupCode(), inotify.GetUid(), inotify.Type())
			hook := l.getConcernConfigHook(inotify.Type(), cfg)

			sendHookResult := hook.ShouldSendHook(inotify)
			if !sendHookResult.Pass {
				nLogger.WithField("Reason", sendHookResult.Reason).Debug("notify filtered by hook ShouldSendHook")
				continue
			}

			newsFilterHook := hook.NewsFilterHook(inotify)
			if !newsFilterHook.Pass {
				nLogger.WithField("Reason", newsFilterHook.Reason).Debug("notify filtered by hook NewsFilterHook")
				continue
			}

			chainMsg = append(chainMsg, l.NotifyMessage(inotify))

			// atConfig
			var atBeforeHook = hook.AtBeforeHook(inotify)
			if !atBeforeHook.Pass {
				nLogger.WithField("Reason", atBeforeHook.Reason).Debug("notify @at filtered by hook AtBeforeHook")
			} else {
				// 有@全体成员 或者 @Someone
				var qqadmin = atBeforeHook.Pass && l.PermissionStateManager.CheckGroupAdministrator(inotify.GetGroupCode(), bot.Uin)
				var checkAtAll = qqadmin && cfg.GroupConcernAt.CheckAtAll(inotify.Type())
				var atAllMark = checkAtAll && innertState.CheckAndSetAtAllMark(inotify.GetGroupCode(), inotify.GetUid())
				nLogger.WithField("atBeforeHook", atBeforeHook).
					WithField("qqadmin", qqadmin).
					WithField("checkAtAll", checkAtAll).
					WithField("atMark", atAllMark).
					Trace("at_all")
				if atBeforeHook.Pass && qqadmin && checkAtAll && atAllMark {
					nLogger = nLogger.WithField("at_all", true)
					chainMsg = append(chainMsg, newAtAllMsg())
				} else {
					ids := cfg.GroupConcernAt.GetAtSomeoneList(inotify.Type())
					nLogger = nLogger.WithField("at_QQ", ids)
					if len(ids) != 0 {
						chainMsg = append(chainMsg, newAtIdsMsg(ids))
					}
				}
			}

			nLogger.Info("notify")

			l.notifyWg.Add(1)
			go func() {
				defer l.notifyWg.Done()
				msgs := l.sendChainGroupMessage(inotify.GetGroupCode(), chainMsg)
				if atBeforeHook.Pass {
					for _, msg := range msgs {
						if msg.Id == -1 {
							// 检查有没有@全体成员
							e := utils.MessageFilter(msg.Elements, func(element message.IMessageElement) bool {
								return element.Type() == message.At && element.(*message.AtElement).Target == 0
							})
							if len(e) == 0 {
								continue
							}
							// @全体成员失败了，可能是次数到了，尝试@列表
							ids := cfg.GroupConcernAt.GetAtSomeoneList(inotify.Type())
							if len(ids) != 0 {
								nLogger = nLogger.WithField("at_QQ", ids)
								nLogger.Errorf("notify atAll failed, try at someone")
								l.sendGroupMessage(inotify.GetGroupCode(), newAtIdsMsg(ids))
							} else {
								nLogger.Errorf("notify atAll failed, at someone not config")
							}
						}
					}
				}
			}()
		}
	}
}

func (l *Lsp) NotifyMessage(inotify concern.Notify) *message.SendingMessage {
	var sendingMsgs = new(message.SendingMessage)
	sendingMsgs.Elements = inotify.ToMessage()
	return sendingMsgs
}

func newAtAllMsg() *message.SendingMessage {
	msg := new(message.SendingMessage)
	msg.Append(message.AtAll())
	return msg
}

func newAtIdsMsg(ids []int64) *message.SendingMessage {
	msg := new(message.SendingMessage)
	for _, id := range ids {
		msg.Append(message.NewAt(id))
	}
	return msg
}
