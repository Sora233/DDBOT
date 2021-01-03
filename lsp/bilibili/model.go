package bilibili

import (
	"encoding/json"
	"errors"
	"github.com/Sora233/Sora233-MiraiGo/concern"
)

type NewsInfo struct {
	UserInfo
	NewsType   DynamicDescType
	OriginType DynamicDescType
	Card       interface{} // should be in CardWithImage / CardWithOrig / CardWithVideo / CardTextOnly
	Display    *DynamicSvrSpaceHistoryResponse_Data_Card_Display
}

func (n *NewsInfo) Type() EventType {
	return News
}

func (n *NewsInfo) GetCardWithImage() (*CardWithImage, error) {
	if n.NewsType == DynamicDescType_WithImage {
		return n.Card.(*CardWithImage), nil
	}
	return nil, errors.New("type mismatch")
}

func (n *NewsInfo) GetCardWithOrig() (*CardWithOrig, error) {
	if n.NewsType == DynamicDescType_WithOrigin {
		return n.Card.(*CardWithOrig), nil
	}
	return nil, errors.New("type mismatch")
}

func (n *NewsInfo) GetCardWithVideo() (*CardWithVideo, error) {
	if n.NewsType == DynamicDescType_WithVideo {
		return n.Card.(*CardWithVideo), nil
	}
	return nil, errors.New("type mismatch")
}

func (n *NewsInfo) GetCardTextOnly() (*CardTextOnly, error) {
	if n.NewsType == DynamicDescType_TextOnly {
		return n.Card.(*CardTextOnly), nil
	}
	return nil, errors.New("type mismatch")
}

type ConcernNewsNotify struct {
	GroupCode int64 `json:"group_code"`
	NewsInfo
}

func (cnn *ConcernNewsNotify) Type() concern.Type {
	return concern.BilibiliNews
}

type ConcernLiveNotify struct {
	GroupCode int64 `json:"group_code"`
	LiveInfo
}

func (cln *ConcernLiveNotify) Type() concern.Type {
	return concern.BibiliLive
}

type UserInfo struct {
	Mid     int64  `json:"mid"`
	Name    string `json:"name"`
	RoomId  int64  `json:"room_id"`
	RoomUrl string `json:"room_url"`
}

func (ui *UserInfo) ToString() string {
	if ui == nil {
		return ""
	}
	content, _ := json.Marshal(ui)
	return string(content)
}

type LiveInfo struct {
	UserInfo
	Status    LiveStatus `json:"status"`
	LiveTitle string     `json:"live_title"`
	Cover     string     `json:"cover"`
}

func (l *LiveInfo) Type() EventType {
	return Live
}

func (l *LiveInfo) ToString() string {
	if l == nil {
		return ""
	}
	content, _ := json.Marshal(l)
	return string(content)
}

func NewUserInfo(mid, roomId int64, name, url string) *UserInfo {
	return &UserInfo{
		Mid:     mid,
		RoomId:  roomId,
		Name:    name,
		RoomUrl: url,
	}
}

func NewLiveInfo(userInfo *UserInfo, liveTitle string, cover string, status LiveStatus) *LiveInfo {
	return &LiveInfo{
		UserInfo:  *userInfo,
		Status:    status,
		LiveTitle: liveTitle,
		Cover:     cover,
	}
}

func NewConcernNewsNotify() {
	panic("not impl")
}

func NewConcernLiveNotify(groupCode int64, liveInfo *LiveInfo) *ConcernLiveNotify {
	if liveInfo == nil {
		return nil
	}
	return &ConcernLiveNotify{
		GroupCode: groupCode,
		LiveInfo:  *liveInfo,
	}
}
