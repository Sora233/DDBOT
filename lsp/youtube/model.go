package youtube

import "github.com/Sora233/Sora233-MiraiGo/concern"

type UserInfo struct {
	ChannelId string `json:"channel_id"`
}

type VideoInfo struct {
	VideoId        string      `json:"video_id"`
	VideoTitle     string      `json:"video_title"`
	VideoType      VideoType   `json:"video_type"`
	VideoStatus    VideoStatus `json:"video_status"`
	VideoTimestamp int64       `json:"video_timestamp"`
}

type Info struct {
	VideoInfo []*VideoInfo `json:"video_info"`
}

type ConcernNotify struct {
	Info
	GroupCode int64 `json:"group_code"`
}

func (c *ConcernNotify) Type() concern.Type {
	return concern.Youtube
}
