package lsp

import (
	"context"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Sora233/DDBOT/lsp/concern"
	"github.com/Sora233/DDBOT/lsp/mmsg"
	"github.com/Sora233/DDBOT/utils"
	"github.com/Sora233/DDBOT/utils/msgstringer"
	"github.com/sirupsen/logrus"
	"runtime/debug"
	"time"
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
			target := inotify.GetTarget()
			nLogger := inotify.Logger()

			if l.LspStateManager.IsMuted(inotify.GetTarget(), utils.GetBot().GetUin()) {
				nLogger.Info("BOT被禁言，跳过本次推送")
				continue
			}

			c, err := concern.GetConcernBySiteAndType(inotify.Site(), inotify.Type())
			if err != nil {
				nLogger.Errorf("GetConcernBySiteAndType error %v", err)
				continue
			}
			cfg := c.GetStateManager().GetConcernConfig(inotify.GetTarget(), inotify.GetUid())
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
					l.PermissionStateManager.CheckGroupAdministrator(inotify.GetTarget(), utils.GetBot().GetUin())
				var checkAtAll = qqadmin &&
					cfg.GetConcernAt(inotify.GetTarget().GetTargetType()).CheckAtAll(inotify.Type())
				var atAllMark = checkAtAll &&
					c.GetStateManager().CheckAndSetAtAllMark(inotify.GetTarget(), inotify.GetUid())
				nLogger.WithFields(logrus.Fields{
					"qqAdmin":    qqadmin,
					"checkAtAll": checkAtAll,
					"atMark":     atAllMark,
				}).Trace("at_all condition")
				if atBeforeHook.Pass && qqadmin && checkAtAll && atAllMark {
					nLogger = nLogger.WithField("at_all", true)
					newAtAllMsg(m)
				} else {
					ids := cfg.GetConcernAt(inotify.GetTarget().GetTargetType()).GetAtSomeoneList(inotify.Type())
					nLogger = nLogger.WithField("at_QQ", ids)
					newAtIdsMsg(m, ids)
				}
			}

			ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
			if err := l.msgLimit.Acquire(ctx, 1); err != nil {
				cancel()
				nLogger.WithField("Content", msgstringer.MsgToString(m.Elements())).
					Errorf("BOT负载过高，推送已积压超过一分钟，将舍弃本次推送。")
				continue
			}
			cancel()
			l.notifyWg.Add(1)
			nLogger.Info("notify")
			go func() {
				defer l.notifyWg.Done()
				defer func() {
					l.msgLimit.Release(1)
					if e := recover(); e != nil {
						nLogger.WithField("stack", string(debug.Stack())).
							Errorf("notify panic recovered: %v", e)
					}
				}()
				msgs := l.SendMsg(m, target)
				if len(msgs) > 0 {
					cfg.NotifyAfterCallback(inotify, msgs[0])
				} else {
					cfg.NotifyAfterCallback(inotify, nil)
				}
				if atBeforeHook.Pass {
					for _, msg := range msgs {
						var fail bool
						switch x := msg.(type) {
						case *message.GroupMessage:
							if x.Id == -1 {
								// 检查有没有@全体成员
								e := utils.MessageFilter(x.Elements, func(element message.IMessageElement) bool {
									return element.Type() == message.At && element.(*message.AtElement).Target == 0
								})
								if len(e) == 0 {
									continue
								}
								fail = true
							}
						case *message.GuildChannelMessage:
							if x.Id == 0 {
								// 检查有没有@全体成员
								e := utils.MessageFilter(x.Elements, func(element message.IMessageElement) bool {
									return element.Type() == message.At && element.(*message.AtElement).Target == 0
								})
								if len(e) == 0 {
									continue
								}
								fail = true
							}
						}
						if !fail {
							continue
						}
						// @全体成员失败了，可能是次数到了，尝试@列表
						ids := cfg.GetConcernAt(inotify.GetTarget().GetTargetType()).GetAtSomeoneList(inotify.Type())
						if len(ids) != 0 {
							nLogger = nLogger.WithField("at_QQ", ids)
							nLogger.Debug("notify atAll failed, try at someone")
							l.SendMsg(newAtIdsMsg(mmsg.NewMSG(), ids), target)
						} else {
							nLogger.Debug("notify atAll failed, at someone not config")
						}
						break
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
	return m.Cut().AtAll()
}

func newAtIdsMsg(m *mmsg.MSG, ids []int64) *mmsg.MSG {
	if len(ids) > 0 {
		m.Cut()
		for _, id := range ids {
			m.At(id)
		}
	}
	return m
}
