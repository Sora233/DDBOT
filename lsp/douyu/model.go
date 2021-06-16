package douyu

import "encoding/json"

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

func (m *LiveInfo) Type() EventType {
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
