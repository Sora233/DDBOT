package douyu

import (
	"github.com/Sora233/DDBOT/lsp/concern_type"
	"github.com/Sora233/DDBOT/lsp/mmsg"
	"github.com/Sora233/DDBOT/lsp/mmsg/mt"
	"github.com/Sora233/DDBOT/lsp/template"
	localutils "github.com/Sora233/DDBOT/utils"
	"github.com/sirupsen/logrus"
	"sync"
)

type LiveInfo struct {
	Nickname   string          `json:"nickname"`
	RoomId     int64           `json:"room_id"`
	RoomName   string          `json:"room_name"`
	RoomUrl    string          `json:"room_url"`
	ShowStatus ShowStatus      `json:"show_status"`
	VideoLoop  VideoLoopStatus `json:"videoLoop"`
	Avatar     *Avatar         `json:"avatar"`

	once              sync.Once
	msgCache          *mmsg.MSG
	liveStatusChanged bool
	liveTitleChanged  bool
}

func (m *LiveInfo) GetName() string {
	return m.Nickname
}

func (m *LiveInfo) TitleChanged() bool {
	return m.liveTitleChanged
}

func (m *LiveInfo) LiveStatusChanged() bool {
	return m.liveStatusChanged
}

func (m *LiveInfo) IsLive() bool {
	return true
}

func (m *LiveInfo) Site() string {
	return Site
}

func (m *LiveInfo) GetUid() interface{} {
	return m.RoomId
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

func (m *LiveInfo) GetMSG() *mmsg.MSG {
	m.once.Do(func() {
		var data = map[string]interface{}{
			"title":  m.RoomName,
			"name":   m.Nickname,
			"url":    m.RoomUrl,
			"cover":  m.GetAvatar().GetBig(),
			"living": m.Living(),
		}
		var err error
		m.msgCache, err = template.LoadAndExec("notify.group.douyu.live.tmpl", data)
		if err != nil {
			logger.Errorf("douyu: LiveInfo LoadAndExec error %v", err)
		}
		return
	})
	return m.msgCache
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
		return m.LiveStatusChanged()
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
	*LiveInfo
	Target mt.Target `json:"-"`
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

func NewConcernLiveNotify(target mt.Target, l *LiveInfo) *ConcernLiveNotify {
	if l == nil {
		return nil
	}
	return &ConcernLiveNotify{
		LiveInfo: l,
		Target:   target,
	}
}
