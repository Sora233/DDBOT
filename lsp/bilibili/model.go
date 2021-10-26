package bilibili

import (
	"errors"
	"github.com/Logiase/MiraiGo-Template/config"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Sora233/DDBOT/concern"
	"github.com/Sora233/DDBOT/proxy_pool"
	localutils "github.com/Sora233/DDBOT/utils"
	"github.com/Sora233/DDBOT/utils/blockCache"
	"github.com/sirupsen/logrus"
	"strings"
	"sync"
)

type NewsInfo struct {
	UserInfo
	LastDynamicId int64   `json:"last_dynamic_id"`
	Timestamp     int64   `json:"timestamp"`
	Cards         []*Card `json:"-"`
}

func (n *NewsInfo) Type() EventType {
	return News
}

func (n *NewsInfo) checkIndex(index int) error {
	if len(n.Cards) <= index || n.Cards[index].GetCard() == "" {
		return errors.New("card not found or empty")
	}
	return nil
}

func (n *NewsInfo) GetCardWithImage(index int) (*CardWithImage, error) {
	if err := n.checkIndex(index); err != nil {
		return nil, err
	}
	return n.Cards[index].GetCardWithImage()
}

func (n *NewsInfo) GetCardWithOrig(index int) (*CardWithOrig, error) {
	if err := n.checkIndex(index); err != nil {
		return nil, err
	}
	return n.Cards[index].GetCardWithOrig()
}

func (n *NewsInfo) GetCardWithVideo(index int) (*CardWithVideo, error) {
	if err := n.checkIndex(index); err != nil {
		return nil, err
	}
	return n.Cards[index].GetCardWithVideo()
}

func (n *NewsInfo) GetCardTextOnly(index int) (*CardTextOnly, error) {
	if err := n.checkIndex(index); err != nil {
		return nil, err
	}
	return n.Cards[index].GetCardTextOnly()
}

func (n *NewsInfo) GetCardWithPost(index int) (*CardWithPost, error) {
	if err := n.checkIndex(index); err != nil {
		return nil, err
	}
	return n.Cards[index].GetCardWithPost()
}

func (n *NewsInfo) GetCardWithMusic(index int) (*CardWithMusic, error) {
	if err := n.checkIndex(index); err != nil {
		return nil, err
	}
	return n.Cards[index].GetCardWithMusic()
}

func (n *NewsInfo) GetCardWithSketch(index int) (*CardWithSketch, error) {
	if err := n.checkIndex(index); err != nil {
		return nil, err
	}
	return n.Cards[index].GetCardWithSketch()
}

func (n *NewsInfo) GetCardWithLive(index int) (*CardWithLive, error) {
	if err := n.checkIndex(index); err != nil {
		return nil, err
	}
	return n.Cards[index].GetCardWithLive()
}

func (n *NewsInfo) GetCardWithLiveV2(index int) (*CardWithLiveV2, error) {
	if err := n.checkIndex(index); err != nil {
		return nil, err
	}
	return n.Cards[index].GetCardWithLiveV2()
}

func (n *NewsInfo) ToString() string {
	if n == nil {
		return ""
	}
	content, _ := json.Marshal(n)
	return string(content)
}

func (n *NewsInfo) Logger() *logrus.Entry {
	return logger.WithFields(logrus.Fields{
		"Site":     Site,
		"Mid":      n.Mid,
		"Name":     n.Name,
		"CardSize": len(n.Cards),
		"Type":     "News",
	})
}

type ConcernNewsNotify struct {
	GroupCode int64 `json:"group_code"`
	// NewsInfo len(Cards)必须为1
	*UserInfo
	Card *Card

	// messageCache 导致ConcernNewsNotify的ToMessage()变得线程不安全
	messageCache []message.IMessageElement
	// 用于联合投稿和转发的时候防止多人同时推送
	shouldCompact bool
	compactKey    string
	concern       *Concern
}

func (notify *ConcernNewsNotify) Type() concern.Type {
	return concern.BilibiliNews
}

type ConcernLiveNotify struct {
	GroupCode int64 `json:"group_code"`
	LiveInfo
}

func (notify *ConcernLiveNotify) Type() concern.Type {
	return concern.BibiliLive
}

type UserStat struct {
	Mid int64 `json:"mid"`
	// 关注数
	Following int64 `json:"following"`
	// 粉丝数
	Follower int64 `json:"follower"`
}

func (us *UserStat) ToString() string {
	if us == nil {
		return ""
	}
	content, _ := json.Marshal(us)
	return string(content)
}

type UserInfo struct {
	Mid     int64  `json:"mid"`
	Name    string `json:"name"`
	RoomId  int64  `json:"room_id"`
	RoomUrl string `json:"room_url"`

	UserStat *UserStat `json:"-"`
}

func (ui *UserInfo) GetName() string {
	if ui == nil {
		return ""
	}
	return ui.Name
}

func (ui *UserInfo) ToString() string {
	if ui == nil {
		return ""
	}
	content, _ := json.Marshal(ui)
	return string(content)
}

type LiveInfo struct {
	UserInfo
	Status    LiveStatus `json:"status"`
	LiveTitle string     `json:"live_title"`
	Cover     string     `json:"cover"`

	LiveStatusChanged bool `json:"-"`
	LiveTitleChanged  bool `json:"-"`
}

func (l *LiveInfo) Living() bool {
	if l == nil {
		return false
	}
	return l.Status == LiveStatus_Living
}

func (l *LiveInfo) Type() EventType {
	return Live
}

func (l *LiveInfo) ToString() string {
	if l == nil {
		return ""
	}
	content, _ := json.Marshal(l)
	return string(content)
}

func (l *LiveInfo) Logger() *logrus.Entry {
	return logger.WithFields(logrus.Fields{
		"Site":   Site,
		"Mid":    l.Mid,
		"Name":   l.Name,
		"RoomId": l.RoomId,
		"Title":  l.LiveTitle,
		"Status": l.Status.String(),
		"Type":   "Live",
	})
}

func NewUserStat(mid, following, follower int64) *UserStat {
	return &UserStat{
		Mid:       mid,
		Following: following,
		Follower:  follower,
	}
}

func NewUserInfo(mid, roomId int64, name, url string) *UserInfo {
	return &UserInfo{
		Mid:     mid,
		RoomId:  roomId,
		Name:    name,
		RoomUrl: url,
	}
}

func NewLiveInfo(userInfo *UserInfo, liveTitle string, cover string, status LiveStatus) *LiveInfo {
	if userInfo == nil {
		return nil
	}
	return &LiveInfo{
		UserInfo:  *userInfo,
		Status:    status,
		LiveTitle: liveTitle,
		Cover:     cover,
	}
}

func NewNewsInfo(userInfo *UserInfo, dynamicId int64, timestamp int64) *NewsInfo {
	if userInfo == nil {
		return nil
	}
	return &NewsInfo{
		UserInfo:      *userInfo,
		LastDynamicId: dynamicId,
		Timestamp:     timestamp,
	}
}

func NewNewsInfoWithDetail(userInfo *UserInfo, cards []*Card) *NewsInfo {
	var dynamicId int64
	var timestamp int64
	if len(cards) > 0 {
		dynamicId = cards[0].GetDesc().GetDynamicId()
		timestamp = cards[0].GetDesc().GetTimestamp()
	}
	return &NewsInfo{
		UserInfo:      *userInfo,
		LastDynamicId: dynamicId,
		Timestamp:     timestamp,
		Cards:         cards,
	}
}

func NewConcernNewsNotify(groupCode int64, newsInfo *NewsInfo, c *Concern) []*ConcernNewsNotify {
	if newsInfo == nil {
		return nil
	}
	var result []*ConcernNewsNotify
	for _, card := range newsInfo.Cards {
		result = append(result, &ConcernNewsNotify{
			GroupCode: groupCode,
			UserInfo:  &newsInfo.UserInfo,
			Card:      card,
			concern:   c,
		})
	}
	return result
}

func NewConcernLiveNotify(groupCode int64, liveInfo *LiveInfo) *ConcernLiveNotify {
	if liveInfo == nil {
		return nil
	}
	return &ConcernLiveNotify{
		GroupCode: groupCode,
		LiveInfo:  *liveInfo,
	}
}

func (notify *ConcernNewsNotify) ToMessage() (result []message.IMessageElement) {
	var (
		card       = notify.Card
		log        = notify.Logger()
		dynamicUrl = DynamicUrl(card.GetDesc().GetDynamicIdStr())
		date       = localutils.TimestampFormat(card.GetDesc().GetTimestamp())
	)
	// 推送一条简化动态防止刷屏，主要是联合投稿和转发的时候
	if notify.shouldCompact {
		// 通过回复之前消息的方式简化推送
		msg, _ := notify.concern.GetNotifyMsg(notify.GroupCode, notify.compactKey)
		if msg != nil {
			result = append(result, message.NewReply(msg))
		}
		log.WithField("compact_key", notify.compactKey).Debug("compact notify")
		switch notify.Card.GetDesc().GetType() {
		case DynamicDescType_WithVideo:
			videoCard, _ := notify.Card.GetCardWithVideo()
			result = append(result,
				localutils.MessageTextf("%v%v：\n%v\n%v\n%v",
					notify.Name,
					notify.Card.GetDisplay().GetUsrActionTxt(),
					date,
					videoCard.GetTitle(),
					dynamicUrl),
			)
			return result
		case DynamicDescType_WithOrigin:
			origCard, _ := notify.Card.GetCardWithOrig()
			result = append(result,
				localutils.MessageTextf("%v转发了%v的动态：\n%v\n%v\n%v",
					notify.Name,
					origCard.GetOriginUser().GetInfo().GetUname(),
					date,
					origCard.GetItem().GetContent(),
					dynamicUrl,
				),
			)
			return result
		}
	}
	if notify.messageCache != nil {
		return notify.messageCache
	}
	switch card.GetDesc().GetType() {
	case DynamicDescType_WithOrigin:
		cardOrigin, err := card.GetCardWithOrig()
		if err != nil {
			log.WithField("card", card).Errorf("GetCardWithOrig failed %v", err)
			return
		}
		originName := cardOrigin.GetOriginUser().GetInfo().GetUname()
		// very sb
		switch cardOrigin.GetItem().GetOrigType() {
		case DynamicDescType_WithImage:
			result = append(result, localutils.MessageTextf("%v转发了%v的动态：\n%v\n%v\n\n原动态：\n", notify.Name, originName, date, cardOrigin.GetItem().GetContent()))
			origin := new(CardWithImage)
			err := json.Unmarshal([]byte(cardOrigin.GetOrigin()), origin)
			if err != nil {
				log.WithField("origin", cardOrigin.GetOrigin()).Errorf("Unmarshal origin cardWithImage failed %v", err)
				return
			}
			result = append(result, localutils.MessageTextf("%v\n", origin.GetItem().GetDescription()))
			var skip = false
			if shouldCombineImage(origin.GetItem().GetPictures()) {
				var urls = make([]string, len(origin.GetItem().GetPictures()))
				for index, pic := range origin.GetItem().GetPictures() {
					urls[index] = pic.GetImgSrc()
				}
				resultByte, err := urlsMergeImage(urls)
				if err != nil {
					log.Errorf("urlsMergeImage failed %v", err)
				} else {
					groupImage, err := localutils.UploadGroupImage(notify.GroupCode, resultByte, false)
					if err != nil {
						log.Errorf("upload 9Image group image %v", err)
					} else {
						result = append(result, groupImage)
						skip = true
					}
				}
			}
			if !skip {
				for _, pic := range origin.GetItem().GetPictures() {
					var isNorm = false
					if pic.ImgHeight > 2560 && pic.ImgWidth > 2560 {
						isNorm = true
					}
					groupImage, err := localutils.UploadGroupImageByUrl(notify.GroupCode, pic.GetImgSrc(), isNorm, proxy_pool.PreferNone)
					if err != nil {
						log.WithField("pic", pic).Errorf("upload group image %v", err)
						continue
					}
					result = append(result, groupImage)
				}
			}
		case DynamicDescType_TextOnly:
			result = append(result, localutils.MessageTextf("%v转发了%v的动态：\n%v\n%v\n\n原动态：\n", notify.Name, originName, date, cardOrigin.GetItem().GetContent()))
			origin := new(CardTextOnly)
			err := json.Unmarshal([]byte(cardOrigin.GetOrigin()), origin)
			if err != nil {
				log.WithField("origin", cardOrigin.GetOrigin()).Errorf("Unmarshal origin cardWithText failed %v", err)
				return
			}
			result = append(result, localutils.MessageTextf("%v\n", origin.GetItem().GetContent()))
		case DynamicDescType_WithVideo:
			result = append(result, localutils.MessageTextf("%v转发了%v的投稿：\n%v\n%v\n\n原视频：\n", notify.Name, originName, date, cardOrigin.GetItem().GetContent()))
			origin := new(CardWithVideo)
			err := json.Unmarshal([]byte(cardOrigin.GetOrigin()), origin)
			if err != nil {
				log.WithField("origin", cardOrigin.GetOrigin()).Errorf("Unmarshal origin cardWithVideo failed %v", err)
				return
			}
			result = append(result, localutils.MessageTextf("%v\n%v\n", origin.GetTitle(), origin.GetDesc()))
			cover, err := localutils.UploadGroupImageByUrl(notify.GroupCode, origin.GetPic(), true, proxy_pool.PreferNone)
			if err != nil {
				log.Errorf("upload video cover failed %v", err)
				result = append(result, message.NewText("[封面]\n"))
			} else {
				result = append(result, cover)
			}
		case DynamicDescType_WithPost:
			result = append(result, localutils.MessageTextf("%v转发了%v的专栏：\n%v\n%v\n\n原专栏：\n", notify.Name, originName, date, cardOrigin.GetItem().GetContent()))
			origin := new(CardWithPost)
			err := json.Unmarshal([]byte(cardOrigin.GetOrigin()), origin)
			if err != nil {
				log.WithField("origin", cardOrigin.GetOrigin()).Errorf("Unmarshal origin cardWithPost failed %v", err)
				return
			}
			result = append(result, localutils.MessageTextf("%v\n%v\n", origin.GetTitle(), origin.GetSummary()))
			var cover *message.GroupImageElement
			if len(origin.GetImageUrls()) >= 1 {
				cover, err = localutils.UploadGroupImageByUrl(notify.GroupCode, origin.GetImageUrls()[0], false, proxy_pool.PreferNone)
			} else {
				cover, err = localutils.UploadGroupImageByUrl(notify.GroupCode, origin.GetBannerUrl(), false, proxy_pool.PreferNone)
			}
			if err != nil {
				log.WithFields(logrus.Fields{
					"image_url":  origin.GetImageUrls(),
					"banner_url": origin.GetBannerUrl(),
				}).Errorf("upload image failed %v", err)
			} else {
				result = append(result, cover)
			}
		case DynamicDescType_WithMusic:
			origin := new(CardWithMusic)
			err := json.Unmarshal([]byte(cardOrigin.GetOrigin()), origin)
			if err != nil {
				log.WithField("origin", cardOrigin.GetOrigin()).Errorf("Unmarshal origin CardWithMusic failed %v", err)
				return
			}
			result = append(result, localutils.MessageTextf(
				"%v转发了%v的音频：\n%v\n%v\n\n原音频：\n",
				notify.Name,
				originName,
				date,
				cardOrigin.GetItem().GetContent(),
			))
			result = append(result, localutils.MessageTextf("%v\n%v\n", origin.GetTitle(), origin.GetIntro()))
			if len(origin.GetCover()) != 0 {
				cover, err := localutils.UploadGroupImageByUrl(notify.GroupCode, origin.GetCover(), false, proxy_pool.PreferNone)
				if err != nil {
					log.WithField("cover", origin.GetCover()).Errorf("upload music cover failed %v", err)
				} else {
					result = append(result, cover)
				}
			}
		case DynamicDescType_WithSketch:
			origin := new(CardWithSketch)
			err := json.Unmarshal([]byte(cardOrigin.GetOrigin()), origin)
			if err != nil {
				log.WithField("origin", cardOrigin.GetOrigin()).Errorf("Unmarshal origin CardWithSketch failed %v", err)
				return
			}
			result = append(result, localutils.MessageTextf("%v转发了%v的动态：\n%v\n%v\n原动态：\n%v\n%v\n%v", notify.Name, originName, date, cardOrigin.GetItem().GetContent(),
				origin.GetVest().GetContent(), origin.GetSketch().GetTitle(), origin.GetSketch().GetDescText()))
			if len(origin.GetSketch().GetCoverUrl()) != 0 {
				cover, err := localutils.UploadGroupImageByUrl(notify.GroupCode, origin.GetSketch().GetCoverUrl(), true, proxy_pool.PreferNone)
				if err != nil {
					log.WithField("pic", origin.GetSketch().GetCoverUrl()).
						Errorf("upload sketch cover failed %v", err)
				} else {
					result = append(result, cover)
				}
			}
		case DynamicDescType_WithLive:
			result = append(result, localutils.MessageTextf("%v分享了%v的直播：\n%v\n%v\n\n原直播间：\n", notify.Name, originName, date, cardOrigin.GetItem().GetContent()))
			origin := new(CardWithLive)
			err := json.Unmarshal([]byte(cardOrigin.GetOrigin()), origin)
			if err != nil {
				log.WithField("origin", cardOrigin.GetOrigin()).Errorf("Unmarshal origin CardWithLive failed %v", err)
				return
			}
			result = append(result, localutils.MessageTextf("%v\n", origin.GetTitle()))
			groupImage, err := localutils.UploadGroupImageByUrl(notify.GroupCode, origin.GetCover(), false, proxy_pool.PreferNone)
			if err != nil {
				log.Errorf("upload live cover failed %v", err)
				result = append(result, message.NewText("[封面]\n"))
			} else {
				result = append(result, groupImage)
			}
		case DynamicDescType_WithLiveV2:
			result = append(result, localutils.MessageTextf("%v分享了%v的直播：\n%v\n%v\n\n原直播间：\n", notify.Name, originName, date, cardOrigin.GetItem().GetContent()))
			origin := new(CardWithLiveV2)
			err := json.Unmarshal([]byte(cardOrigin.GetOrigin()), origin)
			if err != nil {
				log.WithField("origin", cardOrigin.GetOrigin()).Errorf("Unmarshal origin CardWithLiveV2 failed %v", err)
				return
			}
			result = append(result, localutils.MessageTextf("%v\n", origin.GetLivePlayInfo().GetTitle()))
			groupImage, err := localutils.UploadGroupImageByUrl(notify.GroupCode, origin.GetLivePlayInfo().GetCover(), false, proxy_pool.PreferNone)
			if err != nil {
				log.WithField("origin", cardOrigin.GetOrigin()).Errorf("upload liveV2 cover failed %v", err)
			} else {
				result = append(result, groupImage)
			}
		case DynamicDescType_WithMylist:
			result = append(result, localutils.MessageTextf("%v分享了%v的收藏夹：\n%v\n%v\n\n原收藏夹：\n", notify.Name, originName, date, cardOrigin.GetItem().GetContent()))
			origin := new(CardWithMylist)
			err := json.Unmarshal([]byte(cardOrigin.GetOrigin()), origin)
			if err != nil {
				log.WithField("origin", cardOrigin.GetOrigin()).Errorf("Unmarshal origin CardWithMylist failed %v", err)
				return
			}
			result = append(result, localutils.MessageTextf("%v\n", origin.GetTitle()))
			groupImage, err := localutils.UploadGroupImageByUrl(notify.GroupCode, origin.GetCover(), false, proxy_pool.PreferNone)
			if err != nil {
				log.WithField("origin", cardOrigin.GetOrigin()).Errorf("upload CardWithMylist cover failed %v", err)
			} else {
				result = append(result, groupImage)
			}
		case DynamicDescType_WithMiss:
			result = append(result, localutils.MessageTextf("%v分享了动态：\n%v\n%v\n\n%v\n", notify.Name, date, cardOrigin.GetItem().GetContent(), cardOrigin.GetItem().GetTips()))
		case DynamicDescType_WithOrigin:
			// 麻了，套起来了
			result = append(result, localutils.MessageTextf("%v转发了%v的动态：%v\n%v\n", notify.Name, originName, date, cardOrigin.GetItem().GetContent()))
		case DynamicDescType_WithCourse:
			origin := new(CardWithCourse)
			err := json.Unmarshal([]byte(cardOrigin.GetOrigin()), origin)
			if err != nil {
				log.WithField("origin", cardOrigin.GetOrigin()).Errorf("Unmarshal origin CardWithCourse failed %v", err)
				return
			}
			result = append(result, localutils.MessageTextf("%v转发了%v的%v：\n%v\n%v\n\n原课程：\n%v", notify.Name,
				origin.GetUpInfo().GetName(),
				origin.GetBadge().GetText(),
				date,
				cardOrigin.GetItem().GetContent(),
				origin.GetTitle(),
			))
			groupImage, err := localutils.UploadGroupImageByUrl(notify.GroupCode, origin.GetCover(), false, proxy_pool.PreferNone)
			if err != nil {
				log.WithField("origin", cardOrigin.GetOrigin()).Errorf("upload CardWithCourse cover failed %v", err)
			} else {
				result = append(result, groupImage)
			}
		default:
			// 试试media
			origin := new(CardWithMedia)
			err := json.Unmarshal([]byte(cardOrigin.GetOrigin()), origin)
			if err == nil && origin.GetApiSeasonInfo() != nil {
				var desc = origin.GetNewDesc()
				if len(desc) == 0 {
					desc = origin.GetIndex()
				}
				result = append(result, localutils.MessageTextf("%v转发了%v【%v】%v：\n%v\n%v\n", notify.Name,
					origin.GetApiSeasonInfo().GetTypeName(),
					origin.GetApiSeasonInfo().GetTitle(),
					desc,
					date,
					cardOrigin.GetItem().GetContent()),
				)
				groupImage, err := localutils.UploadGroupImageByUrl(notify.GroupCode, origin.GetCover(), true, proxy_pool.PreferNone)
				if err != nil {
					log.WithField("origin", cardOrigin.GetOrigin()).Errorf("upload CardWithMedia cover failed %v", err)
				} else {
					result = append(result, groupImage)
				}
			} else {
				log.WithField("content", card.GetCard()).Info("found new type with origin")
				result = append(result, localutils.MessageTextf("%v转发了%v的动态：\n%v\n%v\n", notify.Name, originName, date, cardOrigin.GetItem().GetContent()))
			}
		}
	case DynamicDescType_WithImage:
		cardImage, err := card.GetCardWithImage()
		if err != nil {
			log.WithField("card", card).Errorf("GetCardWithImage cast failed %v", err)
			return
		}
		result = append(result, localutils.MessageTextf("%v发布了新动态：\n%v\n%v\n", notify.Name, date, cardImage.GetItem().GetDescription()))
		var skip = false
		if shouldCombineImage(cardImage.GetItem().GetPictures()) {
			var urls = make([]string, len(cardImage.GetItem().GetPictures()))
			for index, pic := range cardImage.GetItem().GetPictures() {
				urls[index] = pic.GetImgSrc()
			}
			resultByte, err := urlsMergeImage(urls)
			if err != nil {
				log.Errorf("urlsMergeImage failed %v", err)
			} else {
				groupImage, err := localutils.UploadGroupImage(notify.GroupCode, resultByte, false)
				if err != nil {
					log.Errorf("upload 9Image group image %v", err)
				} else {
					result = append(result, groupImage)
					skip = true
				}
			}
		}
		if !skip {
			for _, pic := range cardImage.GetItem().GetPictures() {
				var isNorm = false
				if pic.GetImgHeight() > 1200 && pic.GetImgWidth() > 1200 {
					isNorm = true
				}
				groupImage, err := localutils.UploadGroupImageByUrl(notify.GroupCode, pic.GetImgSrc(), isNorm, proxy_pool.PreferNone)
				if err != nil {
					log.WithField("pic", pic.GetImgSrc()).Errorf("upload image failed %v", err)
					continue
				}
				result = append(result, groupImage)
			}
		}

	case DynamicDescType_TextOnly:
		cardText, err := card.GetCardTextOnly()
		if err != nil {
			log.WithField("card", card).Errorf("GetCardTextOnly cast failed %v", err)
			return
		}
		result = append(result, localutils.MessageTextf("%v发布了新动态：\n%v\n%v\n", notify.Name, date, cardText.GetItem().GetContent()))
	case DynamicDescType_WithVideo:
		cardVideo, err := card.GetCardWithVideo()
		if err != nil {
			log.WithField("card", card).Errorf("GetCardWithVideo cast failed %v", err)
			return
		}
		description := strings.TrimSpace(cardVideo.GetDynamic())
		if description == "" {
			description = cardVideo.GetDesc()
		}
		result = append(result, localutils.MessageTextf("%v%v：\n%v\n%v\n%v\n", notify.Name, card.GetDisplay().GetUsrActionTxt(), date, cardVideo.GetTitle(), description))
		cover, err := localutils.UploadGroupImageByUrl(notify.GroupCode, cardVideo.GetPic(), true, proxy_pool.PreferNone)
		if err != nil {
			log.WithField("pic", cardVideo.GetPic()).Errorf("upload video cover failed %v", err)
		} else {
			result = append(result, cover)
		}
	case DynamicDescType_WithPost:
		cardPost, err := card.GetCardWithPost()
		if err != nil {
			log.WithField("card", card).Errorf("GetCardWithPost cast failed %v", err)
			return
		}
		result = append(result, localutils.MessageTextf("%v发布了新专栏：\n%v\n%v\n%v...\n", notify.Name, date, cardPost.Title, cardPost.Summary))
		var cover *message.GroupImageElement
		if len(cardPost.GetImageUrls()) >= 1 {
			cover, err = localutils.UploadGroupImageByUrl(notify.GroupCode, cardPost.GetImageUrls()[0], false, proxy_pool.PreferNone)
		} else {
			cover, err = localutils.UploadGroupImageByUrl(notify.GroupCode, cardPost.GetBannerUrl(), false, proxy_pool.PreferNone)
		}
		if err != nil {
			log.WithField("image_url", cardPost.GetImageUrls()).
				WithField("banner_url", cardPost.GetBannerUrl()).
				Errorf("upload image failed %v", err)
		} else {
			result = append(result, cover)
		}
	case DynamicDescType_WithMusic:
		cardMusic, err := card.GetCardWithMusic()
		if err != nil {
			log.WithField("card", card).
				Errorf("GetCardWithMusic cast failed %v", err)
			return
		}
		result = append(result, localutils.MessageTextf(
			"%v投稿了新音频：\n%v\n%v\n%v\n",
			notify.Name,
			date,
			cardMusic.GetTitle(),
			cardMusic.GetIntro(),
		))
		cover, err := localutils.UploadGroupImageByUrl(notify.GroupCode, cardMusic.GetCover(), false, proxy_pool.PreferNone)
		if err != nil {
			log.WithField("cover", cardMusic.GetCover()).Errorf("upload image failed %v", err)
		} else {
			result = append(result, cover)
		}
	case DynamicDescType_WithSketch:
		cardSketch, err := card.GetCardWithSketch()
		if err != nil {
			log.WithField("card", card).
				Errorf("GetCardWithSketch cast failed %v", err)
			return
		}
		result = append(result, localutils.MessageTextf(
			"%v发表了新动态：\n%v\n%v\n",
			notify.Name,
			date,
			cardSketch.GetVest().GetContent(),
		))
		if cardSketch.GetSketch().GetTitle() == cardSketch.GetSketch().GetDescText() {
			result = append(result, localutils.MessageTextf(
				"内容：%v", cardSketch.GetSketch().GetTitle()))
		} else {
			result = append(result, localutils.MessageTextf(
				"内容：%v - %v", cardSketch.GetSketch().GetTitle(), cardSketch.GetSketch().GetDescText()))
		}
		if len(cardSketch.GetSketch().GetCoverUrl()) != 0 {
			cover, err := localutils.UploadGroupImageByUrl(notify.GroupCode, cardSketch.GetSketch().GetCoverUrl(), true, proxy_pool.PreferNone)
			if err != nil {
				log.WithField("pic", cardSketch.GetSketch().GetCoverUrl()).
					Errorf("upload sketch cover failed %v", err)
			} else {
				result = append(result, cover)
			}
		}
	case DynamicDescType_WithLive:
		cardLive, err := card.GetCardWithLive()
		if err != nil {
			log.WithField("card", card).
				Errorf("GetCardWithLive cast failed %v", err)
			return
		}
		result = append(result, localutils.MessageTextf("%v发布了直播信息：\n%v\n%v\n", notify.Name, date, cardLive.GetTitle()))
		cover, err := localutils.UploadGroupImageByUrl(notify.GroupCode, cardLive.GetCover(), true, proxy_pool.PreferNone)
		if err != nil {
			log.WithField("pic", cardLive.GetCover()).
				Errorf("upload live cover failed %v", err)
			result = append(result, message.NewText("[封面]\n"))
		} else {
			result = append(result, cover)
		}
	case DynamicDescType_WithLiveV2:
		// 2021-08-15 发现这个是系统推荐的直播间，应该不是人为操作，选择不推送，在filter中过滤
		cardLiveV2, err := card.GetCardWithLiveV2()
		if err != nil {
			log.WithField("card", card).
				Errorf("GetCardWithLiveV2 case failed %v", err)
			return
		}
		result = append(result, localutils.MessageTextf("%v发布了直播信息：\n%v\n%v\n", notify.Name, date, cardLiveV2.GetLivePlayInfo().GetTitle()))
		cover, err := localutils.UploadGroupImageByUrl(notify.GroupCode, cardLiveV2.GetLivePlayInfo().GetCover(), true, proxy_pool.PreferNone)
		if err != nil {
			log.WithField("pic", cardLiveV2.GetLivePlayInfo().GetCover()).
				Errorf("upload live cover failed %v", err)
			result = append(result, message.NewText("[封面]\n"))
		} else {
			result = append(result, cover)
		}
	case DynamicDescType_WithMiss:
		cardWithMiss, err := card.GetCardWithOrig()
		if err != nil {
			log.WithField("card", card).
				Errorf("GetCardWithOrig case failed %v", err)
			return
		}
		result = append(result, localutils.MessageTextf("%v发布了新动态：\n%v\n%v\n\n%v\n", notify.Name, date, cardWithMiss.GetItem().GetContent(), cardWithMiss.GetItem().GetTips()))
	default:
		log.WithField("content", card.GetCard()).Info("found new DynamicDescType")
		result = append(result, localutils.MessageTextf("%v发布了新动态：\n%v\n", notify.Name, date))
	}

	// 2021/04/16发现了有新增一个预约卡片
	for _, addons := range [][]*Card_Display_AddOnCardInfo{
		card.GetDisplay().GetAddOnCardInfo(),
		card.GetDisplay().GetOrigin().GetAddOnCardInfo(),
	} {
		for _, addon := range addons {
			switch addon.AddOnCardShowType {
			case AddOnCardShowType_goods:
				goodsCard := new(Card_Display_AddOnCardInfo_GoodsCard)
				if err := json.Unmarshal([]byte(addon.GetGoodsCard()), goodsCard); err != nil {
					log.WithField("goods", addon.GetGoodsCard()).Errorf("Unmarshal goods card failed %v", err)
					continue
				}
				if len(goodsCard.GetList()) == 0 {
					continue
				}
				var item = goodsCard.GetList()[0]
				result = append(result, localutils.MessageTextf("\n%v：\n%v\n", item.AdMark, item.Name))
				cover, err := localutils.UploadGroupImageByUrl(notify.GroupCode, item.GetImg(), true, proxy_pool.PreferNone)
				if err != nil {
					log.WithField("img", item.GetImg()).Errorf("update goods img failed %v", err)
				} else {
					result = append(result, cover)
				}
			case AddOnCardShowType_reserve:
				if len(addon.GetReserveAttachCard().GetReserveLottery().GetText()) == 0 {
					result = append(result, localutils.MessageTextf("\n附加信息：\n%v\n%v\n",
						addon.GetReserveAttachCard().GetTitle(),
						addon.GetReserveAttachCard().GetDescFirst().GetText()))
				} else {
					result = append(result, localutils.MessageTextf("\n附加信息：\n%v\n%v\n%v\n",
						addon.GetReserveAttachCard().GetTitle(),
						addon.GetReserveAttachCard().GetDescFirst().GetText(),
						addon.GetReserveAttachCard().GetReserveLottery().GetText()))
				}
			case AddOnCardShowType_match:
			// TODO 暂时没必要
			case AddOnCardShowType_related:
				aCard := addon.GetAttachCard()
				// 游戏应该不需要
				if aCard.GetType() != "game" {
					result = append(result, localutils.MessageTextf("\n%v：\n%v\n%v\n",
						aCard.GetHeadText(),
						aCard.GetTitle(),
						aCard.GetDescFirst(),
					))
				}
			case AddOnCardShowType_vote:
				textCard := new(Card_Display_AddOnCardInfo_TextVoteCard)
				if err := json.Unmarshal([]byte(addon.GetVoteCard()), textCard); err == nil {
					result = append(result, message.NewText("\n附加信息：\n选项：\n"))
					for _, opt := range textCard.GetOptions() {
						result = append(result, localutils.MessageTextf("%v - %v\n", opt.GetIdx(), opt.GetDesc()))
					}
				} else {
					log.WithField("content", addon.GetVoteCard()).Info("found new VoteCard")
				}
			case AddOnCardShowType_video:
				ugcCard := addon.GetUgcAttachCard()
				result = append(result, localutils.MessageTextf("\n附加视频：\n%v\n", ugcCard.GetTitle()))
				cover, err := localutils.UploadGroupImageByUrl(notify.GroupCode, ugcCard.GetImageUrl(), true, proxy_pool.PreferNone)
				if err != nil {
					log.WithField("pic", ugcCard.GetImageUrl()).Errorf("upload ugc cover failed %v", err)
					result = append(result, message.NewText("[封面]\n"))
				} else {
					result = append(result, cover)
				}
				result = append(result, localutils.MessageTextf("%v\n%v\n", ugcCard.GetDescSecond(), ugcCard.GetPlayUrl()))
			default:
				if b, err := json.Marshal(card.GetDisplay()); err != nil {
					log.WithField("content", card).Errorf("found new AddOnCardShowType but marshal failed %v", err)
				} else {
					log.WithField("content", string(b)).Info("found new AddOnCardShowType")
				}
			}
		}
	}
	log.WithField("dynamicUrl", dynamicUrl).Debug("create notify")
	result = append(result, message.NewText(dynamicUrl))
	notify.messageCache = result
	return result
}
func (notify *ConcernNewsNotify) GetGroupCode() int64 {
	return notify.GroupCode
}
func (notify *ConcernNewsNotify) GetUid() interface{} {
	return notify.Mid
}

func (notify *ConcernNewsNotify) Logger() *logrus.Entry {
	if notify == nil {
		return logger
	}
	return logger.WithFields(localutils.GroupLogFields(notify.GroupCode)).
		WithFields(logrus.Fields{
			"Site":      Site,
			"Mid":       notify.Mid,
			"Name":      notify.Name,
			"DynamicId": notify.Card.GetDesc().GetDynamicIdStr(),
			"DescType":  notify.Card.GetDesc().GetType().String(),
		})
}

func (notify *ConcernLiveNotify) ToMessage() []message.IMessageElement {
	log := notify.Logger()
	var result []message.IMessageElement
	switch notify.Status {
	case LiveStatus_Living:
		result = append(result, localutils.MessageTextf("%s正在直播【%v】\n%v", notify.Name, notify.LiveTitle, notify.RoomUrl))
	case LiveStatus_NoLiving:
		result = append(result, localutils.MessageTextf("%s直播结束了", notify.Name))
	}
	cover, err := localutils.UploadGroupImageByUrl(notify.GroupCode, notify.Cover, false, proxy_pool.PreferNone)
	if err != nil {
		log.WithField("Cover", notify.Cover).Errorf("add cover failed %v", err)
	} else {
		result = append(result, cover)
	}
	return result
}

func (notify *ConcernLiveNotify) Logger() *logrus.Entry {
	if notify == nil {
		return logger
	}
	return notify.LiveInfo.Logger().WithFields(localutils.GroupLogFields(notify.GroupCode))
}

func (notify *ConcernLiveNotify) GetGroupCode() int64 {
	return notify.GroupCode
}
func (notify *ConcernLiveNotify) GetUid() interface{} {
	return notify.Mid
}

// combineImageCache 是给combineImage用的cache，其他地方禁止使用
var combineImageCache = blockCache.NewBlockCache(8, 5)

var mode = "auto"
var modeSync sync.Once

func shouldCombineImage(pic []*CardWithImage_Item_Picture) bool {
	modeSync.Do(func() {
		if config.GlobalConfig == nil {
			return
		}
		switch config.GlobalConfig.GetString("bilibili.imageMergeMode") {
		case "auto":
			mode = "auto"
		case "off", "false":
			mode = "off"
		case "only9":
			mode = "only9"
		default:
			mode = "auto"
		}
	})
	if mode == "off" {
		return false
	} else if mode == "only9" {
		return len(pic) == 9
	}
	if len(pic) <= 3 {
		return false
	}
	if len(pic) == 9 {
		return true
	}
	// 有竖条形状的图
	for _, i := range pic {
		if i.ImgWidth > 250 && float64(i.ImgHeight) > 3*float64(i.ImgWidth) {
			return true
		}
	}
	// 有超过一半的近似矩形图片尺寸一样
	var size = make(map[int64]int)
	for _, i := range pic {
		var gap float64
		if i.ImgHeight < i.ImgWidth {
			gap = float64(i.ImgHeight) / float64(i.ImgWidth)
		} else {
			gap = float64(i.ImgWidth) / float64(i.ImgHeight)
		}
		if gap >= 0.95 {
			size[int64(i.ImgWidth)*int64(i.ImgHeight)] += 1
		}
	}
	var sizeMerge bool
	for _, count := range size {
		if 2*count > len(pic) {
			sizeMerge = true
		}
	}
	if sizeMerge && (len(pic) == 4 || len(pic) == 6 || len(pic) == 9) {
		return true
	}
	return false
}

func urlsMergeImage(urls []string) (result []byte, err error) {
	cacheR := combineImageCache.WithCacheDo(strings.Join(urls, "+"), func() blockCache.ActionResult {
		var imgBytes = make([][]byte, len(urls))
		for index, url := range urls {
			imgBytes[index], err = localutils.ImageGet(url, proxy_pool.PreferNone)
			if err != nil {
				return blockCache.NewResultWrapper(nil, err)
			}
		}
		return blockCache.NewResultWrapper(localutils.MergeImages(imgBytes))
	})
	if cacheR.Err() != nil {
		return nil, cacheR.Err()
	}
	return cacheR.Result().([]byte), nil
}
