package huya

import (
	"github.com/Sora233/DDBOT/lsp/concern_type"
	"github.com/Sora233/DDBOT/lsp/mmsg"
	"github.com/Sora233/DDBOT/proxy_pool"
	localutils "github.com/Sora233/DDBOT/utils"
	"github.com/sirupsen/logrus"
)

type concernEvent interface {
	Type() concern_type.Type
}

type LiveInfo struct {
	RoomId   string `json:"room_id"`
	RoomUrl  string `json:"room_url"`
	Avatar   string `json:"avatar"`
	Name     string `json:"name"`
	RoomName string `json:"room_name"`
	Living   bool   `json:"living"`

	LiveStatusChanged bool `json:"-"`
	LiveTitleChanged  bool `json:"-"`
}

func (m *LiveInfo) GetName() string {
	if m == nil {
		return ""
	}
	return m.Name
}

func (m *LiveInfo) Type() concern_type.Type {
	return Live
}

func (m *LiveInfo) ToString() string {
	if m == nil {
		return ""
	}
	bin, err := json.Marshal(m)
	if err != nil {
		return ""
	}
	return string(bin)
}

func (m *LiveInfo) Logger() *logrus.Entry {
	return logger.WithFields(logrus.Fields{
		"Site":   Site,
		"Name":   m.Name,
		"RoomId": m.RoomId,
		"Title":  m.RoomName,
		"Living": m.Living,
	})
}

func (m *LiveInfo) Site() string {
	return Site
}

type ConcernLiveNotify struct {
	LiveInfo
	GroupCode int64 `json:"group_code"`
}

func (notify *ConcernLiveNotify) GetGroupCode() int64 {
	return notify.GroupCode
}
func (notify *ConcernLiveNotify) GetUid() interface{} {
	return notify.RoomId
}

func (notify *ConcernLiveNotify) ToMessage() (m *mmsg.MSG) {
	if notify.Living {
		m.Textf("虎牙-%s正在直播【%v】\n%v", notify.Name, notify.RoomName, notify.RoomUrl)
	} else {
		m.Textf("虎牙-%s直播结束了", notify.Name)
	}
	m.ImageByUrl(notify.Avatar, "[封面]", proxy_pool.PreferNone)
	return
}

func (notify *ConcernLiveNotify) Logger() *logrus.Entry {
	if notify == nil {
		return logger
	}
	return notify.LiveInfo.Logger().WithFields(localutils.GroupLogFields(notify.GroupCode))
}

func (notify *ConcernLiveNotify) Type() concern_type.Type {
	return Live
}

func NewConcernLiveNotify(groupCode int64, l *LiveInfo) *ConcernLiveNotify {
	if l == nil {
		return nil
	}
	return &ConcernLiveNotify{
		*l,
		groupCode,
	}
}
