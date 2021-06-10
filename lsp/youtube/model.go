package youtube

import (
	"encoding/json"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Sora233/DDBOT/concern"
	"github.com/Sora233/DDBOT/proxy_pool"
	localutils "github.com/Sora233/DDBOT/utils"
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

func (notify *ConcernNotify) ToMessage() []message.IMessageElement {
	var result []message.IMessageElement
	if notify.IsLive() {
		if notify.IsLiving() {
			result = append(result, localutils.MessageTextf("YTB-%v正在直播：\n%v\n", notify.ChannelName, notify.VideoTitle))
		} else {
			result = append(result, localutils.MessageTextf("YTB-%v发布了直播预约：\n%v\n时间：%v\n", notify.ChannelName, notify.VideoTitle, localutils.TimestampFormat(notify.VideoTimestamp)))
		}
	} else if notify.IsVideo() {
		result = append(result, localutils.MessageTextf("YTB-%s发布了新视频：\n%v\n", notify.ChannelName, notify.VideoTitle))
	}
	groupImg, err := localutils.UploadGroupImageByUrl(notify.GroupCode, notify.Cover, false, proxy_pool.PreferOversea)
	if err != nil {
		logger.WithField("channel_name", notify.ChannelName).
			WithField("video_id", notify.VideoId).
			WithField("group_code", notify.GroupCode).
			Errorf("upload cover failed %v", err)
	} else {
		result = append(result, groupImg)
	}
	result = append(result, message.NewText(VideoViewUrl(notify.VideoId)+"\n"))
	return result
}

func (notify *ConcernNotify) Type() concern.Type {
	if notify.IsLive() {
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

func NewUserInfo(channelId, channelName string) *UserInfo {
	return &UserInfo{
		ChannelId:   channelId,
		ChannelName: channelName,
	}
}
