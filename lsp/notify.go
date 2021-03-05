package lsp

import (
	"encoding/json"
	"github.com/Logiase/MiraiGo-Template/bot"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Sora233/Sora233-MiraiGo/concern"
	"github.com/Sora233/Sora233-MiraiGo/lsp/bilibili"
	"github.com/Sora233/Sora233-MiraiGo/lsp/douyu"
	"github.com/Sora233/Sora233-MiraiGo/lsp/youtube"
	"github.com/Sora233/Sora233-MiraiGo/proxy_pool"
	localutils "github.com/Sora233/Sora233-MiraiGo/utils"
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
					WithField("GroupName", bot.FindGroup(notify.GroupCode).Name).
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
					l.sendGroupMessage(notify.GroupCode, sendingMsg)
				}
			case concern.BilibiliNews:
				notify := (inotify).(*bilibili.ConcernNewsNotify)
				logger.WithField("site", bilibili.Site).
					WithField("GroupCode", notify.GroupCode).
					WithField("Uid", notify.Mid).
					WithField("GroupName", bot.FindGroup(notify.GroupCode).Name).
					WithField("Name", notify.Name).
					WithField("NewsCount", len(notify.Cards)).
					Info("notify")
				sendingMsg := message.NewSendingMessage()
				notifyMsg := l.NotifyMessage(bot, notify)
				for _, msg := range notifyMsg {
					sendingMsg.Append(msg)
				}
				l.sendGroupMessage(notify.GroupCode, sendingMsg)
			case concern.DouyuLive:
				notify := (inotify).(*douyu.ConcernLiveNotify)
				logger.WithField("site", douyu.Site).
					WithField("GroupCode", notify.GroupCode).
					WithField("GroupName", bot.FindGroup(notify.GroupCode).Name).
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
					l.sendGroupMessage(notify.GroupCode, sendingMsg)
				}
			case concern.YoutubeLive, concern.YoutubeVideo:
				notify := (inotify).(*youtube.ConcernNotify)
				logger.WithField("site", youtube.Site).
					WithField("GroupCode", notify.GroupCode).
					WithField("GroupName", bot.FindGroup(notify.GroupCode).Name).
					WithField("ChannelName", notify.ChannelName).
					WithField("ChannelID", notify.ChannelId).
					WithField("VideoId", notify.VideoId).
					WithField("VideoTitle", notify.VideoTitle).
					WithField("VideoStatus", notify.VideoStatus.String()).
					WithField("VideoType", notify.VideoType.String()).
					Info("notify")
				sendingMsg := message.NewSendingMessage()
				notifyMsg := l.notifyYoutube(bot, notify)
				for _, msg := range notifyMsg {
					sendingMsg.Append(msg)
				}
				l.sendGroupMessage(notify.GroupCode, sendingMsg)
			}
		}
	}
}

func (l *Lsp) NotifyMessage(bot *bot.Bot, inotify concern.Notify) []message.IMessageElement {
	var result []message.IMessageElement
	switch inotify.Type() {
	case concern.BibiliLive:
		notify := (inotify).(*bilibili.ConcernLiveNotify)
		result = append(result, l.notifyBilibiliLive(bot, notify)...)
	case concern.BilibiliNews:
		notify := (inotify).(*bilibili.ConcernNewsNotify)
		result = append(result, l.notifyBilibiliNews(bot, notify)...)
	case concern.DouyuLive:
		notify := (inotify).(*douyu.ConcernLiveNotify)
		result = append(result, l.notifyDouyuLive(bot, notify)...)
	case concern.YoutubeLive, concern.YoutubeVideo:
		notify := (inotify).(*youtube.ConcernNotify)
		result = append(result, l.notifyYoutube(bot, notify)...)
	}
	return result
}

func (l *Lsp) notifyBilibiliLive(bot *bot.Bot, notify *bilibili.ConcernLiveNotify) []message.IMessageElement {
	var result []message.IMessageElement
	switch notify.Status {
	case bilibili.LiveStatus_Living:
		result = append(result, localutils.MessageTextf("%s正在直播【%v】\n", notify.Name, notify.LiveTitle))
		result = append(result, message.NewText(notify.RoomUrl))
		cover, err := localutils.UploadGroupImageByUrl(notify.GroupCode, notify.Cover, false, proxy_pool.PreferAny)
		if err != nil {
			logger.WithField("group_code", notify.GroupCode).
				WithField("cover", notify.Cover).
				Errorf("add cover failed %v", err)
		} else {
			result = append(result, cover)
		}
	case bilibili.LiveStatus_NoLiving:
		result = append(result, localutils.MessageTextf("%s暂未直播\n", notify.Name))
		result = append(result, message.NewText(notify.RoomUrl))
	}
	return result
}

func (l *Lsp) notifyBilibiliNews(bot *bot.Bot, notify *bilibili.ConcernNewsNotify) []message.IMessageElement {
	var results []message.IMessageElement
	for index, card := range notify.Cards {
		var result []message.IMessageElement
		log := logger.WithField("DescType", card.GetDesc().GetType().String())
		dynamicUrl := bilibili.DynamicUrl(card.GetDesc().GetDynamicIdStr())
		date := localutils.TimestampFormat(int64(card.GetDesc().GetTimestamp()))
		switch card.GetDesc().GetType() {
		case bilibili.DynamicDescType_WithOrigin:
			cardOrigin, err := notify.GetCardWithOrig(index)
			if err != nil {
				log.WithField("name", notify.Name).WithField("card", card).Errorf("cast failed %v", err)
				continue
			}
			originName := cardOrigin.GetOriginUser().GetInfo().GetUname()
			// very sb
			switch cardOrigin.GetItem().GetOrigType() {
			case bilibili.DynamicDescType_WithImage:
				result = append(result, localutils.MessageTextf("%v转发了%v的动态：\n%v\n%v\n\n原动态：\n", notify.Name, originName, date, cardOrigin.GetItem().GetContent()))
				origin := new(bilibili.CardWithImage)
				err := json.Unmarshal([]byte(cardOrigin.GetOrigin()), origin)
				if err != nil {
					log.Errorf("Unmarshal origin cardWithImage failed %v", err)
					continue
				}
				result = append(result, localutils.MessageTextf("%v\n", origin.GetItem().GetDescription()))
				for _, pic := range origin.GetItem().GetPictures() {
					groupImage, err := localutils.UploadGroupImageByUrl(notify.GroupCode, pic.GetImgSrc(), false, proxy_pool.PreferAny)
					if err != nil {
						log.WithField("pic", pic).Errorf("upload group image %v", err)
						continue
					}
					result = append(result, groupImage)
				}
			case bilibili.DynamicDescType_TextOnly:
				result = append(result, localutils.MessageTextf("%v转发了%v的动态：\n%v\n%v\n\n原动态：\n", notify.Name, originName, date, cardOrigin.GetItem().GetContent()))
				origin := new(bilibili.CardTextOnly)
				err := json.Unmarshal([]byte(cardOrigin.GetOrigin()), origin)
				if err != nil {
					log.Errorf("Unmarshal origin cardWithText failed %v", err)
					continue
				}
				result = append(result, localutils.MessageTextf("%v\n", origin.GetItem().GetContent()))
			case bilibili.DynamicDescType_WithVideo:
				result = append(result, localutils.MessageTextf("%v转发了%v的投稿：\n%v\n%v\n\n原视频：\n", notify.Name, originName, date, cardOrigin.GetItem().GetContent()))
				origin := new(bilibili.CardWithVideo)
				err := json.Unmarshal([]byte(cardOrigin.GetOrigin()), origin)
				if err != nil {
					log.Errorf("Unmarshal origin cardWithVideo failed %v", err)
					continue
				}
				result = append(result, localutils.MessageTextf("%v\n%v\n", origin.GetTitle(), origin.GetDesc()))
				cover, err := localutils.UploadGroupImageByUrl(notify.GroupCode, origin.GetPic(), true, proxy_pool.PreferAny)
				if err != nil {
					log.Errorf("upload video cover failed %v", err)
					continue
				}
				result = append(result, cover)
			case bilibili.DynamicDescType_WithPost:
				result = append(result, localutils.MessageTextf("%v转发了%v的专栏：\n%v\n%v\n\n原专栏：\n", notify.Name, originName, date, cardOrigin.GetItem().GetContent()))
				origin := new(bilibili.CardWithPost)
				err := json.Unmarshal([]byte(cardOrigin.GetOrigin()), origin)
				if err != nil {
					log.Errorf("Unmarshal origin cardWithPost failed %v", err)
					continue
				}
				result = append(result, localutils.MessageTextf("%v\n%v\n", origin.GetTitle(), origin.GetSummary()))
				var cover *message.GroupImageElement
				if len(origin.GetImageUrls()) >= 1 {
					cover, err = localutils.UploadGroupImageByUrl(notify.GroupCode, origin.GetImageUrls()[0], false, proxy_pool.PreferAny)
				} else {
					cover, err = localutils.UploadGroupImageByUrl(notify.GroupCode, origin.GetBannerUrl(), false, proxy_pool.PreferAny)
				}
				if err != nil {
					log.Errorf("upload post cover failed %v", err)
					continue
				}
				result = append(result, cover)
			case bilibili.DynamicDescType_WithMusic:
				// TODO
				result = append(result, localutils.MessageTextf("%v转发了%v的动态音乐：\n%v\n%v\n", notify.Name, originName, date, cardOrigin.GetItem().GetContent()))
			case bilibili.DynamicDescType_WithSketch:
				// TODO
				result = append(result, localutils.MessageTextf("%v转发了%v的动态：\n%v\n%v\n", notify.Name, originName, date, cardOrigin.GetItem().GetContent()))
			case bilibili.DynamicDescType_WithLive:
				result = append(result, localutils.MessageTextf("%v分享了%v的直播：\n%v\n%v\n\n原直播间：\n", notify.Name, originName, date, cardOrigin.GetItem().GetContent()))
				origin := new(bilibili.CardWithLive)
				err := json.Unmarshal([]byte(cardOrigin.GetOrigin()), origin)
				if err != nil {
					log.Errorf("Unmarshal origin CardWithLive failed %v", err)
					continue
				}
				result = append(result, localutils.MessageTextf("%v\n", origin.GetTitle()))
				groupImage, err := localutils.UploadGroupImageByUrl(notify.GroupCode, origin.GetCover(), false, proxy_pool.PreferAny)
				if err != nil {
					log.Errorf("upload live cover failed %v", err)
					continue
				}
				result = append(result, groupImage)
			case bilibili.DynamicDescType_WithLiveV2:
				result = append(result, localutils.MessageTextf("%v分享了%v的直播：\n%v\n%v\n\n原直播间：\n", notify.Name, originName, date, cardOrigin.GetItem().GetContent()))
				origin := new(bilibili.CardWithLiveV2)
				err := json.Unmarshal([]byte(cardOrigin.GetOrigin()), origin)
				if err != nil {
					log.Errorf("Unmarshal origin CardWithLiveV2 failed %v", err)
					continue
				}
				result = append(result, localutils.MessageTextf("%v\n", origin.GetLivePlayInfo().GetTitle()))
				groupImage, err := localutils.UploadGroupImageByUrl(notify.GroupCode, origin.GetLivePlayInfo().GetCover(), false, proxy_pool.PreferAny)
				if err != nil {
					log.Errorf("upload liveV2 cover failed %v", err)
					continue
				}
				result = append(result, groupImage)
			default:
				log.WithField("content", card.GetCard()).Info("found new type with origin")
				result = append(result, localutils.MessageTextf("%v转发了%v的动态：\n%v\n%v\n", notify.Name, originName, date, cardOrigin.GetItem().GetContent()))
			}
		case bilibili.DynamicDescType_WithImage:
			cardImage, err := notify.GetCardWithImage(index)
			if err != nil {
				log.WithField("name", notify.Name).WithField("card", card).Errorf("cast failed %v", err)
				continue
			}
			result = append(result, localutils.MessageTextf("%v发布了新态：\n%v\n%v\n", notify.Name, date, cardImage.GetItem().GetDescription()))
			for _, pic := range cardImage.GetItem().GetPictures() {
				groupImage, err := localutils.UploadGroupImageByUrl(notify.GroupCode, pic.GetImgSrc(), false, proxy_pool.PreferAny)
				if err != nil {
					log.WithField("pic", pic.GetImgSrc()).Errorf("upload image failed %v", err)
					continue
				}
				result = append(result, groupImage)
			}
		case bilibili.DynamicDescType_TextOnly:
			cardText, err := notify.GetCardTextOnly(index)
			if err != nil {
				log.WithField("name", notify.Name).WithField("card", card).Errorf("cast failed %v", err)
				continue
			}
			result = append(result, localutils.MessageTextf("%v发布了新动态：\n%v\n%v\n", notify.Name, date, cardText.GetItem().GetContent()))
		case bilibili.DynamicDescType_WithVideo:
			cardVideo, err := notify.GetCardWithVideo(index)
			if err != nil {
				log.WithField("name", notify.Name).WithField("card", card).Errorf("cast failed %v", err)
				continue
			}
			result = append(result, localutils.MessageTextf("%v发布了新视频：\n%v\n%v\n%v\n", notify.Name, date, cardVideo.GetTitle(), cardVideo.GetDynamic()))
			cover, err := localutils.UploadGroupImageByUrl(notify.GroupCode, cardVideo.GetPic(), true, proxy_pool.PreferAny)
			if err != nil {
				log.WithField("pic", cardVideo.GetPic()).Errorf("upload video cover failed %v", err)
				continue
			}
			result = append(result, cover)
		case bilibili.DynamicDescType_WithPost:
			cardPost, err := notify.GetCardWithPost(index)
			if err != nil {
				log.WithField("name", notify.Name).WithField("card", card).Errorf("cast failed %v", err)
				continue
			}
			result = append(result, localutils.MessageTextf("%v发布了新专栏：\n%v\n%v\n%v...\n", notify.Name, date, cardPost.Title, cardPost.Summary))
			var cover *message.GroupImageElement
			if len(cardPost.GetImageUrls()) >= 1 {
				cover, err = localutils.UploadGroupImageByUrl(notify.GroupCode, cardPost.GetImageUrls()[0], false, proxy_pool.PreferAny)
			} else {
				cover, err = localutils.UploadGroupImageByUrl(notify.GroupCode, cardPost.GetBannerUrl(), false, proxy_pool.PreferAny)
			}
			if err != nil {
				log.WithField("image_url", cardPost.GetImageUrls()).
					WithField("banner_url", cardPost.GetBannerUrl()).
					Errorf("upload image failed %v", err)
				continue
			}
			result = append(result, cover)
		case bilibili.DynamicDescType_WithMusic:
			// TODO
			log.Debugf("not supported")
			result = append(result, localutils.MessageTextf("%v发布了新动态音乐：\n%v\n", notify.Name, date))
		case bilibili.DynamicDescType_WithSketch:
			cardSketch, err := notify.GetCardWithSketch(index)
			if err != nil {
				log.WithField("name", notify.Name).
					WithField("card", card).
					Errorf("cast failed %v", err)
				continue
			}
			result = append(result, localutils.MessageTextf("%v更新了%v：\n%v\n%v\n", notify.Name, cardSketch.GetSketch().GetDescText(), date, cardSketch.GetVest().GetContent()))
		case bilibili.DynamicDescType_WithLive:
			cardLive, err := notify.GetCardWithLive(index)
			if err != nil {
				log.WithField("name", notify.Name).
					WithField("card", card).
					Errorf("cast failed %v", err)
				continue
			}
			result = append(result, localutils.MessageTextf("%v发布了直播信息：\n%v\n%v\n", notify.Name, date, cardLive.GetTitle()))
			cover, err := localutils.UploadGroupImageByUrl(notify.GroupCode, cardLive.GetCover(), true, proxy_pool.PreferAny)
			if err != nil {
				log.WithField("pic", cardLive.GetCover()).Errorf("upload live cover failed %v", err)
				continue
			}
			result = append(result, cover)
		case bilibili.DynamicDescType_WithLiveV2:
			cardLiveV2, err := notify.GetCardWithLiveV2(index)
			if err != nil {
				log.WithField("name", notify.Name).
					WithField("card", card).
					Errorf("case failed %v", err)
				continue
			}
			result = append(result, localutils.MessageTextf("%v发布了直播信息：\n%v\n%v\n", notify.Name, date, cardLiveV2.GetLivePlayInfo().GetTitle()))
			cover, err := localutils.UploadGroupImageByUrl(notify.GroupCode, cardLiveV2.GetLivePlayInfo().GetCover(), true, proxy_pool.PreferAny)
			if err != nil {
				log.WithField("pic", cardLiveV2.GetLivePlayInfo().GetCover()).Errorf("upload live cover failed %v", err)
				continue
			}
			result = append(result, cover)
		default:
			log.WithField("content", card.GetCard()).Info("found new type")
			result = append(result, localutils.MessageTextf("%v发布了新动态：\n%v\n", notify.Name, date))
		}
		log.WithField("uid", notify.Mid).WithField("name", notify.Name).WithField("dynamicUrl", dynamicUrl).Debug("append")
		result = append(result, message.NewText(dynamicUrl+"\n"))
		results = append(results, result...)
	}
	return results
}

func (l *Lsp) notifyDouyuLive(bot *bot.Bot, notify *douyu.ConcernLiveNotify) []message.IMessageElement {
	var result []message.IMessageElement
	switch notify.ShowStatus {
	case douyu.ShowStatus_Living:
		result = append(result, localutils.MessageTextf("斗鱼-%s正在直播【%v】\n", notify.Nickname, notify.RoomName))
		result = append(result, message.NewText(notify.RoomUrl))
		cover, err := localutils.UploadGroupImageByUrl(notify.GroupCode, notify.GetAvatar().GetBig(), false, proxy_pool.PreferAny)
		if err != nil {
			logger.WithField("avatar", notify.GetAvatar().GetBig()).Errorf("upload avatar failed %v", err)
		} else {
			result = append(result, cover)
		}
	case douyu.ShowStatus_NoLiving:
		result = append(result, localutils.MessageTextf("斗鱼-%s暂未直播\n", notify.Nickname))
		result = append(result, message.NewText(notify.RoomUrl))
	}
	return result
}

func (l *Lsp) notifyYoutube(bot *bot.Bot, notify *youtube.ConcernNotify) []message.IMessageElement {
	var result []message.IMessageElement
	if notify.IsLive() {
		if notify.IsLiving() {
			result = append(result, localutils.MessageTextf("YTB-%v正在直播：\n%v\n", notify.ChannelName, notify.VideoTitle))
		} else {
			result = append(result, localutils.MessageTextf("YTB-%v发布了直播预约：\n%v\n时间：%v\n", notify.ChannelName, notify.VideoTitle, localutils.TimestampFormat(notify.VideoTimestamp)))
		}
	} else if notify.IsVideo() {
		result = append(result, localutils.MessageTextf("YTB-%s发布了新视频：\n%v\n", notify.ChannelName, notify.VideoTitle))
	}
	groupImg, err := localutils.UploadGroupImageByUrl(notify.GroupCode, notify.Cover, false, proxy_pool.PreferOversea)
	if err != nil {
		logger.WithField("channel_name", notify.ChannelName).
			WithField("video_id", notify.VideoId).
			WithField("group_code", notify.GroupCode).
			Errorf("upload cover failed %v", err)
	} else {
		result = append(result, groupImg)
	}
	result = append(result, message.NewText(youtube.VideoViewUrl(notify.VideoId)+"\n"))
	return result
}
