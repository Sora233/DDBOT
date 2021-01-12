package lsp

import (
	"bytes"
	"fmt"
	"github.com/Logiase/MiraiGo-Template/bot"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Sora233/Sora233-MiraiGo/lsp/bilibili"
	"github.com/Sora233/Sora233-MiraiGo/lsp/douyu"
	localutils "github.com/Sora233/Sora233-MiraiGo/utils"
	"github.com/asmcos/requests"
	"time"
)

func (l *Lsp) notifyBilibiliLive(bot *bot.Bot, notify *bilibili.ConcernLiveNotify) []message.IMessageElement {
	var result []message.IMessageElement
	switch notify.Status {
	case bilibili.LiveStatus_Living:
		result = append(result, message.NewText(fmt.Sprintf("%s正在直播【%s】\n", notify.Name, notify.LiveTitle)))
		result = append(result, message.NewText(notify.RoomUrl))
		coverResp, err := requests.Get(notify.Cover)
		if err == nil {
			if cover, err := bot.UploadGroupImage(notify.GroupCode, bytes.NewReader(coverResp.Content())); err == nil {
				result = append(result, cover)
			}
		}
	case bilibili.LiveStatus_NoLiving:
		result = append(result, message.NewText(fmt.Sprintf("%s暂未直播\n", notify.Name)))
		result = append(result, message.NewText(notify.RoomUrl))
	}
	return result
}

func (l *Lsp) notifyBilibiliNews(bot *bot.Bot, notify *bilibili.ConcernNewsNotify) []message.IMessageElement {
	var result []message.IMessageElement
	for index, card := range notify.Cards {
		log := logger.WithField("DescType", card.GetDesc().GetType().String())
		dynamicUrl := bilibili.DynamicUrl(card.GetDesc().GetDynamicIdStr())
		date := time.Unix(int64(card.GetDesc().GetTimestamp()), 0).Format("2006-01-02 15:04:05")
		switch card.GetDesc().GetType() {
		case bilibili.DynamicDescType_WithOrigin:
			cardOrigin, err := notify.GetCardWithOrig(index)
			if err != nil {
				log.WithField("name", notify.Name).WithField("card", card).Errorf("cast failed %v", err)
				continue
			}
			originName := cardOrigin.GetOriginUser().GetInfo().GetUname()
			result = append(result, message.NewText(fmt.Sprintf("%v转发了%v的动态：\n%v\n", notify.Name, originName, cardOrigin.GetItem().GetContent())))
		case bilibili.DynamicDescType_WithImage:
			cardImage, err := notify.GetCardWithImage(index)
			if err != nil {
				log.WithField("name", notify.Name).WithField("card", card).Errorf("cast failed %v", err)
				continue
			}
			cardImage.GetItem()
			result = append(result, message.NewText(fmt.Sprintf("%v发布了新态：\n%v\n%v\n", notify.Name, date, cardImage.GetItem().GetDescription())))
			if cardImage.GetItem().GetPicturesCount() == 1 {
				pic := cardImage.GetItem().GetPictures()[0]
				img, err := localutils.ImageGet(pic.GetImgSrc())
				if err != nil {
					log.WithField("pic", pic).Errorf("get image failed %v", err)
					continue
				}
				groupImage, err := bot.UploadGroupImage(notify.GroupCode, bytes.NewReader(img))
				if err != nil {
					log.WithField("pic", pic).Errorf("upload group image %v", err)
					continue
				}
				result = append(result, groupImage)
			} else {
				for _, pic := range cardImage.GetItem().GetPictures() {
					img, err := localutils.ImageGetAndNorm(pic.GetImgSrc())
					if err != nil {
						log.WithField("pic", pic).Errorf("get image failed %v", err)
						continue
					}
					groupImage, err := bot.UploadGroupImage(notify.GroupCode, bytes.NewReader(img))
					if err != nil {
						log.WithField("pic", pic).Errorf("upload group image %v", err)
						continue
					}
					result = append(result, groupImage)
				}
			}
		case bilibili.DynamicDescType_TextOnly:
			cardText, err := notify.GetCardTextOnly(index)
			if err != nil {
				log.WithField("name", notify.Name).WithField("card", card).Errorf("cast failed %v", err)
				continue
			}
			result = append(result, message.NewText(fmt.Sprintf("%v发布了新动态：\n%v\n%v\n", notify.Name, date, cardText.GetItem().GetContent())))
		case bilibili.DynamicDescType_WithVideo:
			cardVideo, err := notify.GetCardWithVideo(index)
			if err != nil {
				log.WithField("name", notify.Name).WithField("card", card).Errorf("cast failed %v", err)
				continue
			}
			result = append(result, message.NewText(fmt.Sprintf("%v发布了新视频：\n%v\n%v\n%v\n", notify.Name, date, cardVideo.GetTitle(), cardVideo.GetDynamic())))
			img, err := localutils.ImageGetAndNorm(cardVideo.GetPic())
			if err != nil {
				log.WithField("pic", cardVideo.GetPic()).Errorf("get image failed %v", err)
				continue
			}
			cover, err := bot.UploadGroupImage(notify.GroupCode, bytes.NewReader(img))
			if err != nil {
				log.WithField("pic", cardVideo.GetPic()).Errorf("upload group image failed %v", err)
				continue
			}
			result = append(result, cover)
		case bilibili.DynamicDescType_WithPost:
			// TODO
			log.Debugf("not supported")
		case bilibili.DynamicDescType_WithMusic:
			// TODO
			log.Debugf("not supported")
		}
		log.WithField("dynamicUrl", dynamicUrl).Debug("append")
		result = append(result, message.NewText(dynamicUrl+"\n"))
	}
	return result
}

func (l *Lsp) notifyDouyuLive(bot *bot.Bot, notify *douyu.ConcernLiveNotify) []message.IMessageElement {
	var result []message.IMessageElement
	switch notify.ShowStatus {
	case douyu.ShowStatus_Living:
		result = append(result, message.NewText(fmt.Sprintf("斗鱼-%s正在直播【%s】\n", notify.Nickname, notify.RoomName)))
		result = append(result, message.NewText(notify.RoomUrl))
		coverResp, err := requests.Get(notify.GetAvatar().GetBig())
		if err == nil {
			if cover, err := bot.UploadGroupImage(notify.GroupCode, bytes.NewReader(coverResp.Content())); err == nil {
				result = append(result, cover)
			}
		}
	case douyu.ShowStatus_NoLiving:
		result = append(result, message.NewText(fmt.Sprintf("斗鱼-%s暂未直播\n", notify.Nickname)))
		result = append(result, message.NewText(notify.RoomUrl))
	}
	return result
}
