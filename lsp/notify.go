package lsp

import (
	"context"
	"fmt"
	"math"
	"runtime/debug"
	"time"

	"github.com/LagrangeDev/LagrangeGo/message"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"

	"github.com/Sora233/DDBOT/v2/lsp/concern"
	"github.com/Sora233/DDBOT/v2/lsp/mmsg"
	"github.com/Sora233/DDBOT/v2/utils"
	"github.com/Sora233/DDBOT/v2/utils/msgstringer"
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
			target := mmsg.NewGroupTarget(uint32(inotify.GetGroupCode()))
			nLogger := inotify.Logger()

			if l.LspStateManager.IsMuted(uint32(inotify.GetGroupCode()), utils.GetBot().GetUin()) {
				nLogger.Info("BOT群内被禁言，跳过本次推送")
				continue
			}

			c, err := concern.GetConcernBySiteAndType(inotify.Site(), inotify.Type())
			if err != nil {
				nLogger.Errorf("GetConcernBySiteAndType error %v", err)
				continue
			}
			cfg := c.GetStateManager().GetGroupConcernConfig(uint32(inotify.GetGroupCode()), inotify.GetUid())
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
					l.PermissionStateManager.CheckGroupAdministrator(uint32(inotify.GetGroupCode()), utils.GetBot().GetUin())
				var checkAtAll = qqadmin &&
					cfg.GetGroupConcernAt().CheckAtAll(inotify.Type())
				var atAllMark = checkAtAll &&
					c.GetStateManager().CheckAndSetAtAllMark(uint32(inotify.GetGroupCode()), inotify.GetUid())
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
				msgs := l.GM(l.SendMsg(m, target))
				if len(msgs) > 0 {
					cfg.NotifyAfterCallback(inotify, msgs[0])
				} else {
					cfg.NotifyAfterCallback(inotify, nil)
				}
				if atBeforeHook.Pass {
					var atIdsOnce bool
					for _, msg := range msgs {
						if msg.ID == 0 {
							// 检查有没有@全体成员
							//e := utils.MessageFilter(msg.Elements, func(element message.IMessageElement) bool {
							//	return element.Type() == message.At && element.(*message.AtElement).Target == 0
							//})
							e := lo.Filter(msg.Elements, func(element message.IMessageElement, _ int) bool {
								return element.Type() == message.At && element.(*message.AtElement).TargetUin == 0
							})
							if len(e) == 0 {
								continue
							}
							// 2022/09/24 现在@全员不会再作为单独一条消息
							// 有@全体成员的消息应该去掉之后重试
							secondM := mmsg.NewMSGFromGroupMessage(msg)
							secondM.Drop(func(e message.IMessageElement, _ int) bool {
								return e.Type() == message.At && e.(*message.AtElement).TargetUin == 0
							})

							secondRes := l.GM(l.SendMsg(secondM, target))
							// secondRes一定是一条
							if len(secondRes) != 1 {
								panic(fmt.Sprintf("INTERNAL: len(secondRes) is %v", len(secondRes)))
							}
							if secondRes[0].ID == math.MaxUint32 {
								// 去掉@全员还是发送失败
								continue
							}
							if !atIdsOnce {
								// 去掉@全员之后发送成功，可能是次数到了，尝试@列表
								atIdsOnce = true
							}
						}
					}
					if atIdsOnce {
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
			}()
		}
	}
}

func (l *Lsp) NotifyMessage(inotify concern.Notify) *mmsg.MSG {
	return inotify.ToMessage()
}

func newAtAllMsg(m *mmsg.MSG) *mmsg.MSG {
	return m.AtAll(true)
}

func newAtIdsMsg(m *mmsg.MSG, ids []uint32) *mmsg.MSG {
	if len(ids) > 0 {
		m.Cut()
		for _, id := range ids {
			m.At(id)
		}
	}
	return m
}
