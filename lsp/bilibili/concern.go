package bilibili

import (
	"encoding/json"
	"errors"
	"github.com/Logiase/MiraiGo-Template/utils"
	"github.com/Sora233/Sora233-MiraiGo/lsp/concern"
	"io/ioutil"
	"sync"
	"time"
)

var logger = utils.GetModuleLogger("bilibili-concern")

type EventType int64

const (
	Live EventType = iota
	News
)

type ConcernEvent interface {
	Type() EventType
}

type ConcernLiveNotify struct {
	GroupCode int64 `json:"group_code"`
	LiveInfo
}

func (cln *ConcernLiveNotify) Type() concern.Type {
	return concern.BibiliLive
}

func NewConcernLiveNotify(groupCode, mid, roomId int64, url, liveTitle, username, cover string, status LiveStatus) *ConcernLiveNotify {
	return &ConcernLiveNotify{
		GroupCode: groupCode,
		LiveInfo:  *NewLiveInfo(mid, roomId, url, liveTitle, username, cover, status),
	}
}

// TODO
type ConcernNewsNotify struct {
}

func (cnn *ConcernNewsNotify) Type() concern.Type {
	return concern.BilibiliNews
}

func NewConcernNewsNotify() {
	panic("not impl")
}

type LiveInfo struct {
	UserInfo
	Status    LiveStatus `json:"status"`
	LiveTitle string     `json:"live_title"`
	Cover     string     `json:"cover"`
}

func NewLiveInfo(mid, roomId int64, url, liveTitle, username, cover string, status LiveStatus) *LiveInfo {
	return &LiveInfo{
		UserInfo: UserInfo{
			Mid:     mid,
			RoomId:  roomId,
			RoomUrl: url,
			Name:    username,
		},
		Status:    status,
		LiveTitle: liveTitle,
		Cover:     cover,
	}
}

func (l *LiveInfo) Type() EventType {
	return Live
}

type UserInfo struct {
	Mid     int64  `json:"mid"`
	Name    string `json:"name"`
	RoomId  int64  `json:"room_id"`
	RoomUrl string `json:"room_url"`
}

func NewUserInfo(mid, roomId int64, name, url string) *UserInfo {
	return &UserInfo{
		Mid:     mid,
		RoomId:  roomId,
		Name:    name,
		RoomUrl: url,
	}
}

type Concern struct {
	ConcernState map[int64]map[int64]concern.Type `json:"concern_state"`
	CurrentLive  map[int64]*LiveInfo              `json:"current_live"`

	eventChan    chan ConcernEvent
	currentMutex *sync.RWMutex
	concernMutex *sync.RWMutex

	notify chan<- concern.Notify

	stopped bool
	stop    chan interface{}
}

func NewConcern(notify chan<- concern.Notify) *Concern {
	c := &Concern{
		ConcernState: make(map[int64]map[int64]concern.Type),
		CurrentLive:  make(map[int64]*LiveInfo),
		eventChan:    make(chan ConcernEvent, 500),
		currentMutex: new(sync.RWMutex),
		concernMutex: new(sync.RWMutex),
		notify:       notify,
		stop:         make(chan interface{}),
	}
	return c
}

func (c *Concern) Start() {
	err := c.Load()
	if err != nil {
		logger.Errorf("bilibili concern load failed %v", err)
	}
	go c.notifyLoop()
	c.Fresh()
	go func() {
		ticker := time.NewTicker(time.Minute)
		for {
			select {
			case <-ticker.C:
				c.Fresh()
				if err := c.Save(); err != nil {
					logger.Errorf("bilibili concern save failed %v", err)
				} else {
					logger.Debugf("bilibili concern saved")
				}
			}
		}
	}()
}

func (c *Concern) Stop() {
	if err := c.Save(); err != nil {
		logger.Errorf("save concern failed")
	} else {
		logger.Debugf("save concern done")
	}
}

func (c *Concern) AddLiveRoom(groupCode int64, mid int64, roomId int64) error {
	c.concernMutex.Lock()
	if _, ok := c.ConcernState[groupCode]; !ok {
		c.ConcernState[groupCode] = make(map[int64]concern.Type)
	}
	if _, ok := c.ConcernState[groupCode][mid]; !ok {
		c.ConcernState[groupCode][mid] = concern.BibiliLive
	} else {
		c.ConcernState[groupCode][mid] |= concern.BibiliLive
	}
	c.concernMutex.Unlock()
	return nil
}

func (c *Concern) Remove(groupCode int64, mid int64, ctype concern.Type) error {
	c.currentMutex.RLock()
	c.concernMutex.RLock()
	defer c.currentMutex.RUnlock()
	defer c.concernMutex.RUnlock()

	if _, ok := c.ConcernState[groupCode]; !ok {
		return errors.New("未订阅的用户")
	}
	if _, ok := c.ConcernState[groupCode][mid]; !ok {
		return errors.New("未订阅的用户")
	} else if c.ConcernState[groupCode][mid]&ctype == 0 {
		return errors.New("未订阅的用户")
	}
	c.ConcernState[groupCode][mid] ^= ctype
	if c.ConcernState[groupCode][mid] == 0 {
		delete(c.ConcernState[groupCode], mid)
	}
	delete(c.CurrentLive, mid)
	return nil
}

func (c *Concern) ListLiving(groupCode int64) ([]*ConcernLiveNotify, error) {

	result := make([]*ConcernLiveNotify, 0)

	c.concernMutex.RLock()

	if _, ok := c.ConcernState[groupCode]; !ok {
		c.concernMutex.RUnlock()
		return result, nil
	}

	var concernMidSet = make(map[int64]bool)

	for mid, concernState := range c.ConcernState[groupCode] {
		if concernState&concern.BibiliLive != 0 {
			concernMidSet[mid] = true
		}
	}
	c.concernMutex.RUnlock()

	c.currentMutex.RLock()

	for mid := range concernMidSet {
		liveInfo, err := c.findUserLiving(mid, false)
		if err != nil || liveInfo == nil {
			logger.WithField("mid", mid).Errorf("get live info failed %v", err)
			c.currentMutex.RUnlock()
			return nil, err
		}
		if liveInfo.Status == LiveStatus_NoLiving {
			continue
		}
		cln := &ConcernLiveNotify{
			GroupCode: groupCode,
			LiveInfo:  *liveInfo,
		}
		result = append(result, cln)
	}
	c.currentMutex.RUnlock()

	return result, nil
}

func (c *Concern) Save() error {
	c.currentMutex.RLock()
	c.concernMutex.RLock()
	defer c.currentMutex.RUnlock()
	defer c.concernMutex.RUnlock()

	logger.Debugf("bilibili concern saving")

	content, err := json.Marshal(c)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(".bilibili-concern", content, 0644)
	if err != nil {
		return err
	}
	return nil
}
func (c *Concern) Load() error {
	c.currentMutex.RLock()
	c.concernMutex.RLock()
	defer c.currentMutex.RUnlock()
	defer c.concernMutex.RUnlock()

	content, err := ioutil.ReadFile(".bilibili-concern")
	if err != nil {
		return err
	}
	err = json.Unmarshal(content, c)
	if err != nil {
		return err
	}
	return nil
}

func (c *Concern) notifyLoop() {
	for ievent := range c.eventChan {
		if c.stopped {
			return
		}

		switch ievent.Type() {
		case Live:
			event := (ievent).(*LiveInfo)
			mid := event.Mid
			logger.WithField("mid", event.Mid).
				WithField("roomid", event.RoomId).
				WithField("title", event.LiveTitle).
				WithField("status", event.Status.String()).
				Debugf("event debug")
			c.concernMutex.RLock()
			switch event.Status {
			case LiveStatus_Living:
				for groupCode, groupConcern := range c.ConcernState {
					if _concern, ok := groupConcern[mid]; ok && _concern&concern.BibiliLive != 0 {
						logger.WithField("mid", mid).
							WithField("name", event.Name).
							Debugf("living notify")
						c.notify <- NewConcernLiveNotify(groupCode,
							event.Mid,
							event.RoomId,
							event.RoomUrl,
							event.LiveTitle,
							event.Name,
							event.Cover,
							event.Status)
					}
				}
			case LiveStatus_NoLiving:
				for groupCode, groupConcern := range c.ConcernState {
					if _concern, ok := groupConcern[mid]; ok && _concern&concern.BibiliLive != 0 {
						logger.WithField("mid", mid).
							Debugf("noliving notify")
						c.notify <- NewConcernLiveNotify(groupCode,
							event.Mid,
							event.RoomId,
							"",
							"",
							"",
							"",
							event.Status)
					}
				}
			}
			c.concernMutex.RUnlock()
		case News:
			// TODO
			logger.Errorf("concern event news not supported")
		}

	}
}

func (c *Concern) Fresh() {
	c.currentMutex.Lock()
	c.concernMutex.RLock()
	defer c.currentMutex.Unlock()
	defer c.concernMutex.RUnlock()
	logger.Debugf("start fresh")

	var (
		concernMidSet = make(map[int64]bool)
	)

	for _, groupConcern := range c.ConcernState {
		for mid, _ := range groupConcern {
			concernMidSet[mid] = true
		}
	}

	for mid := range concernMidSet {
		oldInfo, _ := c.findUserLiving(mid, false)
		liveInfo, err := c.findUserLiving(mid, true)
		if err != nil {
			logger.WithField("mid", mid).Debugf("bilibili concern fresh failed %v", err)
			continue
		}
		if oldInfo == nil || oldInfo.Status != liveInfo.Status && liveInfo.Status == LiveStatus_Living {
			c.eventChan <- liveInfo
		}
	}

}

func (c *Concern) findUserLiving(mid int64, load bool) (*LiveInfo, error) {
	if load {
		resp, err := XSpaceAccInfo(mid)
		if err != nil {
			return nil, err
		}
		c.CurrentLive[mid] = NewLiveInfo(mid,
			resp.GetData().GetLiveRoom().GetRoomid(),
			resp.GetData().GetLiveRoom().GetUrl(),
			resp.GetData().GetLiveRoom().GetTitle(),
			resp.GetData().GetName(),
			resp.GetData().GetLiveRoom().GetCover(),
			LiveStatus(resp.GetData().GetLiveRoom().GetLiveStatus()),
		)
	}
	return c.CurrentLive[mid], nil
}
