package weibo

import (
	"github.com/Sora233/DDBOT/lsp/concern_type"
	"github.com/Sora233/DDBOT/lsp/mmsg"
	"github.com/sirupsen/logrus"
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
	GroupCode int64 `json:"group_code"`
	*UserInfo
	Card *Card
}

func (c *ConcernNewsNotify) Type() concern_type.Type {
	return News
}

func (c *ConcernNewsNotify) GetGroupCode() int64 {
	return c.GroupCode
}

func (c *ConcernNewsNotify) ToMessage() (m *mmsg.MSG) {
	m = mmsg.NewMSG()
	switch c.Card.GetCardType() {
	//case CardType_Normal:
	default:
		m.Textf("weibo-%v发布了新微博：\n%v\n%v\n",
			c.GetName(),
			c.Card.GetMblog().GetCreatedAt(),
			c.Card.GetSchema())
	}
	return
}

func NewConcernNewsNotify(groupCode int64, info *NewsInfo) []*ConcernNewsNotify {
	var result []*ConcernNewsNotify
	for _, card := range info.Cards {
		result = append(result, &ConcernNewsNotify{
			GroupCode: groupCode,
			UserInfo:  info.UserInfo,
			Card:      card,
		})
	}
	return result
}
