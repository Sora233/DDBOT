package lsp

import (
	"github.com/Logiase/MiraiGo-Template/bot"
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
		select {
		case inotify := <-l.concernNotify:
			switch inotify.Type() {
			case concern.BibiliLive:
				notify := (inotify).(*bilibili.ConcernLiveNotify)
				logger.WithField("site", bilibili.Site).
					WithField("GroupCode", notify.GroupCode).
					WithField("Uid", notify.Mid).
					WithField("GroupName", l.findGroupName(notify.GroupCode)).
					WithField("Name", notify.Name).
					WithField("Title", notify.LiveTitle).
					WithField("Status", notify.Status.String()).
					Info("notify")
				if notify.Status == bilibili.LiveStatus_Living {
					sendingMsg := message.NewSendingMessage()
					notifyMsg := l.NotifyMessage(bot, notify)
					for _, msg := range notifyMsg {
						sendingMsg.Append(msg)
					}
					go l.sendGroupMessage(notify.GroupCode, sendingMsg)
				}
			case concern.BilibiliNews:
				notify := (inotify).(*bilibili.ConcernNewsNotify)
				logger.WithField("site", bilibili.Site).
					WithField("GroupCode", notify.GroupCode).
					WithField("Uid", notify.Mid).
					WithField("GroupName", l.findGroupName(notify.GroupCode)).
					WithField("Name", notify.Name).
					WithField("NewsCount", len(notify.Cards)).
					Info("notify")
				sendingMsg := message.NewSendingMessage()
				notifyMsg := l.NotifyMessage(bot, notify)
				for _, msg := range notifyMsg {
					sendingMsg.Append(msg)
				}
				go l.sendGroupMessage(notify.GroupCode, sendingMsg)
			case concern.DouyuLive:
				notify := (inotify).(*douyu.ConcernLiveNotify)
				logger.WithField("site", douyu.Site).
					WithField("GroupCode", notify.GroupCode).
					WithField("GroupName", l.findGroupName(notify.GroupCode)).
					WithField("Name", notify.Nickname).
					WithField("Title", notify.RoomName).
					WithField("Status", notify.ShowStatus.String()).
					Info("notify")
				if notify.Living() {
					sendingMsg := message.NewSendingMessage()
					notifyMsg := l.NotifyMessage(bot, notify)
					for _, msg := range notifyMsg {
						sendingMsg.Append(msg)
					}
					go l.sendGroupMessage(notify.GroupCode, sendingMsg)
				}
			case concern.YoutubeLive, concern.YoutubeVideo:
				notify := (inotify).(*youtube.ConcernNotify)
				logger.WithField("site", youtube.Site).
					WithField("GroupCode", notify.GroupCode).
					WithField("GroupName", l.findGroupName(notify.GroupCode)).
					WithField("ChannelName", notify.ChannelName).
					WithField("ChannelID", notify.ChannelId).
					WithField("VideoId", notify.VideoId).
					WithField("VideoTitle", notify.VideoTitle).
					WithField("VideoStatus", notify.VideoStatus.String()).
					WithField("VideoType", notify.VideoType.String()).
					Info("notify")
				sendingMsg := message.NewSendingMessage()
				notifyMsg := l.NotifyMessage(bot, notify)
				for _, msg := range notifyMsg {
					sendingMsg.Append(msg)
				}
				go l.sendGroupMessage(notify.GroupCode, sendingMsg)
			case concern.HuyaLive:
				notify := (inotify).(*huya.ConcernLiveNotify)
				logger.WithField("site", huya.Site).
					WithField("GroupCode", notify.GroupCode).
					WithField("GroupName", l.findGroupName(notify.GroupCode)).
					WithField("Name", notify.Name).
					WithField("Title", notify.RoomName).
					WithField("Status", notify.Living).
					Info("notify")
				if notify.Living {
					sendingMsg := message.NewSendingMessage()
					notifyMsg := l.NotifyMessage(bot, notify)
					for _, msg := range notifyMsg {
						sendingMsg.Append(msg)
					}
					go l.sendGroupMessage(notify.GroupCode, sendingMsg)
				}
			}
		}
	}
}

func (l *Lsp) NotifyMessage(bot *bot.Bot, inotify concern.Notify) []message.IMessageElement {
	return inotify.ToMessage()
}

func (l *Lsp) findGroupName(groupCode int64) string {
	gi := bot.Instance.FindGroup(groupCode)
	if gi == nil {
		return ""
	}
	return gi.Name
}
