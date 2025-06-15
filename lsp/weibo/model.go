package weibo

import (
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/Sora233/DDBOT/v2/lsp/concern_type"
	"github.com/Sora233/DDBOT/v2/lsp/mmsg"
	localutils "github.com/Sora233/DDBOT/v2/utils"
)

const (
	News concern_type.Type = "news"
)

type UserInfo struct {
	Uid             int64  `json:"uid"`
	Name            string `json:"name"`
	ProfileImageUrl string `json:"profile_image_url"`
	ProfileUrl      string `json:"profile_url"`
}

func (u *UserInfo) Site() string {
	return Site
}

func (u *UserInfo) GetUid() interface{} {
	return u.Uid
}

func (u *UserInfo) GetName() string {
	return u.Name
}

func (u *UserInfo) Logger() *logrus.Entry {
	return logger.WithFields(logrus.Fields{
		"Site": Site,
		"Uid":  u.Uid,
		"Name": u.Name,
	})
}

type NewsInfo struct {
	*UserInfo
	LatestNewsTs int64   `json:"latest_news_time"`
	Cards        []*Card `json:"-"`
}

func (n *NewsInfo) Type() concern_type.Type {
	return News
}

func (n *NewsInfo) Logger() *logrus.Entry {
	return n.UserInfo.Logger().WithFields(logrus.Fields{
		"Type":     n.Type().String(),
		"CardSize": len(n.Cards),
	})
}

type ConcernNewsNotify struct {
	GroupCode uint32 `json:"group_code"`
	*UserInfo
	Card *CacheCard
}

func (c *ConcernNewsNotify) Type() concern_type.Type {
	return News
}

func (c *ConcernNewsNotify) GetGroupCode() uint32 {
	return c.GroupCode
}

func (c *ConcernNewsNotify) Logger() *logrus.Entry {
	return c.UserInfo.Logger().WithFields(localutils.GroupLogFields(c.GroupCode))
}

func (c *ConcernNewsNotify) ToMessage() (m *mmsg.MSG) {
	return c.Card.GetMSG()
}

func NewConcernNewsNotify(groupCode uint32, info *NewsInfo) []*ConcernNewsNotify {
	var result []*ConcernNewsNotify
	for _, card := range info.Cards {
		result = append(result, &ConcernNewsNotify{
			GroupCode: groupCode,
			UserInfo:  info.UserInfo,
			Card:      NewCacheCard(card, info.GetName()),
		})
	}
	return result
}

type CacheCard struct {
	*Card
	Name string

	once     sync.Once
	msgCache *mmsg.MSG
}

func NewCacheCard(card *Card, name string) *CacheCard {
	return &CacheCard{Card: card, Name: name}
}

func (c *CacheCard) prepare() {
	m := mmsg.NewMSG()
	var createdTime string
	newsTime, err := time.Parse(time.RubyDate, c.Card.GetMblog().GetCreatedAt())
	if err == nil {
		createdTime = newsTime.Format("2006-01-02 15:04:05")
	} else {
		createdTime = c.Card.GetMblog().GetCreatedAt()
	}
	if c.Card.GetMblog().GetRetweetedStatus() != nil {
		m.Textf("weibo-%v转发了%v的微博：\n%v",
			c.Name,
			c.Card.GetMblog().GetRetweetedStatus().GetUser().GetScreenName(),
			createdTime,
		)
	} else {
		m.Textf("weibo-%v发布了新微博：\n%v",
			c.Name,
			createdTime,
		)
	}
	switch c.Card.GetCardType() {
	case CardType_Normal:
		if len(c.Card.GetMblog().GetRawText()) > 0 {
			m.Textf("\n%v", localutils.RemoveHtmlTag(c.Card.GetMblog().GetRawText()))
		} else {
			m.Textf("\n%v", localutils.RemoveHtmlTag(c.Card.GetMblog().GetText()))
		}
		for _, pic := range c.Card.GetMblog().GetPics() {
			m.ImageByUrl(pic.GetLarge().GetUrl(), "")
		}
	default:
		logger.WithField("Type", c.CardType.String()).Debug("found new card_types")
	}
	if idx := strings.Index(c.Card.GetScheme(), "?"); idx > 0 {
		m.Textf("\n%v", c.Card.GetScheme()[:idx])
	} else {
		m.Textf("\n%v", c.Card.GetScheme())
	}
	c.msgCache = m
}

func (c *CacheCard) GetMSG() *mmsg.MSG {
	c.once.Do(c.prepare)
	return c.msgCache
}
