package bilibili

import (
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Sora233/DDBOT/lsp/concern_type"
	"github.com/Sora233/DDBOT/lsp/mmsg"
	"github.com/Sora233/DDBOT/lsp/template"
	localutils "github.com/Sora233/DDBOT/utils"
	"github.com/Sora233/DDBOT/utils/blockCache"
	"github.com/Sora233/MiraiGo-Template/config"
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

func (n *NewsInfo) Site() string {
	return Site
}

func (n *NewsInfo) Type() concern_type.Type {
	return News
}

func (n *NewsInfo) Logger() *logrus.Entry {
	return logger.WithFields(logrus.Fields{
		"Site":     Site,
		"Mid":      n.Mid,
		"Name":     n.Name,
		"CardSize": len(n.Cards),
		"Type":     n.Type().String(),
	})
}

type ConcernNewsNotify struct {
	GroupCode int64 `json:"group_code"`
	*UserInfo
	Card *CacheCard

	// 用于联合投稿和转发的时候防止多人同时推送
	shouldCompact bool
	compactKey    string
	concern       *Concern
}

func (notify *ConcernNewsNotify) IsLive() bool {
	return false
}

func (notify *ConcernNewsNotify) Living() bool {
	return false
}

type ConcernLiveNotify struct {
	GroupCode int64 `json:"group_code"`
	*LiveInfo
}

type UserStat struct {
	Mid int64 `json:"mid"`
	// 关注数
	Following int64 `json:"following"`
	// 粉丝数
	Follower int64 `json:"follower"`
}

type UserInfo struct {
	Mid     int64  `json:"mid"`
	Name    string `json:"name"`
	RoomId  int64  `json:"room_id"`
	RoomUrl string `json:"room_url"`

	UserStat *UserStat `json:"-"`
}

func (ui *UserInfo) GetUid() interface{} {
	return ui.Mid
}

func (ui *UserInfo) GetName() string {
	if ui == nil {
		return ""
	}
	return ui.Name
}

type LiveInfo struct {
	UserInfo
	Status    LiveStatus `json:"status"`
	LiveTitle string     `json:"live_title"`
	Cover     string     `json:"cover"`

	once              sync.Once
	msgCache          *mmsg.MSG
	liveStatusChanged bool
	liveTitleChanged  bool
}

func (l *LiveInfo) GetMSG() *mmsg.MSG {
	if l == nil {
		return nil
	}
	// 现在直播url会带一个`?broadcast_type=0`，好像删掉也行
	cleanRoomUrl := func(url string) string {
		if pos := strings.Index(url, "?"); pos > 0 {
			return url[:pos]
		}
		return url
	}
	l.once.Do(func() {
		var data = map[string]interface{}{
			"title":  l.LiveTitle,
			"name":   l.Name,
			"url":    cleanRoomUrl(l.RoomUrl),
			"cover":  l.Cover,
			"living": l.Living(),
		}
		var err error
		l.msgCache, err = template.LoadAndExec("notify.group.bilibili.live.tmpl", data)
		if err != nil {
			logger.Errorf("bilibili: LiveInfo LoadAndExec error %v", err)
		}
		return
	})
	return l.msgCache
}

func (l *LiveInfo) TitleChanged() bool {
	return l.liveTitleChanged
}

func (l *LiveInfo) LiveStatusChanged() bool {
	return l.liveStatusChanged
}

func (l *LiveInfo) IsLive() bool {
	return true
}

func (l *LiveInfo) Site() string {
	return Site
}

func (l *LiveInfo) Living() bool {
	if l == nil {
		return false
	}
	return l.Status == LiveStatus_Living
}

func (l *LiveInfo) Type() concern_type.Type {
	return Live
}

func (l *LiveInfo) Logger() *logrus.Entry {
	return logger.WithFields(logrus.Fields{
		"Site":   Site,
		"Mid":    l.Mid,
		"Name":   l.Name,
		"RoomId": l.RoomId,
		"Title":  l.LiveTitle,
		"Status": l.Status.String(),
		"Type":   l.Type().String(),
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
			Card:      NewCacheCard(card),
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
		LiveInfo:  liveInfo,
	}
}

func (notify *ConcernNewsNotify) ToMessage() (m *mmsg.MSG) {
	var (
		card       = notify.Card
		log        = notify.Logger()
		dynamicUrl = DynamicUrl(card.GetDesc().GetDynamicIdStr())
		date       = localutils.TimestampFormat(card.GetDesc().GetTimestamp())
	)
	// 推送一条简化动态防止刷屏，主要是联合投稿和转发的时候
	if notify.shouldCompact {
		// 通过回复之前消息的方式简化推送
		m = mmsg.NewMSG()
		msg, _ := notify.concern.GetNotifyMsg(notify.GroupCode, notify.compactKey)
		if msg != nil {
			m.Append(message.NewReply(msg))
		}
		log.WithField("compact_key", notify.compactKey).Debug("compact notify")
		switch notify.Card.GetDesc().GetType() {
		case DynamicDescType_WithVideo:
			videoCard, _ := notify.Card.GetCardWithVideo()
			m.Textf("%v%v：\n%v\n%v\n%v",
				notify.Name,
				notify.Card.GetDisplay().GetUsrActionTxt(),
				date,
				videoCard.GetTitle(),
				dynamicUrl)
			return
		case DynamicDescType_WithOrigin:
			origCard, _ := notify.Card.GetCardWithOrig()
			m.Textf("%v转发了%v的动态：\n%v\n%v\n%v",
				notify.Name,
				origCard.GetOriginUser().GetInfo().GetUname(),
				date,
				origCard.GetItem().GetContent(),
				dynamicUrl)
			return
		}
	}
	m = notify.Card.GetMSG()
	return
}

func (notify *ConcernNewsNotify) Type() concern_type.Type {
	return News
}

func (notify *ConcernNewsNotify) Site() string {
	return Site
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
			"Type":      notify.Type().String(),
		})
}

func (notify *ConcernLiveNotify) ToMessage() (m *mmsg.MSG) {
	return notify.LiveInfo.GetMSG()
}

func (notify *ConcernLiveNotify) Logger() *logrus.Entry {
	if notify == nil {
		return logger
	}
	return notify.LiveInfo.Logger().
		WithFields(localutils.GroupLogFields(notify.GroupCode))
}

func (notify *ConcernLiveNotify) GetGroupCode() int64 {
	return notify.GroupCode
}

// combineImageCache 是给combineImage用的cache，其他地方禁止使用
var combineImageCache = blockCache.NewBlockCache(5, 3)

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
			imgBytes[index], err = localutils.ImageGet(url)
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

type CacheCard struct {
	*Card
	once     sync.Once
	msgCache *mmsg.MSG
}

func NewCacheCard(card *Card) *CacheCard {
	cacheCard := new(CacheCard)
	cacheCard.Card = card
	return cacheCard
}

func (c *CacheCard) prepare() {
	var (
		card       = c.Card
		log        = logger
		dynamicUrl = DynamicUrl(card.GetDesc().GetDynamicIdStr())
		date       = localutils.TimestampFormat(card.GetDesc().GetTimestamp())
		name       = c.Card.GetDesc().GetUserProfile().GetInfo().GetUname()
	)
	m := mmsg.NewMSG()
	defer func() {
		c.msgCache = m
	}()
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
			m.Textf("%v转发了%v的动态：\n%v\n%v\n\n原动态：\n",
				name, originName, date, cardOrigin.GetItem().GetContent())
			origin := new(CardWithImage)
			err := json.Unmarshal([]byte(cardOrigin.GetOrigin()), origin)
			if err != nil {
				log.WithField("origin", cardOrigin.GetOrigin()).
					Errorf("Unmarshal origin cardWithImage failed %v", err)
				return
			}
			m.Textf("%v\n", origin.GetItem().GetDescription())
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
					m.Image(resultByte, "")
					skip = true
				}
			}
			if !skip {
				for _, pic := range origin.GetItem().GetPictures() {
					var isNorm = false
					if pic.ImgHeight > 1200 && pic.ImgWidth > 1200 {
						isNorm = true
					}
					if isNorm {
						m.ImageByUrlWithNorm(pic.GetImgSrc(), "")
					} else {
						m.ImageByUrl(pic.GetImgSrc(), "")
					}
				}
			}
		case DynamicDescType_TextOnly:
			m.Textf("%v转发了%v的动态：\n%v\n%v\n\n原动态：\n", name, originName, date, cardOrigin.GetItem().GetContent())
			origin := new(CardTextOnly)
			err := json.Unmarshal([]byte(cardOrigin.GetOrigin()), origin)
			if err != nil {
				log.WithField("origin", cardOrigin.GetOrigin()).Errorf("Unmarshal origin cardWithText failed %v", err)
				return
			}
			m.Textf("%v\n", origin.GetItem().GetContent())
		case DynamicDescType_WithVideo:
			m.Textf("%v转发了%v的投稿：\n%v\n%v\n\n原视频：\n", name, originName, date, cardOrigin.GetItem().GetContent())
			origin := new(CardWithVideo)
			err := json.Unmarshal([]byte(cardOrigin.GetOrigin()), origin)
			if err != nil {
				log.WithField("origin", cardOrigin.GetOrigin()).Errorf("Unmarshal origin cardWithVideo failed %v", err)
				return
			}
			m.Textf("%v\n%v\n", origin.GetTitle(), origin.GetDesc())
			m.ImageByUrlWithNorm(origin.GetPic(), "[封面]")
		case DynamicDescType_WithPost:
			m.Textf("%v转发了%v的专栏：\n%v\n%v\n\n原专栏：\n", name, originName, date, cardOrigin.GetItem().GetContent())
			origin := new(CardWithPost)
			err := json.Unmarshal([]byte(cardOrigin.GetOrigin()), origin)
			if err != nil {
				log.WithField("origin", cardOrigin.GetOrigin()).Errorf("Unmarshal origin cardWithPost failed %v", err)
				return
			}
			m.Textf("%v\n%v\n", origin.GetTitle(), origin.GetSummary())
			if len(origin.GetImageUrls()) >= 1 {
				m.ImageByUrl(origin.GetImageUrls()[0], "")
			} else if len(origin.GetBannerUrl()) != 0 {
				m.ImageByUrl(origin.GetBannerUrl(), "")
			}
		case DynamicDescType_WithMusic:
			origin := new(CardWithMusic)
			err := json.Unmarshal([]byte(cardOrigin.GetOrigin()), origin)
			if err != nil {
				log.WithField("origin", cardOrigin.GetOrigin()).Errorf("Unmarshal origin CardWithMusic failed %v", err)
				return
			}
			m.Textf("%v转发了%v的音频：\n%v\n%v\n\n原音频：\n",
				name, originName, date, cardOrigin.GetItem().GetContent())
			m.Textf("%v\n%v\n", origin.GetTitle(), origin.GetIntro())
			m.ImageByUrl(origin.GetCover(), "")
		case DynamicDescType_WithSketch:
			origin := new(CardWithSketch)
			err := json.Unmarshal([]byte(cardOrigin.GetOrigin()), origin)
			if err != nil {
				log.WithField("origin", cardOrigin.GetOrigin()).Errorf("Unmarshal origin CardWithSketch failed %v", err)
				return
			}
			m.Textf("%v转发了%v的动态：\n%v\n%v\n原动态：\n%v\n%v\n%v", name, originName, date, cardOrigin.GetItem().GetContent(),
				origin.GetVest().GetContent(), origin.GetSketch().GetTitle(), origin.GetSketch().GetDescText())
			if len(origin.GetSketch().GetCoverUrl()) != 0 {
				m.ImageByUrlWithNorm(origin.GetSketch().GetCoverUrl(), "")
			}
		case DynamicDescType_WithLive:
			m.Textf("%v分享了%v的直播：\n%v\n%v\n\n原直播间：\n", name, originName, date, cardOrigin.GetItem().GetContent())
			origin := new(CardWithLive)
			err := json.Unmarshal([]byte(cardOrigin.GetOrigin()), origin)
			if err != nil {
				log.WithField("origin", cardOrigin.GetOrigin()).Errorf("Unmarshal origin CardWithLive failed %v", err)
				return
			}
			m.Textf("%v\n", origin.GetTitle())
			m.ImageByUrl(origin.GetCover(), "[封面]")
		case DynamicDescType_WithLiveV2:
			m.Textf("%v分享了%v的直播：\n%v\n%v\n\n原直播间：\n", name, originName, date, cardOrigin.GetItem().GetContent())
			origin := new(CardWithLiveV2)
			err := json.Unmarshal([]byte(cardOrigin.GetOrigin()), origin)
			if err != nil {
				log.WithField("origin", cardOrigin.GetOrigin()).Errorf("Unmarshal origin CardWithLiveV2 failed %v", err)
				return
			}
			m.Textf("%v\n", origin.GetLivePlayInfo().GetTitle())
			m.ImageByUrl(origin.GetLivePlayInfo().GetCover(), "[封面]")
		case DynamicDescType_WithMylist:
			m.Textf("%v分享了%v的收藏夹：\n%v\n%v\n\n原收藏夹：\n", name, originName, date, cardOrigin.GetItem().GetContent())
			origin := new(CardWithMylist)
			err := json.Unmarshal([]byte(cardOrigin.GetOrigin()), origin)
			if err != nil {
				log.WithField("origin", cardOrigin.GetOrigin()).Errorf("Unmarshal origin CardWithMylist failed %v", err)
				return
			}
			m.Textf("%v\n", origin.GetTitle())
			m.ImageByUrl(origin.GetCover(), "")
		case DynamicDescType_WithMiss:
			m.Textf("%v分享了动态：\n%v\n%v\n\n%v\n", name, date, cardOrigin.GetItem().GetContent(), cardOrigin.GetItem().GetTips())
		case DynamicDescType_WithOrigin:
			// 麻了，套起来了
			m.Textf("%v转发了%v的动态：%v\n%v\n", name, originName, date, cardOrigin.GetItem().GetContent())
		case DynamicDescType_WithCourse:
			origin := new(CardWithCourse)
			err := json.Unmarshal([]byte(cardOrigin.GetOrigin()), origin)
			if err != nil {
				log.WithField("origin", cardOrigin.GetOrigin()).Errorf("Unmarshal origin CardWithCourse failed %v", err)
				return
			}
			m.Textf("%v转发了%v的%v：\n%v\n%v\n\n原课程：\n%v", name,
				origin.GetUpInfo().GetName(),
				origin.GetBadge().GetText(),
				date,
				cardOrigin.GetItem().GetContent(),
				origin.GetTitle())
			m.ImageByUrl(origin.GetCover(), "")
		default:
			// 试试media
			origin := new(CardWithMedia)
			err := json.Unmarshal([]byte(cardOrigin.GetOrigin()), origin)
			if err == nil && origin.GetApiSeasonInfo() != nil {
				var desc = origin.GetNewDesc()
				if len(desc) == 0 {
					desc = origin.GetIndex()
				}
				m.Textf("%v转发了%v【%v】%v：\n%v\n%v\n", name,
					origin.GetApiSeasonInfo().GetTypeName(),
					origin.GetApiSeasonInfo().GetTitle(),
					desc,
					date,
					cardOrigin.GetItem().GetContent())
				m.ImageByUrlWithNorm(origin.GetCover(), "[封面]")
			} else {
				log.WithField("content", card.GetCard()).Info("found new type with origin")
				m.Textf("%v转发了%v的动态：\n%v\n%v\n", name, originName, date, cardOrigin.GetItem().GetContent())
			}
		}
	case DynamicDescType_WithImage:
		cardImage, err := card.GetCardWithImage()
		if err != nil {
			log.WithField("card", card).Errorf("GetCardWithImage cast failed %v", err)
			return
		}
		m.Textf("%v发布了新动态：\n%v\n%v\n", name, date, cardImage.GetItem().GetDescription())
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
				m.Image(resultByte, "")
				skip = true
			}
		}
		if !skip {
			for _, pic := range cardImage.GetItem().GetPictures() {
				var isNorm = false
				if pic.ImgHeight > 1200 && pic.ImgWidth > 1200 {
					isNorm = true
				}
				if isNorm {
					m.ImageByUrlWithNorm(pic.GetImgSrc(), "")
				} else {
					m.ImageByUrl(pic.GetImgSrc(), "")
				}
			}
		}

	case DynamicDescType_TextOnly:
		cardText, err := card.GetCardTextOnly()
		if err != nil {
			log.WithField("card", card).Errorf("GetCardTextOnly cast failed %v", err)
			return
		}
		m.Textf("%v发布了新动态：\n%v\n%v\n", name, date, cardText.GetItem().GetContent())
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
		if description == cardVideo.GetTitle() {
			description = ""
		}
		// web接口好像还区分不了动态视频，先不处理了
		actionText := card.GetDisplay().GetUsrActionTxt()
		m.Textf("%v%v：\n%v\n%v\n", name, actionText, date, cardVideo.GetTitle())
		if len(description) != 0 {
			m.Textf("%v\n", description)
		}
		m.ImageByUrlWithNorm(cardVideo.GetPic(), "[封面]")
	case DynamicDescType_WithPost:
		cardPost, err := card.GetCardWithPost()
		if err != nil {
			log.WithField("card", card).Errorf("GetCardWithPost cast failed %v", err)
			return
		}
		m.Textf("%v发布了新专栏：\n%v\n%v\n%v...\n", name, date, cardPost.Title, cardPost.Summary)
		if len(cardPost.GetImageUrls()) >= 1 {
			m.ImageByUrl(cardPost.GetImageUrls()[0], "")
		} else if len(cardPost.GetBannerUrl()) != 0 {
			m.ImageByUrl(cardPost.GetBannerUrl(), "")
		}
	case DynamicDescType_WithMusic:
		cardMusic, err := card.GetCardWithMusic()
		if err != nil {
			log.WithField("card", card).
				Errorf("GetCardWithMusic cast failed %v", err)
			return
		}
		m.Textf("%v投稿了新音频：\n%v\n%v\n%v\n", name, date, cardMusic.GetTitle(), cardMusic.GetIntro())
		m.ImageByUrl(cardMusic.GetCover(), "[封面]")
	case DynamicDescType_WithSketch:
		cardSketch, err := card.GetCardWithSketch()
		if err != nil {
			log.WithField("card", card).
				Errorf("GetCardWithSketch cast failed %v", err)
			return
		}
		m.Textf("%v发表了新动态：\n%v\n%v\n", name, date, cardSketch.GetVest().GetContent())
		if cardSketch.GetSketch().GetTitle() == cardSketch.GetSketch().GetDescText() {
			m.Textf("内容：%v", cardSketch.GetSketch().GetTitle())
		} else {
			m.Textf("内容：%v - %v", cardSketch.GetSketch().GetTitle(), cardSketch.GetSketch().GetDescText())
		}
		if len(cardSketch.GetSketch().GetCoverUrl()) > 0 {
			m.ImageByUrlWithNorm(cardSketch.GetSketch().GetCoverUrl(), "")
		}
	case DynamicDescType_WithLive:
		cardLive, err := card.GetCardWithLive()
		if err != nil {
			log.WithField("card", card).
				Errorf("GetCardWithLive cast failed %v", err)
			return
		}
		m.Textf("%v发布了直播信息：\n%v\n%v\n", name, date, cardLive.GetTitle())
		m.ImageByUrlWithNorm(cardLive.GetCover(), "[封面]")
	case DynamicDescType_WithLiveV2:
		// 2021-08-15 发现这个是系统推荐的直播间，应该不是人为操作，选择不推送，在filter中过滤
		cardLiveV2, err := card.GetCardWithLiveV2()
		if err != nil {
			log.WithField("card", card).
				Errorf("GetCardWithLiveV2 case failed %v", err)
			return
		}
		m.Textf("%v发布了直播信息：\n%v\n%v\n", name, date, cardLiveV2.GetLivePlayInfo().GetTitle())
		// LiveV2 会被过滤，图片就不占用带宽了
		// m.ImageByUrlWithNorm(cardLiveV2.GetLivePlayInfo().GetCover(), "")
	case DynamicDescType_WithMiss:
		cardWithMiss, err := card.GetCardWithOrig()
		if err != nil {
			log.WithField("card", card).
				Errorf("GetCardWithOrig case failed %v", err)
			return
		}
		m.Textf("%v发布了新动态：\n%v\n%v\n\n%v\n", name, date, cardWithMiss.GetItem().GetContent(), cardWithMiss.GetItem().GetTips())
	default:
		log.WithField("content", card.GetCard()).Info("found new DynamicDescType")
		m.Textf("%v发布了新动态：\n%v\n", name, date)
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
				m.Textf("\n%v：\n%v\n", item.AdMark, item.Name)
				m.ImageByUrlWithNorm(item.GetImg(), "")
			case AddOnCardShowType_reserve:
				if len(addon.GetReserveAttachCard().GetReserveLottery().GetText()) == 0 {
					m.Textf("\n附加信息：\n%v\n%v\n",
						addon.GetReserveAttachCard().GetTitle(),
						addon.GetReserveAttachCard().GetDescFirst().GetText())
				} else {
					m.Textf("\n附加信息：\n%v\n%v\n%v\n",
						addon.GetReserveAttachCard().GetTitle(),
						addon.GetReserveAttachCard().GetDescFirst().GetText(),
						addon.GetReserveAttachCard().GetReserveLottery().GetText())
				}
			case AddOnCardShowType_match:
			// TODO 暂时没必要
			case AddOnCardShowType_related:
				aCard := addon.GetAttachCard()
				// 游戏应该不需要
				if aCard.GetType() != "game" {
					m.Textf("\n%v：\n%v\n%v\n",
						aCard.GetHeadText(),
						aCard.GetTitle(),
						aCard.GetDescFirst())
				}
			case AddOnCardShowType_vote:
				textCard := new(Card_Display_AddOnCardInfo_TextVoteCard)
				if err := json.Unmarshal([]byte(addon.GetVoteCard()), textCard); err == nil {
					m.Textf("\n附加信息：\n选项：\n")
					for _, opt := range textCard.GetOptions() {
						m.Textf("%v - %v\n", opt.GetIdx(), opt.GetDesc())
					}
				} else {
					log.WithField("content", addon.GetVoteCard()).Info("found new VoteCard")
				}
			case AddOnCardShowType_video:
				ugcCard := addon.GetUgcAttachCard()
				m.Textf("\n附加视频：\n%v\n", ugcCard.GetTitle())
				m.ImageByUrlWithNorm(ugcCard.GetImageUrl(), "[封面]")
				m.Textf("%v\n%v\n", ugcCard.GetDescSecond(), ugcCard.GetPlayUrl())
			default:
				if b, err := json.Marshal(card.GetDisplay()); err != nil {
					log.WithField("content", card).Errorf("found new AddOnCardShowType but marshal failed %v", err)
				} else {
					log.WithField("content", string(b)).Info("found new AddOnCardShowType")
				}
			}
		}
	}
	m.Text(dynamicUrl)
}

func (c *CacheCard) GetMSG() *mmsg.MSG {
	c.once.Do(c.prepare)
	return c.msgCache
}
