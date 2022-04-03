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
		select {
		case _inotify, ok := <-l.concernNotify:
			if !ok {
				return
			}
			if _inotify == nil {
				continue
			}
			var inotify = _inotify
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

			cfg.NotifyBeforeCallback(inotify)

			// 注意notify可能会缓存MSG
			var m = l.NotifyMessage(inotify).Clone()

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
					newAtAllMsg(m)
				} else {
					ids := cfg.GetGroupConcernAt().GetAtSomeoneList(inotify.Type())
					nLogger = nLogger.WithField("at_QQ", ids)
					newAtIdsMsg(m, ids)
				}
			}

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
				msgs := l.GM(l.SendMsg(m, target))
				if len(msgs) > 0 {
					cfg.NotifyAfterCallback(inotify, msgs[0])
				} else {
					cfg.NotifyAfterCallback(inotify, nil)
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
							ids := cfg.GetGroupConcernAt().GetAtSomeoneList(inotify.Type())
							if len(ids) != 0 {
								nLogger = nLogger.WithField("at_QQ", ids)
								nLogger.Debug("notify atAll failed, try at someone")
								l.SendMsg(newAtIdsMsg(mmsg.NewMSG(), ids), target)
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

func newAtAllMsg(m *mmsg.MSG) *mmsg.MSG {
	m.Cut()
	m.Append(message.AtAll())
	return m
}

func newAtIdsMsg(m *mmsg.MSG, ids []int64) *mmsg.MSG {
	if len(ids) > 0 {
		m.Cut()
		for _, id := range ids {
			m.Append(message.NewAt(id))
		}
	}
	return m
}
