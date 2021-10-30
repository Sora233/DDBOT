package douyu

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
	Nickname   string          `json:"nickname"`
	RoomId     int64           `json:"room_id"`
	RoomName   string          `json:"room_name"`
	RoomUrl    string          `json:"room_url"`
	ShowStatus ShowStatus      `json:"show_status"`
	VideoLoop  VideoLoopStatus `json:"videoLoop"`
	Avatar     *Avatar         `json:"avatar"`

	LiveStatusChanged bool `json:"-"`
	LiveTitleChanged  bool `json:"-"`
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

func (m *LiveInfo) Living() bool {
	return m.ShowStatus == ShowStatus_Living && m.VideoLoop == VideoLoopStatus_Off
}

func (m *LiveInfo) Type() concern_type.Type {
	return Live
}

func (m *LiveInfo) GetNickname() string {
	if m != nil {
		return m.Nickname
	}
	return ""
}

func (m *LiveInfo) GetRoomId() int64 {
	if m != nil {
		return m.RoomId
	}
	return 0
}

func (m *LiveInfo) GetRoomName() string {
	if m != nil {
		return m.RoomName
	}
	return ""
}

func (m *LiveInfo) GetRoomUrl() string {
	if m != nil {
		return m.RoomUrl
	}
	return ""
}

func (m *LiveInfo) GetShowStatus() ShowStatus {
	if m != nil {
		return m.ShowStatus
	}
	return ShowStatus_Unknown
}

func (m *LiveInfo) GetVideoLoop() VideoLoopStatus {
	if m != nil {
		return m.VideoLoop
	}
	return VideoLoopStatus_Off
}

func (m *LiveInfo) GetAvatar() *Avatar {
	if m != nil {
		return m.Avatar
	}
	return nil
}

func (m *LiveInfo) GetLiveStatusChanged() bool {
	if m != nil {
		return m.LiveStatusChanged
	}
	return false
}

func (m *LiveInfo) Logger() *logrus.Entry {
	return logger.WithFields(logrus.Fields{
		"Site":      Site,
		"Name":      m.Nickname,
		"Title":     m.RoomName,
		"Status":    m.ShowStatus.String(),
		"VideoLoop": m.GetVideoLoop().String(),
	})
}

type ConcernLiveNotify struct {
	LiveInfo
	GroupCode int64 `json:"group_code"`
}

func (notify *ConcernLiveNotify) Type() concern_type.Type {
	return Live
}
func (notify *ConcernLiveNotify) GetGroupCode() int64 {
	return notify.GroupCode
}
func (notify *ConcernLiveNotify) GetUid() interface{} {
	return notify.RoomId
}

func (notify *ConcernLiveNotify) Site() string {
	return Site
}

func (notify *ConcernLiveNotify) ToMessage() (m *mmsg.MSG) {
	m = mmsg.NewMSG()
	switch notify.ShowStatus {
	case ShowStatus_Living:
		m.Textf("斗鱼-%s正在直播【%v】\n%v", notify.Nickname, notify.RoomName, notify.RoomUrl)
	case ShowStatus_NoLiving:
		m.Textf("斗鱼-%s直播结束了", notify.Nickname)
	}
	m.ImageByUrl(notify.GetAvatar().GetBig(), "[封面]", proxy_pool.PreferNone)
	return m
}

func (notify *ConcernLiveNotify) Logger() *logrus.Entry {
	if notify == nil {
		return logger
	}
	return notify.LiveInfo.Logger().WithFields(localutils.GroupLogFields(notify.GroupCode))
}

func NewConcernLiveNotify(groupCode int64, l *LiveInfo) *ConcernLiveNotify {
	if l == nil {
		return nil
	}
	return &ConcernLiveNotify{
		LiveInfo:  *l,
		GroupCode: groupCode,
	}
}
