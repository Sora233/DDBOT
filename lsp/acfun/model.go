package acfun

import (
	"github.com/Sora233/DDBOT/lsp/concern_type"
	"github.com/Sora233/DDBOT/lsp/mmsg"
	"github.com/Sora233/DDBOT/proxy_pool"
	localutils "github.com/Sora233/DDBOT/utils"
	"github.com/sirupsen/logrus"
)

type UserInfo struct {
	Uid      int64  `json:"uid"`
	Name     string `json:"name"`
	Followed int    `json:"followed"`
	UserImg  string `json:"user_img"`
	LiveUrl  string `json:"live_url"`
}

func (u *UserInfo) GetUid() interface{} {
	return u.Uid
}

func (u *UserInfo) GetName() string {
	return u.Name
}

type LiveInfo struct {
	UserInfo
	LiveId   string `json:"live_id"`
	Title    string `json:"title"`
	Cover    string `json:"cover"`
	StartTs  int64  `json:"start_ts"`
	IsLiving bool   `json:"living"`

	liveStatusChanged bool
	liveTitleChanged  bool
}

func (l *LiveInfo) IsLive() bool {
	return true
}

func (l *LiveInfo) Living() bool {
	return l.IsLiving
}

func (l *LiveInfo) LiveStatusChanged() bool {
	return l.liveStatusChanged
}

func (l *LiveInfo) TitleChanged() bool {
	return l.liveTitleChanged
}

func (l *LiveInfo) Site() string {
	return Site
}

func (l *LiveInfo) Type() concern_type.Type {
	return Live
}

func (l *LiveInfo) Logger() *logrus.Entry {
	return logger.WithFields(logrus.Fields{
		"Site":  Site,
		"Uid":   l.Uid,
		"Name":  l.Name,
		"Title": l.Title,
		"Type":  l.Type().String(),
	})
}

type ConcernLiveNotify struct {
	GroupCode int64
	*LiveInfo
}

func (notify *ConcernLiveNotify) GetGroupCode() int64 {
	return notify.GroupCode
}

func (notify *ConcernLiveNotify) ToMessage() (m *mmsg.MSG) {
	m = mmsg.NewMSG()
	if notify.Living() {
		m.Textf("ACFUN-%s正在直播【%v】\n%v", notify.Name, notify.Title, notify.LiveUrl)
	} else {
		m.Textf("ACFUN-%s直播结束了", notify.Name)
	}
	m.ImageByUrl(notify.Cover, "[封面]", proxy_pool.PreferNone)
	return
}

func (notify *ConcernLiveNotify) Logger() *logrus.Entry {
	if notify == nil {
		return logger
	}
	return notify.LiveInfo.Logger().WithFields(localutils.GroupLogFields(notify.GroupCode))
}

func NewConcernLiveNotify(groupCode int64, info *LiveInfo) *ConcernLiveNotify {
	return &ConcernLiveNotify{
		GroupCode: groupCode,
		LiveInfo:  info,
	}
}
