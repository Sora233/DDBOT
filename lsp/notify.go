package lsp

import (
	"github.com/Logiase/MiraiGo-Template/bot"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Sora233/DDBOT/concern"
	"github.com/Sora233/DDBOT/lsp/bilibili"
	"github.com/Sora233/DDBOT/lsp/douyu"
	"github.com/Sora233/DDBOT/lsp/huya"
	"github.com/Sora233/DDBOT/lsp/youtube"
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
	for {
		var chainMsg []*message.SendingMessage
		select {
		case inotify := <-l.concernNotify:
			log := getLog(inotify)
			innertState := l.getInnerState(inotify.Type())
			cfg := l.getConcernConfig(inotify.GetGroupCode(), inotify.GetUid(), inotify.Type())
			hook := l.getConcernConfigHook(inotify.Type(), cfg)

			sendHookResult := hook.ShouldSendHook(inotify)
			if !sendHookResult.Pass {
				log.WithField("Reason", sendHookResult.Reason).Debug("notify filtered by hook ShouldSendHook")
				continue
			}

			newsFilterHook := hook.NewsFilterHook(inotify)
			if !newsFilterHook.Pass {
				log.WithField("Reason", newsFilterHook.Reason).Debug("notify filtered by hook NewsFilterHook")
				continue
			}

			chainMsg = append(chainMsg, l.NotifyMessage(inotify))

			// atConfig
			var atBeforeHook = hook.AtBeforeHook(inotify)
			if !atBeforeHook.Pass {
				log.WithField("Reason", atBeforeHook.Reason).Debug("notify @at filtered by hook AtBeforeHook")
			} else {
				// 有@全体成员 或者 @Someone
				var qqadmin = atBeforeHook.Pass && l.PermissionStateManager.CheckGroupAdministrator(inotify.GetGroupCode(), bot.Uin)
				var checkAtAll = qqadmin && cfg.GroupConcernAt.CheckAtAll(inotify.Type())
				var atAllMark = checkAtAll && innertState.CheckAndSetAtAllMark(inotify.GetGroupCode(), inotify.GetUid())
				log.WithField("atBeforeHook", atBeforeHook).
					WithField("qqadmin", qqadmin).
					WithField("checkAtAll", checkAtAll).
					WithField("atMark", atAllMark).
					Trace("at_all")
				if atBeforeHook.Pass && qqadmin && checkAtAll && atAllMark {
					log = log.WithField("at_all", true)
					chainMsg = append(chainMsg, newAtAllMsg())
				} else {
					ids := cfg.GroupConcernAt.GetAtSomeoneList(inotify.Type())
					log = log.WithField("at_QQ", ids)
					if len(ids) != 0 {
						chainMsg = append(chainMsg, newAtIdsMsg(ids))
					}
				}
			}

			log.Info("notify")

			go func() {
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
								log = log.WithField("at_QQ", ids)
								log.Errorf("notify atAll failed, try at someone")
								l.sendGroupMessage(inotify.GetGroupCode(), newAtIdsMsg(ids))
							} else {
								log.Errorf("notify atAll failed, at someone not config")
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

func getLog(inotify concern.Notify) *logrus.Entry {
	switch inotify.Type() {
	case concern.BibiliLive:
		notify := (inotify).(*bilibili.ConcernLiveNotify)
		return logger.WithField("site", bilibili.Site).
			WithField("GroupCode", notify.GroupCode).
			WithField("Uid", notify.Mid).
			WithField("GroupName", findGroupName(notify.GroupCode)).
			WithField("Name", notify.Name).
			WithField("Title", notify.LiveTitle).
			WithField("Status", notify.Status.String())
	case concern.BilibiliNews:
		notify := (inotify).(*bilibili.ConcernNewsNotify)
		return logger.WithField("site", bilibili.Site).
			WithField("GroupCode", notify.GroupCode).
			WithField("Uid", notify.Mid).
			WithField("GroupName", findGroupName(notify.GroupCode)).
			WithField("Name", notify.Name).
			WithField("NewsCount", len(notify.Cards))
	case concern.DouyuLive:
		notify := (inotify).(*douyu.ConcernLiveNotify)
		return logger.WithField("site", douyu.Site).
			WithField("GroupCode", notify.GroupCode).
			WithField("GroupName", findGroupName(notify.GroupCode)).
			WithField("Name", notify.Nickname).
			WithField("Title", notify.RoomName).
			WithField("Status", notify.ShowStatus.String())
	case concern.YoutubeLive, concern.YoutubeVideo:
		notify := (inotify).(*youtube.ConcernNotify)
		return logger.WithField("site", youtube.Site).
			WithField("GroupCode", notify.GroupCode).
			WithField("GroupName", findGroupName(notify.GroupCode)).
			WithField("ChannelName", notify.ChannelName).
			WithField("ChannelID", notify.ChannelId).
			WithField("VideoId", notify.VideoId).
			WithField("VideoTitle", notify.VideoTitle).
			WithField("VideoStatus", notify.VideoStatus.String()).
			WithField("VideoType", notify.VideoType.String())
	case concern.HuyaLive:
		notify := (inotify).(*huya.ConcernLiveNotify)
		return logger.WithField("site", huya.Site).
			WithField("GroupCode", notify.GroupCode).
			WithField("GroupName", findGroupName(notify.GroupCode)).
			WithField("Name", notify.Name).
			WithField("Title", notify.RoomName).
			WithField("Status", notify.Living)
	}
	return logger.WithFields(utils.GroupLogFields(inotify.GetGroupCode()))
}

func findGroupName(groupCode int64) string {
	gi := bot.Instance.FindGroup(groupCode)
	if gi == nil {
		return ""
	}
	return gi.Name
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
