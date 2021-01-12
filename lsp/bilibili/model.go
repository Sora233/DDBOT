package bilibili

import (
	"encoding/json"
	"errors"
	"github.com/Sora233/Sora233-MiraiGo/concern"
)

type NewsInfo struct {
	UserInfo
	LastDynamicId int64                                       `json:"last_dynamic_id"`
	Timestamp     int32                                       `json:"timestamp"`
	Cards         []*DynamicSvrSpaceHistoryResponse_Data_Card `json:"-"`
}

func (n *NewsInfo) Type() EventType {
	return News
}

func (n *NewsInfo) GetCardWithImage(index int) (*CardWithImage, error) {
	if len(n.Cards) <= index || n.Cards[index].GetCard() == "" {
		return nil, errors.New("card not found or empty")
	}
	if n.Cards[index].GetDesc().GetType() == DynamicDescType_WithImage {
		var card = new(CardWithImage)
		err := json.Unmarshal([]byte(n.Cards[index].GetCard()), card)
		return card, err
	}
	return nil, errors.New("type mismatch")
}

func (n *NewsInfo) GetCardWithOrig(index int) (*CardWithOrig, error) {
	if len(n.Cards) <= index || n.Cards[index].GetCard() == "" {
		return nil, errors.New("card not found or empty")
	}
	if n.Cards[index].GetDesc().GetType() == DynamicDescType_WithOrigin {
		var card = new(CardWithOrig)
		err := json.Unmarshal([]byte(n.Cards[index].GetCard()), card)
		return card, err
	}
	return nil, errors.New("type mismatch")
}

func (n *NewsInfo) GetCardWithVideo(index int) (*CardWithVideo, error) {
	if len(n.Cards) <= index || n.Cards[index].GetCard() == "" {
		return nil, errors.New("card not found or empty")
	}
	if n.Cards[index].GetDesc().GetType() == DynamicDescType_WithVideo {
		var card = new(CardWithVideo)
		err := json.Unmarshal([]byte(n.Cards[index].GetCard()), card)
		return card, err
	}
	return nil, errors.New("type mismatch")
}

func (n *NewsInfo) GetCardTextOnly(index int) (*CardTextOnly, error) {
	if len(n.Cards) <= index || n.Cards[index].GetCard() == "" {
		return nil, errors.New("card not found or empty")
	}
	if n.Cards[index].GetDesc().GetType() == DynamicDescType_TextOnly {
		var card = new(CardTextOnly)
		err := json.Unmarshal([]byte(n.Cards[index].GetCard()), card)
		return card, err
	}
	return nil, errors.New("type mismatch")
}

func (n *NewsInfo) GetCardWithPost(index int) (*CardWithPost, error) {
	if len(n.Cards) <= index || n.Cards[index].GetCard() == "" {
		return nil, errors.New("card not found or empty")
	}
	if n.Cards[index].GetDesc().GetType() == DynamicDescType_WithPost {
		var card = new(CardWithPost)
		err := json.Unmarshal([]byte(n.Cards[index].GetCard()), card)
		return card, err
	}
	return nil, errors.New("type mismatch")
}

func (n *NewsInfo) ToString() string {
	if n == nil {
		return ""
	}
	content, _ := json.Marshal(n)
	return string(content)
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
	if userInfo == nil {
		return nil
	}
	return &LiveInfo{
		UserInfo:  *userInfo,
		Status:    status,
		LiveTitle: liveTitle,
		Cover:     cover,
	}
}

func NewNewsInfo(userInfo *UserInfo, dynamicId int64, timestamp int32) *NewsInfo {
	if userInfo == nil {
		return nil
	}
	return &NewsInfo{
		UserInfo:      *userInfo,
		LastDynamicId: dynamicId,
		Timestamp:     timestamp,
	}
}

func NewNewsInfoWithDetail(userInfo *UserInfo, cards []*DynamicSvrSpaceHistoryResponse_Data_Card) *NewsInfo {
	var dynamicId int64
	var timestamp int32
	if len(cards) > 0 {
		dynamicId = cards[0].GetDesc().GetDynamicId()
		timestamp = cards[0].GetDesc().GetTimestamp()
	}
	return &NewsInfo{
		UserInfo:      *userInfo,
		LastDynamicId: dynamicId,
		Timestamp:     timestamp,
		Cards:         cards,
	}
}

func NewConcernNewsNotify(groupCode int64, newsInfo *NewsInfo) *ConcernNewsNotify {
	if newsInfo == nil {
		return nil
	}
	return &ConcernNewsNotify{
		GroupCode: groupCode,
		NewsInfo:  *newsInfo,
	}
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
