package youtube

import (
	"encoding/json"
	"github.com/Sora233/Sora233-MiraiGo/concern"
)

type UserInfo struct {
	ChannelId   string `json:"channel_id"`
	ChannelName string `json:"channel_name"`
}

type EventType int64

const (
	Video EventType = iota
	Live
)

type ConcernEvent interface {
	Type() EventType
}

// VideoInfo may be a video or a live, depend on the VideoType
type VideoInfo struct {
	UserInfo
	Cover          string      `json:"cover"`
	VideoId        string      `json:"video_id"`
	VideoTitle     string      `json:"video_title"`
	VideoType      VideoType   `json:"video_type"`
	VideoStatus    VideoStatus `json:"video_status"`
	VideoTimestamp int64       `json:"video_timestamp"`
}

func (v *VideoInfo) Type() EventType {
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

func (c *ConcernNotify) Type() concern.Type {
	if c.IsLive() {
		return concern.YoutubeLive
	} else {
		return concern.YoutubeVideo
	}
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
