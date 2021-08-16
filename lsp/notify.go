package lsp

import (
	"github.com/Logiase/MiraiGo-Template/bot"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Sora233/DDBOT/concern"
	"github.com/Sora233/DDBOT/utils"
	"github.com/sirupsen/logrus"
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
				nLogger.Info("BOT群内被禁言，跳过本次推送")
				continue
			}

			innertState := l.getInnerState(inotify.Type())
			cfg := l.getConcernConfig(inotify.GetGroupCode(), inotify.GetUid(), inotify.Type())
			notifyManager := l.getConcernConfigNotifyManager(inotify.Type(), cfg)

			sendHookResult := notifyManager.ShouldSendHook(inotify)
			if !sendHookResult.Pass {
				nLogger.WithField("Reason", sendHookResult.Reason).Debug("notify filtered by hook ShouldSendHook")
				continue
			}

			newsFilterHook := notifyManager.NewsFilterHook(inotify)
			if !newsFilterHook.Pass {
				nLogger.WithField("Reason", newsFilterHook.Reason).Debug("notify filtered by hook NewsFilterHook")
				continue
			}

			// atConfig
			var atBeforeHook = notifyManager.AtBeforeHook(inotify)
			if !atBeforeHook.Pass {
				nLogger.WithField("Reason", atBeforeHook.Reason).Debug("notify @at filtered by hook AtBeforeHook")
			} else {
				// 有@全体成员 或者 @Someone
				var qqadmin = atBeforeHook.Pass && l.PermissionStateManager.CheckGroupAdministrator(inotify.GetGroupCode(), bot.Uin)
				var checkAtAll = qqadmin && cfg.GroupConcernAt.CheckAtAll(inotify.Type())
				var atAllMark = checkAtAll && innertState.CheckAndSetAtAllMark(inotify.GetGroupCode(), inotify.GetUid())
				nLogger.WithFields(logrus.Fields{
					"atBeforeHook": atBeforeHook,
					"qqAdmin":      qqadmin,
					"checkAtAll":   checkAtAll,
					"atMark":       atAllMark,
				}).Trace("at_all")
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

			notifyManager.NotifyBeforeCallback(inotify)

			chainMsg = append([]*message.SendingMessage{l.NotifyMessage(inotify)}, chainMsg...)

			nLogger.Info("notify")

			l.notifyWg.Add(1)
			go func() {
				defer l.notifyWg.Done()
				defer func() {
					if e := recover(); e != nil {
						nLogger.WithField("stack", string(debug.Stack())).
							Errorf("notify panic recovered: %v", e)
					}
				}()
				msgs := l.sendChainGroupMessage(inotify.GetGroupCode(), chainMsg)
				if len(msgs) > 0 {
					notifyManager.NotifyAfterCallback(inotify, msgs[0])
				} else {
					notifyManager.NotifyAfterCallback(inotify, nil)
				}
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
								nLogger.Debug("notify atAll failed, try at someone")
								l.sendGroupMessage(inotify.GetGroupCode(), newAtIdsMsg(ids))
							} else {
								nLogger.Debug("notify atAll failed, at someone not config")
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
