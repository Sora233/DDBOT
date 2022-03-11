package lsp

import (
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Sora233/DDBOT/lsp/concern"
	"github.com/Sora233/DDBOT/lsp/mmsg"
	"github.com/Sora233/DDBOT/utils"
	"github.com/sirupsen/logrus"
	"runtime/debug"
)

func (l *Lsp) ConcernNotify() {
	defer func() {
		if err := recover(); err != nil {
			logger.WithField("stack", string(debug.Stack())).Errorf("concern notify recoverd %v", err)
			go l.ConcernNotify()
		}
	}()
	l.wg.Add(1)
	defer l.wg.Done()
	for {
		var chainMsg []*mmsg.MSG
		select {
		case inotify, ok := <-l.concernNotify:
			if !ok {
				return
			}
			if inotify == nil {
				continue
			}
			target := mmsg.NewGroupTarget(inotify.GetGroupCode())
			nLogger := inotify.Logger()

			if l.LspStateManager.IsMuted(inotify.GetGroupCode(), utils.GetBot().GetUin()) {
				nLogger.Info("BOT群内被禁言，跳过本次推送")
				continue
			}

			c, err := concern.GetConcernBySiteAndType(inotify.Site(), inotify.Type())
			if err != nil {
				nLogger.Errorf("GetConcernBySiteAndType error %v", err)
				continue
			}
			cfg := c.GetStateManager().GetGroupConcernConfig(inotify.GetGroupCode(), inotify.GetUid())

			sendHookResult := cfg.ShouldSendHook(inotify)
			if !sendHookResult.Pass {
				nLogger.WithField("Reason", sendHookResult.Reason).Info("notify filtered by hook ShouldSendHook")
				continue
			}

			newsFilterHook := cfg.FilterHook(inotify)
			if !newsFilterHook.Pass {
				nLogger.WithField("Reason", newsFilterHook.Reason).Info("notify filtered by hook FilterHook")
				continue
			}

			// atConfig
			var atBeforeHook = cfg.AtBeforeHook(inotify)
			if !atBeforeHook.Pass {
				nLogger.WithField("Reason", atBeforeHook.Reason).Debug("notify @at filtered by hook AtBeforeHook")
			} else {
				// 有@全体成员 或者 @Someone
				var qqadmin = atBeforeHook.Pass &&
					l.PermissionStateManager.CheckGroupAdministrator(inotify.GetGroupCode(), utils.GetBot().GetUin())
				var checkAtAll = qqadmin &&
					cfg.GetGroupConcernAt().CheckAtAll(inotify.Type())
				var atAllMark = checkAtAll &&
					c.GetStateManager().CheckAndSetAtAllMark(inotify.GetGroupCode(), inotify.GetUid())
				nLogger.WithFields(logrus.Fields{
					"qqAdmin":    qqadmin,
					"checkAtAll": checkAtAll,
					"atMark":     atAllMark,
				}).Trace("at_all condition")
				if atBeforeHook.Pass && qqadmin && checkAtAll && atAllMark {
					nLogger = nLogger.WithField("at_all", true)
					chainMsg = append(chainMsg, newAtAllMsg())
				} else {
					ids := cfg.GetGroupConcernAt().GetAtSomeoneList(inotify.Type())
					nLogger = nLogger.WithField("at_QQ", ids)
					if len(ids) != 0 {
						chainMsg = append(chainMsg, newAtIdsMsg(ids))
					}
				}
			}

			cfg.NotifyBeforeCallback(inotify)

			chainMsg = append([]*mmsg.MSG{l.NotifyMessage(inotify)}, chainMsg...)

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
				msgs := l.SendChainMsg(chainMsg, target)
				if len(msgs) > 0 {
					cfg.NotifyAfterCallback(inotify, msgs[0].(*message.GroupMessage))
				} else {
					cfg.NotifyAfterCallback(inotify, nil)
				}
				if atBeforeHook.Pass {
					for _, msg := range msgs {
						msg := msg.(*message.GroupMessage)
						if msg.Id == -1 {
							// 检查有没有@全体成员
							e := utils.MessageFilter(msg.Elements, func(element message.IMessageElement) bool {
								return element.Type() == message.At && element.(*message.AtElement).Target == 0
							})
							if len(e) == 0 {
								continue
							}
							// @全体成员失败了，可能是次数到了，尝试@列表
							ids := cfg.GetGroupConcernAt().GetAtSomeoneList(inotify.Type())
							if len(ids) != 0 {
								nLogger = nLogger.WithField("at_QQ", ids)
								nLogger.Debug("notify atAll failed, try at someone")
								l.SendMsg(newAtIdsMsg(ids), target)
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

func (l *Lsp) NotifyMessage(inotify concern.Notify) *mmsg.MSG {
	return inotify.ToMessage()
}

func newAtAllMsg() *mmsg.MSG {
	msg := mmsg.NewMSG()
	msg.Append(message.AtAll())
	return msg
}

func newAtIdsMsg(ids []int64) *mmsg.MSG {
	msg := new(mmsg.MSG)
	for _, id := range ids {
		msg.Append(message.NewAt(id))
	}
	return msg
}
