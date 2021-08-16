package huya

import (
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Sora233/DDBOT/concern"
	"github.com/Sora233/DDBOT/proxy_pool"
	localutils "github.com/Sora233/DDBOT/utils"
	"github.com/sirupsen/logrus"
)

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

func (m *LiveInfo) Type() EventType {
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

func (notify *ConcernLiveNotify) ToMessage() []message.IMessageElement {
	log := notify.Logger()
	var result []message.IMessageElement
	if notify.Living {
		result = append(result, localutils.MessageTextf("虎牙-%s正在直播【%v】\n%v", notify.Name, notify.RoomName, notify.RoomUrl))
	} else {
		result = append(result, localutils.MessageTextf("虎牙-%s直播结束了", notify.Name))
	}
	cover, err := localutils.UploadGroupImageByUrl(notify.GroupCode, notify.Avatar, false, proxy_pool.PreferNone)
	if err != nil {
		log.WithField("Avatar", notify.Avatar).Errorf("upload avatar failed %v", err)
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

func (notify *ConcernLiveNotify) Type() concern.Type {
	return concern.HuyaLive
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
