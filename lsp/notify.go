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
			debugNotify(inotify)
			if !inotify.ShouldSend() {
				continue
			}
			chainMsg = append(chainMsg, l.NotifyMessage(inotify))
			innertState := l.getInnerState(inotify.Type())
			cfg := l.getConcernConfig(inotify.GetGroupCode(), inotify.GetUid(), inotify.Type())
			if cfg != nil {
				// atConfig
				{
					var qqadmin = checkGroupQQAdministrator(inotify.GetGroupCode(), bot.Uin)
					var checkAtAll = qqadmin && cfg.GroupConcernAt.CheckAtAll(inotify.Type())
					var atAllMark = qqadmin && checkAtAll && innertState.CheckAndSetAtAllMark(inotify.GetGroupCode(), inotify.GetUid())
					logger.WithField("qqadmin", qqadmin).WithField("checkAtAll", checkAtAll).WithField("atMark", atAllMark).Trace("at_all")
					if qqadmin && checkAtAll && atAllMark {
						chainMsg = append(chainMsg, newAtAllMsg())
					} else {
						ids := cfg.GroupConcernAt.GetAtSomeoneList(inotify.Type())
						logger.WithField("ids", ids).Trace("at someone")
						if len(ids) != 0 {
							chainMsg = append(chainMsg, newAtIdsMsg(ids))
						}
					}
				}
			}
			go l.sendChainGroupMessage(inotify.GetGroupCode(), chainMsg)
		}
	}
}

func (l *Lsp) NotifyMessage(inotify concern.Notify) *message.SendingMessage {
	var sendingMsgs = new(message.SendingMessage)
	sendingMsgs.Elements = inotify.ToMessage()
	return sendingMsgs
}

func debugNotify(inotify concern.Notify) {
	switch inotify.Type() {
	case concern.BibiliLive:
		notify := (inotify).(*bilibili.ConcernLiveNotify)
		logger.WithField("site", bilibili.Site).
			WithField("GroupCode", notify.GroupCode).
			WithField("Uid", notify.Mid).
			WithField("GroupName", findGroupName(notify.GroupCode)).
			WithField("Name", notify.Name).
			WithField("Title", notify.LiveTitle).
			WithField("Status", notify.Status.String()).
			Info("notify")
	case concern.BilibiliNews:
		notify := (inotify).(*bilibili.ConcernNewsNotify)
		logger.WithField("site", bilibili.Site).
			WithField("GroupCode", notify.GroupCode).
			WithField("Uid", notify.Mid).
			WithField("GroupName", findGroupName(notify.GroupCode)).
			WithField("Name", notify.Name).
			WithField("NewsCount", len(notify.Cards)).
			Info("notify")
	case concern.DouyuLive:
		notify := (inotify).(*douyu.ConcernLiveNotify)
		logger.WithField("site", douyu.Site).
			WithField("GroupCode", notify.GroupCode).
			WithField("GroupName", findGroupName(notify.GroupCode)).
			WithField("Name", notify.Nickname).
			WithField("Title", notify.RoomName).
			WithField("Status", notify.ShowStatus.String()).
			Info("notify")
	case concern.YoutubeLive, concern.YoutubeVideo:
		notify := (inotify).(*youtube.ConcernNotify)
		logger.WithField("site", youtube.Site).
			WithField("GroupCode", notify.GroupCode).
			WithField("GroupName", findGroupName(notify.GroupCode)).
			WithField("ChannelName", notify.ChannelName).
			WithField("ChannelID", notify.ChannelId).
			WithField("VideoId", notify.VideoId).
			WithField("VideoTitle", notify.VideoTitle).
			WithField("VideoStatus", notify.VideoStatus.String()).
			WithField("VideoType", notify.VideoType.String()).
			Info("notify")
	case concern.HuyaLive:
		notify := (inotify).(*huya.ConcernLiveNotify)
		logger.WithField("site", huya.Site).
			WithField("GroupCode", notify.GroupCode).
			WithField("GroupName", findGroupName(notify.GroupCode)).
			WithField("Name", notify.Name).
			WithField("Title", notify.RoomName).
			WithField("Status", notify.Living).
			Info("notify")
	}
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
