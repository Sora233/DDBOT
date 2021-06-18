package lsp

import (
	"github.com/Logiase/MiraiGo-Template/bot"
	"github.com/Mrs4s/MiraiGo/client"
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

			if !hook.ShouldSendHook(inotify) {
				log.Debug("notify filtered by hook ShouldSendHook")
				continue
			}

			chainMsg = append(chainMsg, l.NotifyMessage(inotify))

			// atConfig
			{
				var atAllBefore = hook.AtAllBeforeHook(inotify)
				var qqadmin = atAllBefore && checkGroupQQAdministrator(inotify.GetGroupCode(), bot.Uin)
				var checkAtAll = qqadmin && cfg.GroupConcernAt.CheckAtAll(inotify.Type())
				var atAllMark = checkAtAll && innertState.CheckAndSetAtAllMark(inotify.GetGroupCode(), inotify.GetUid())
				log.WithField("atAllBefore", atAllBefore).
					WithField("qqadmin", qqadmin).
					WithField("checkAtAll", checkAtAll).
					WithField("atMark", atAllMark).
					Trace("at_all")
				if atAllBefore && qqadmin && checkAtAll && atAllMark {
					log = log.WithField("at_all", true)
					chainMsg = append(chainMsg, newAtAllMsg())
				} else {
					ids := cfg.GroupConcernAt.GetAtSomeoneList(inotify.Type())
					log.WithField("at_ids", ids).Trace("at someone")
					if len(ids) != 0 {
						log = log.WithField("at_ids", ids)
						chainMsg = append(chainMsg, newAtIdsMsg(ids))
					}
				}
			}
			log.Info("notify")
			go l.sendChainGroupMessage(inotify.GetGroupCode(), chainMsg)
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

func checkGroupQQAdministrator(groupCode int64, uin int64) bool {
	g := bot.Instance.FindGroup(groupCode)
	if g == nil {
		return false
	}
	m := g.FindMember(uin)
	if m == nil {
		return false
	}
	return m.Permission == client.Administrator || m.Permission == client.Owner
}
