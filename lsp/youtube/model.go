package youtube

import (
	"github.com/Sora233/DDBOT/lsp/concern_type"
	"github.com/Sora233/DDBOT/lsp/mmsg"
	"github.com/Sora233/DDBOT/proxy_pool"
	localutils "github.com/Sora233/DDBOT/utils"
	"github.com/sirupsen/logrus"
)

type UserInfo struct {
	ChannelId   string `json:"channel_id"`
	ChannelName string `json:"channel_name"`
}

func (ui *UserInfo) GetChannelName() string {
	if ui == nil {
		return ""
	}
	return ui.ChannelName
}

const (
	Video concern_type.Type = "news"
	Live  concern_type.Type = "live"
)

// VideoInfo may be a video or a live, depend on the VideoType
type VideoInfo struct {
	UserInfo
	Cover          string      `json:"cover"`
	VideoId        string      `json:"video_id"`
	VideoTitle     string      `json:"video_title"`
	VideoType      VideoType   `json:"video_type"`
	VideoStatus    VideoStatus `json:"video_status"`
	VideoTimestamp int64       `json:"video_timestamp"`

	LiveStatusChanged bool `json:"-"`
	LiveTitleChanged  bool `json:"-"`
}

func (v *VideoInfo) Site() string {
	return Site
}

func (v *VideoInfo) GetUid() interface{} {
	return v.ChannelId
}

func (v *VideoInfo) Logger() *logrus.Entry {
	return logger.WithFields(logrus.Fields{
		"Site":        Site,
		"ChannelId":   v.ChannelId,
		"ChannelName": v.ChannelName,
		"VideoId":     v.VideoId,
		"VideoType":   v.VideoType.String(),
		"VideoTitle":  v.VideoTitle,
		"VideoStatus": v.VideoStatus.String(),
	})
}

func (v *VideoInfo) Type() concern_type.Type {
	if v.IsLive() {
		return Live
	} else {
		return Video
	}
}

func (v *VideoInfo) IsLive() bool {
	if v == nil {
		return false
	}
	return v.VideoType == VideoType_FirstLive || v.VideoType == VideoType_Live
}

func (v *VideoInfo) IsLiving() bool {
	if v == nil {
		return false
	}
	return v.IsLive() && v.VideoStatus == VideoStatus_Living
}

func (v *VideoInfo) IsWaiting() bool {
	if v == nil {
		return false
	}
	return v.IsLive() && v.VideoStatus == VideoStatus_Waiting
}

func (v *VideoInfo) IsVideo() bool {
	if v == nil {
		return false
	}
	return v.VideoType == VideoType_Video
}

type Info struct {
	VideoInfo []*VideoInfo `json:"video_info"`
	UserInfo
}

func (i *Info) ToString() string {
	if i == nil {
		return ""
	}
	b, _ := json.Marshal(i)
	return string(b)
}

func NewInfo(vinfo []*VideoInfo) *Info {
	info := new(Info)
	info.VideoInfo = vinfo
	if len(vinfo) > 0 {
		info.ChannelId = vinfo[0].ChannelId
		info.ChannelName = vinfo[0].ChannelName
	}
	return info
}

type ConcernNotify struct {
	VideoInfo
	GroupCode int64 `json:"group_code"`
}

func (notify *ConcernNotify) GetGroupCode() int64 {
	return notify.GroupCode
}

func (notify *ConcernNotify) ToMessage() (m *mmsg.MSG) {
	m = mmsg.NewMSG()
	if notify.IsLive() {
		if notify.IsLiving() {
			m.Textf("YTB-%v正在直播：\n%v\n", notify.ChannelName, notify.VideoTitle)
		} else {
			m.Textf("YTB-%v发布了直播预约：\n%v\n时间：%v\n",
				notify.ChannelName, notify.VideoTitle, localutils.TimestampFormat(notify.VideoTimestamp))
		}
	} else if notify.IsVideo() {
		m.Textf("YTB-%s发布了新视频：\n%v\n", notify.ChannelName, notify.VideoTitle)
	}
	m.ImageByUrl(notify.Cover, "[封面]", proxy_pool.PreferOversea)
	m.Text(VideoViewUrl(notify.VideoId) + "\n")
	return
}

func (notify *ConcernNotify) Logger() *logrus.Entry {
	if notify == nil {
		return logger
	}
	return notify.VideoInfo.Logger().WithFields(localutils.GroupLogFields(notify.GroupCode))
}

func NewConcernNotify(groupCode int64, info *VideoInfo) *ConcernNotify {
	if info == nil {
		return nil
	}
	return &ConcernNotify{
		VideoInfo: *info,
		GroupCode: groupCode,
	}
}

func NewUserInfo(channelId, channelName string) *UserInfo {
	return &UserInfo{
		ChannelId:   channelId,
		ChannelName: channelName,
	}
}
