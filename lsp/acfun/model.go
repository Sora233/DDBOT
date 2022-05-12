package acfun

import (
	"github.com/Sora233/DDBOT/lsp/concern_type"
	"github.com/Sora233/DDBOT/lsp/mmsg"
	"github.com/Sora233/DDBOT/lsp/mmsg/mt"
	"github.com/Sora233/DDBOT/lsp/template"
	localutils "github.com/Sora233/DDBOT/utils"
	"github.com/sirupsen/logrus"
	"sync"
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

	once              sync.Once
	msgCache          *mmsg.MSG
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

func (l *LiveInfo) GetMSG() *mmsg.MSG {
	l.once.Do(func() {
		var data = map[string]interface{}{
			"title":  l.Title,
			"name":   l.Name,
			"url":    l.LiveUrl,
			"cover":  l.Cover,
			"living": l.Living(),
		}
		var err error
		l.msgCache, err = template.LoadAndExec("notify.group.acfun.live.tmpl", data)
		if err != nil {
			logger.Errorf("acfun: LiveInfo LoadAndExec error %v", err)
		}
		return
	})
	return l.msgCache
}

type ConcernLiveNotify struct {
	Target mt.Target
	*LiveInfo
}

func (notify *ConcernLiveNotify) GetTarget() mt.Target {
	return notify.Target
}

func (notify *ConcernLiveNotify) ToMessage() (m *mmsg.MSG) {
	return notify.LiveInfo.GetMSG()
}

func (notify *ConcernLiveNotify) Logger() *logrus.Entry {
	if notify == nil {
		return logger
	}
	return notify.LiveInfo.Logger().WithFields(localutils.TargetFields(notify.GetTarget()))
}

func NewConcernLiveNotify(target mt.Target, info *LiveInfo) *ConcernLiveNotify {
	return &ConcernLiveNotify{
		Target:   target,
		LiveInfo: info,
	}
}
