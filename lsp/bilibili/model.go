package bilibili

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Sora233/DDBOT/concern"
	"github.com/Sora233/DDBOT/proxy_pool"
	localutils "github.com/Sora233/DDBOT/utils"
	lru "github.com/hashicorp/golang-lru"
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

func (n *NewsInfo) GetCardWithImage(index int) (*CardWithImage, error) {
	if len(n.Cards) <= index || n.Cards[index].GetCard() == "" {
		return nil, errors.New("card not found or empty")
	}
	if n.Cards[index].GetDesc().GetType() == DynamicDescType_WithImage {
		var card = new(CardWithImage)
		err := json.Unmarshal([]byte(n.Cards[index].GetCard()), card)
		return card, err
	}
	return nil, errors.New("type mismatch")
}

func (n *NewsInfo) GetCardWithOrig(index int) (*CardWithOrig, error) {
	if len(n.Cards) <= index || n.Cards[index].GetCard() == "" {
		return nil, errors.New("card not found or empty")
	}
	if n.Cards[index].GetDesc().GetType() == DynamicDescType_WithOrigin {
		var card = new(CardWithOrig)
		err := json.Unmarshal([]byte(n.Cards[index].GetCard()), card)
		return card, err
	}
	return nil, errors.New("type mismatch")
}

func (n *NewsInfo) GetCardWithVideo(index int) (*CardWithVideo, error) {
	if len(n.Cards) <= index || n.Cards[index].GetCard() == "" {
		return nil, errors.New("card not found or empty")
	}
	if n.Cards[index].GetDesc().GetType() == DynamicDescType_WithVideo {
		var card = new(CardWithVideo)
		err := json.Unmarshal([]byte(n.Cards[index].GetCard()), card)
		return card, err
	}
	return nil, errors.New("type mismatch")
}

func (n *NewsInfo) GetCardTextOnly(index int) (*CardTextOnly, error) {
	if len(n.Cards) <= index || n.Cards[index].GetCard() == "" {
		return nil, errors.New("card not found or empty")
	}
	if n.Cards[index].GetDesc().GetType() == DynamicDescType_TextOnly {
		var card = new(CardTextOnly)
		err := json.Unmarshal([]byte(n.Cards[index].GetCard()), card)
		return card, err
	}
	return nil, errors.New("type mismatch")
}

func (n *NewsInfo) GetCardWithPost(index int) (*CardWithPost, error) {
	if len(n.Cards) <= index || n.Cards[index].GetCard() == "" {
		return nil, errors.New("card not found or empty")
	}
	if n.Cards[index].GetDesc().GetType() == DynamicDescType_WithPost {
		var card = new(CardWithPost)
		err := json.Unmarshal([]byte(n.Cards[index].GetCard()), card)
		return card, err
	}
	return nil, errors.New("type mismatch")
}

func (n *NewsInfo) GetCardWithMusic(index int) (*CardWithMusic, error) {
	if len(n.Cards) <= index || n.Cards[index].GetCard() == "" {
		return nil, errors.New("card not found or empty")
	}
	if n.Cards[index].GetDesc().GetType() == DynamicDescType_WithMusic {
		var card = new(CardWithMusic)
		err := json.Unmarshal([]byte(n.Cards[index].GetCard()), card)
		return card, err
	}
	return nil, errors.New("type mismatch")

}

func (n *NewsInfo) GetCardWithSketch(index int) (*CardWithSketch, error) {
	if len(n.Cards) <= index || n.Cards[index].GetCard() == "" {
		return nil, errors.New("card not found or empty")
	}
	if n.Cards[index].GetDesc().GetType() == DynamicDescType_WithSketch {
		var card = new(CardWithSketch)
		err := json.Unmarshal([]byte(n.Cards[index].GetCard()), card)
		return card, err
	}
	return nil, errors.New("type mismatch")
}

func (n *NewsInfo) GetCardWithLive(index int) (*CardWithLive, error) {
	if len(n.Cards) <= index || n.Cards[index].GetCard() == "" {
		return nil, errors.New("card not found or empty")
	}
	if n.Cards[index].GetDesc().GetType() == DynamicDescType_WithLive {
		var card = new(CardWithLive)
		err := json.Unmarshal([]byte(n.Cards[index].GetCard()), card)
		return card, err
	}
	return nil, errors.New("type mismatch")
}

func (n *NewsInfo) GetCardWithLiveV2(index int) (*CardWithLiveV2, error) {
	if len(n.Cards) <= index || n.Cards[index].GetCard() == "" {
		return nil, errors.New("card not found or empty")
	}
	if n.Cards[index].GetDesc().GetType() == DynamicDescType_WithLiveV2 {
		var card = new(CardWithLiveV2)
		err := json.Unmarshal([]byte(n.Cards[index].GetCard()), card)
		return card, err
	}
	return nil, errors.New("type mismatch")
}

func (n *NewsInfo) ToString() string {
	if n == nil {
		return ""
	}
	content, _ := json.Marshal(n)
	return string(content)
}

type ConcernNewsNotify struct {
	GroupCode int64 `json:"group_code"`
	NewsInfo

	// messageCache 导致ConcernNewsNotify的ToMessage()变得线程不安全
	messageCache []message.IMessageElement
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

type UserInfo struct {
	Mid     int64  `json:"mid"`
	Name    string `json:"name"`
	RoomId  int64  `json:"room_id"`
	RoomUrl string `json:"room_url"`
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

func NewConcernNewsNotify(groupCode int64, newsInfo *NewsInfo) *ConcernNewsNotify {
	if newsInfo == nil {
		return nil
	}
	return &ConcernNewsNotify{
		GroupCode: groupCode,
		NewsInfo:  *newsInfo,
	}
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

func (notify *ConcernNewsNotify) ToMessage() []message.IMessageElement {
	if notify.messageCache != nil {
		return notify.messageCache
	}
	var results []message.IMessageElement
	for index, card := range notify.Cards {
		var result []message.IMessageElement
		log := logger.WithField("DescType", card.GetDesc().GetType().String())
		dynamicUrl := DynamicUrl(card.GetDesc().GetDynamicIdStr())
		date := localutils.TimestampFormat(card.GetDesc().GetTimestamp())
		switch card.GetDesc().GetType() {
		case DynamicDescType_WithOrigin:
			cardOrigin, err := notify.GetCardWithOrig(index)
			if err != nil {
				log.WithField("name", notify.Name).WithField("card", card).Errorf("cast failed %v", err)
				continue
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
					continue
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
						if pic.GetImgHeight() > 1200 && pic.GetImgWidth() > 1200 {
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
					continue
				}
				result = append(result, localutils.MessageTextf("%v\n", origin.GetItem().GetContent()))
			case DynamicDescType_WithVideo:
				result = append(result, localutils.MessageTextf("%v转发了%v的投稿：\n%v\n%v\n\n原视频：\n", notify.Name, originName, date, cardOrigin.GetItem().GetContent()))
				origin := new(CardWithVideo)
				err := json.Unmarshal([]byte(cardOrigin.GetOrigin()), origin)
				if err != nil {
					log.WithField("origin", cardOrigin.GetOrigin()).Errorf("Unmarshal origin cardWithVideo failed %v", err)
					continue
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
					continue
				}
				result = append(result, localutils.MessageTextf("%v\n%v\n", origin.GetTitle(), origin.GetSummary()))
				var cover *message.GroupImageElement
				if len(origin.GetImageUrls()) >= 1 {
					cover, err = localutils.UploadGroupImageByUrl(notify.GroupCode, origin.GetImageUrls()[0], false, proxy_pool.PreferNone)
				} else {
					cover, err = localutils.UploadGroupImageByUrl(notify.GroupCode, origin.GetBannerUrl(), false, proxy_pool.PreferNone)
				}
				if err != nil {
					log.WithField("image_url", origin.GetImageUrls()).
						WithField("banner_url", origin.GetBannerUrl()).
						Errorf("upload image failed %v", err)
				} else {
					result = append(result, cover)
				}
			case DynamicDescType_WithMusic:
				origin := new(CardWithMusic)
				err := json.Unmarshal([]byte(cardOrigin.GetOrigin()), origin)
				if err != nil {
					log.WithField("origin", cardOrigin.GetOrigin()).Errorf("Unmarshal origin CardWithMusic failed %v", err)
					continue
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
			case DynamicDescType_WithAnime:
				origin := new(CardWithAnime)
				err := json.Unmarshal([]byte(cardOrigin.GetOrigin()), origin)
				if err != nil {
					log.WithField("origin", cardOrigin.GetOrigin()).Errorf("Unmarshal origin CardWithAnime failed %v", err)
					continue
				}
				result = append(result, localutils.MessageTextf("%v转发了%v【%v】：\n%v\n%v\n", notify.Name,
					origin.GetApiSeasonInfo().GetTypeName(),
					origin.GetApiSeasonInfo().GetTitle(),
					date,
					cardOrigin.GetItem().GetContent()),
				)
			case DynamicDescType_WithSketch:
				origin := new(CardWithSketch)
				err := json.Unmarshal([]byte(cardOrigin.GetOrigin()), origin)
				if err != nil {
					log.WithField("origin", cardOrigin.GetOrigin()).Errorf("Unmarshal origin CardWithSketch failed %v", err)
					continue
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
					continue
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
					continue
				}
				result = append(result, localutils.MessageTextf("%v\n", origin.GetLivePlayInfo().GetTitle()))
				groupImage, err := localutils.UploadGroupImageByUrl(notify.GroupCode, origin.GetLivePlayInfo().GetCover(), false, proxy_pool.PreferNone)
				if err != nil {
					log.WithField("origin", cardOrigin.GetOrigin()).Errorf("upload liveV2 cover failed %v", err)
				} else {
					result = append(result, groupImage)
				}
			case DynamicDescType_WithMiss:
				result = append(result, localutils.MessageTextf("%v分享了动态：\n%v\n%v\n\n%v\n", notify.Name, date, cardOrigin.GetItem().GetContent(), cardOrigin.GetItem().GetTips()))
			default:
				log.WithField("content", card.GetCard()).Info("found new type with origin")
				result = append(result, localutils.MessageTextf("%v转发了%v的动态：\n%v\n%v\n", notify.Name, originName, date, cardOrigin.GetItem().GetContent()))
			}
		case DynamicDescType_WithImage:
			cardImage, err := notify.GetCardWithImage(index)
			if err != nil {
				log.WithField("name", notify.Name).WithField("card", card).Errorf("cast failed %v", err)
				continue
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
			cardText, err := notify.GetCardTextOnly(index)
			if err != nil {
				log.WithField("name", notify.Name).WithField("card", card).Errorf("cast failed %v", err)
				continue
			}
			result = append(result, localutils.MessageTextf("%v发布了新动态：\n%v\n%v\n", notify.Name, date, cardText.GetItem().GetContent()))
		case DynamicDescType_WithVideo:
			cardVideo, err := notify.GetCardWithVideo(index)
			if err != nil {
				log.WithField("name", notify.Name).WithField("card", card).Errorf("cast failed %v", err)
				continue
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
			cardPost, err := notify.GetCardWithPost(index)
			if err != nil {
				log.WithField("name", notify.Name).WithField("card", card).Errorf("cast failed %v", err)
				continue
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
			cardMusic, err := notify.GetCardWithMusic(index)
			if err != nil {
				log.WithField("name", notify.Name).
					WithField("card", card).
					Errorf("cast failed %v", err)
				continue
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
			cardSketch, err := notify.GetCardWithSketch(index)
			if err != nil {
				log.WithField("name", notify.Name).
					WithField("card", card).
					Errorf("cast failed %v", err)
				continue
			}
			result = append(result, localutils.MessageTextf(
				"%v发表了新动态：\n%v\n%v\n内容：%v - %v",
				notify.Name,
				date,
				cardSketch.GetVest().GetContent(),
				cardSketch.GetSketch().GetTitle(),
				cardSketch.GetSketch().GetDescText(),
			))
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
			cardLive, err := notify.GetCardWithLive(index)
			if err != nil {
				log.WithField("name", notify.Name).
					WithField("card", card).
					Errorf("cast failed %v", err)
				continue
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
			cardLiveV2, err := notify.GetCardWithLiveV2(index)
			if err != nil {
				log.WithField("name", notify.Name).
					WithField("card", card).
					Errorf("case failed %v", err)
				continue
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
			cardWithMiss, err := notify.GetCardWithOrig(index)
			if err != nil {
				log.WithField("name", notify.Name).
					WithField("card", card).
					Errorf("case failed %v", err)
				continue
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
		log.WithField("uid", notify.Mid).WithField("name", notify.Name).WithField("dynamicUrl", dynamicUrl).Debug("append")
		result = append(result, message.NewText(dynamicUrl+"\n"))
		results = append(results, result...)
	}
	notify.messageCache = results
	return results
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
	return logger.WithField("site", Site).
		WithFields(localutils.GroupLogFields(notify.GroupCode)).
		WithField("Uid", notify.Mid).
		WithField("Name", notify.Name).
		WithField("NewsCount", len(notify.Cards))
}

func (notify *ConcernLiveNotify) ToMessage() []message.IMessageElement {
	var result []message.IMessageElement
	switch notify.Status {
	case LiveStatus_Living:
		result = append(result, localutils.MessageTextf("%s正在直播【%v】\n%v", notify.Name, notify.LiveTitle, notify.RoomUrl))
	case LiveStatus_NoLiving:
		result = append(result, localutils.MessageTextf("%s直播结束了\n%v", notify.Name, notify.RoomUrl))
	}
	cover, err := localutils.UploadGroupImageByUrl(notify.GroupCode, notify.Cover, false, proxy_pool.PreferNone)
	if err != nil {
		logger.WithFields(localutils.GroupLogFields(notify.GroupCode)).
			WithField("cover", notify.Cover).
			Errorf("add cover failed %v", err)
	} else {
		result = append(result, cover)
	}
	return result
}

func (notify *ConcernLiveNotify) Logger() *logrus.Entry {
	if notify == nil {
		return logger
	}
	return logger.WithField("site", Site).
		WithFields(localutils.GroupLogFields(notify.GroupCode)).
		WithField("Uid", notify.Mid).
		WithField("Name", notify.Name).
		WithField("Title", notify.LiveTitle).
		WithField("Status", notify.Status.String())
}

func (notify *ConcernLiveNotify) GetGroupCode() int64 {
	return notify.GroupCode
}
func (notify *ConcernLiveNotify) GetUid() interface{} {
	return notify.Mid
}

const blockSize = 8

var mergeImageMutex = func() [blockSize]*sync.Mutex {
	var m = [blockSize]*sync.Mutex{}
	for i := 0; i < blockSize; i++ {
		m[int32(i)] = new(sync.Mutex)
	}
	return m
}()

var mergeImageCache = func() *lru.ARCCache {
	c, err := lru.NewARC(5)
	if err != nil {
		panic(err)
	}
	return c
}()

func shouldCombineImage(pic []*CardWithImage_Item_Picture) bool {
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
	// 所有图尺寸相同
	var sameSize = true
	for idx, i := range pic {
		if idx == 0 {
			continue
		}
		if i.ImgWidth == pic[0].ImgWidth && i.ImgHeight == pic[0].ImgHeight {
			continue
		}
		sameSize = false
	}
	if sameSize && (len(pic) == 4 || len(pic) == 6 || len(pic) == 9) {
		return true
	}
	return false
}

func urlsMergeImage(urls []string) (result []byte, err error) {
	hashFn := func(s []string) string {
		hash := md5.New()
		for _, i := range s {
			hash.Write([]byte(i))
		}
		return hex.EncodeToString(hash.Sum(nil))
	}
	cacheKey := hashFn(urls)
	if v, ok := mergeImageCache.Get(cacheKey); ok {
		switch v.(type) {
		case error:
			return nil, v.(error)
		case []byte:
			return v.([]byte), nil
		default:
			return nil, errors.New("unknown mergeImage cache")
		}
	}
	cacheIndex := int32(cacheKey[0]) % blockSize
	mergeImageMutex[cacheIndex].Lock()
	defer mergeImageMutex[cacheIndex].Unlock()

	if v, ok := mergeImageCache.Get(cacheKey); ok {
		switch v.(type) {
		case error:
			return nil, v.(error)
		case []byte:
			return v.([]byte), nil
		default:
			return nil, errors.New("unknown mergeImage cache")
		}
	}

	defer func() {
		if err != nil {
			mergeImageCache.Add(cacheKey, err)
		} else {
			mergeImageCache.Add(cacheKey, result)
		}
	}()

	var imgBytes = make([][]byte, len(urls))
	for index, url := range urls {
		imgBytes[index], err = localutils.ImageGet(url, proxy_pool.PreferNone)
		if err != nil {
			return
		}
	}
	result, err = localutils.MergeImages(imgBytes)
	return
}
