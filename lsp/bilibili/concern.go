package bilibili

import (
	"encoding/json"
	"errors"
	"github.com/Logiase/MiraiGo-Template/utils"
	"io/ioutil"
	"sync"
	"time"
)

var logger = utils.GetModuleLogger("bilibili-concern")

type ConcernType int64

const (
	ConcernLive ConcernType = 1 << iota
	ConcernNews
)

type ConcernLiveEvent struct {
	Mid    int64
	Status LiveStatus
}

type ConcernLiveNotify struct {
	Status    LiveStatus
	GroupCode int64
	Url       string
	LiveTitle string
	Username  string
	Cover     string
}

func NewConcernLiveNotify(groupCode int64, url, liveTitle, username, cover string, status LiveStatus) *ConcernLiveNotify {
	return &ConcernLiveNotify{
		Status:    status,
		GroupCode: groupCode,
		Url:       url,
		LiveTitle: liveTitle,
		Username:  username,
		Cover:     cover,
	}
}

type Concern struct {
	ConcernState map[int64]map[int64]ConcernType `json:"concern_state"`
	CurrentLive  map[int64]LiveStatus            `json:"current_live"`
	RoomMap      map[int64]int64                 `json:"room_map"`

	eventChan    chan *ConcernLiveEvent
	currentMutex *sync.RWMutex
	concernMutex *sync.RWMutex
	roomMutex    *sync.RWMutex

	c chan<- *ConcernLiveNotify

	stopped bool
	stop    chan interface{}
}

func NewConcern(c chan<- *ConcernLiveNotify) *Concern {
	concern := &Concern{
		ConcernState: make(map[int64]map[int64]ConcernType),
		CurrentLive:  make(map[int64]LiveStatus),
		eventChan:    make(chan *ConcernLiveEvent, 500),
		RoomMap:      make(map[int64]int64),
		currentMutex: new(sync.RWMutex),
		concernMutex: new(sync.RWMutex),
		roomMutex:    new(sync.RWMutex),
		c:            c,
		stop:         make(chan interface{}),
	}
	return concern
}

func (c *Concern) Start() {
	err := c.Load()
	if err != nil {
		logger.Errorf("bilibili concern load failed %v", err)
	}
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
	go c.notifyLoop()
}

func (c *Concern) Stop() {
	if err := c.Save(); err != nil {
		logger.Errorf("save concern failed")
	}
}

func (c *Concern) Add(groupCode int64, mid int64, ctype ConcernType) error {
	c.concernMutex.Lock()

	if _, ok := c.ConcernState[groupCode]; !ok {
		c.ConcernState[groupCode] = make(map[int64]ConcernType)
	}
	if _, ok := c.ConcernState[groupCode][mid]; !ok {
		c.ConcernState[groupCode][mid] = ctype
	} else {
		c.ConcernState[groupCode][mid] |= ctype
	}
	c.concernMutex.Unlock()

	c.currentMutex.Lock()
	c.CurrentLive[mid] = LiveStatus_NoLiving
	c.currentMutex.Unlock()
	return nil
}

func (c *Concern) Save() error {
	c.currentMutex.RLock()
	c.concernMutex.RLock()
	c.roomMutex.RLock()
	defer c.roomMutex.RUnlock()
	defer c.concernMutex.RUnlock()
	defer c.currentMutex.RUnlock()

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
	c.roomMutex.RLock()
	defer c.roomMutex.RUnlock()
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
	for event := range c.eventChan {
		if c.stopped {
			return
		}

		mid := event.Mid
		status := event.Status

		infoResp, err := XSpaceAccInfo(mid)
		if err != nil {
			logger.WithField("mid", mid).Errorf("get space info failed %v", err)
			continue
		}
		c.concernMutex.RLock()
		for groupCode, groupConcern := range c.ConcernState {
			if concern, ok := groupConcern[mid]; ok && concern&ConcernLive != 0 {
				logger.WithField("mid", mid).WithField("name", infoResp.GetData().GetName()).Debugf("notify")
				c.c <- NewConcernLiveNotify(groupCode,
					infoResp.GetData().GetLiveRoom().GetUrl(),
					infoResp.GetData().GetLiveRoom().GetTitle(),
					infoResp.GetData().GetName(),
					infoResp.GetData().GetLiveRoom().GetCover(),
					status)
			}
		}
		c.concernMutex.RUnlock()
	}
}

func (c *Concern) Fresh() {
	c.currentMutex.Lock()
	defer c.currentMutex.Unlock()
	logger.Debugf("start fresh")
	for mid, state := range c.CurrentLive {
		roomResp, err := c.getRoom(mid)
		if err != nil {
			logger.WithField("mid", mid).Errorf("bilibili concern fresh failed %v", err)
			continue
		}
		newStatus := LiveStatus(roomResp.GetData().GetLiveStatus())
		if newStatus != state {
			c.CurrentLive[mid] = newStatus
			c.eventChan <- &ConcernLiveEvent{
				Mid:    mid,
				Status: LiveStatus_Living,
			}
		}
	}
}

func (c *Concern) getRoom(mid int64) (*RoomInitResponse, error) {
	c.roomMutex.Lock()
	defer c.roomMutex.Unlock()

	logger.WithField("mid", mid).Debugf("getRoom")

	if _, ok := c.RoomMap[mid]; !ok {
		resp, err := XSpaceAccInfo(mid)
		if err != nil {
			return nil, err
		}
		roomId := resp.GetData().GetLiveRoom().GetRoomid()
		if roomId != 0 {
			c.RoomMap[mid] = roomId
		} else {
			return nil, errors.New("roomid get 0")
		}
		logger.WithField("mid", mid).Debugf("fill room %v", roomId)
	}

	roomId := c.RoomMap[mid]
	resp, err := RoomInit(roomId)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
