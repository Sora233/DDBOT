package bilibili

import (
	"encoding/json"
	"errors"
	"github.com/Logiase/MiraiGo-Template/utils"
	localdb "github.com/Sora233/Sora233-MiraiGo/lsp/buntdb"
	"github.com/Sora233/Sora233-MiraiGo/lsp/concern"
	"github.com/tidwall/buntdb"
	"strconv"
	"strings"
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

func (l *LiveInfo) ToString() string {
	if l == nil {
		return ""
	}
	content, _ := json.Marshal(l)
	return string(content)
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
	eventChan chan ConcernEvent

	notify chan<- concern.Notify

	stopped bool
	stop    chan interface{}
}

func NewConcern(notify chan<- concern.Notify) *Concern {
	c := &Concern{
		eventChan: make(chan ConcernEvent, 500),
		notify:    notify,
		stop:      make(chan interface{}),
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
		timer := time.NewTimer(time.Minute)
		for {
			select {
			case <-timer.C:
				c.Fresh()
				timer.Reset(time.Minute)
			}
		}
	}()
}

func (c *Concern) Stop() {
}

func (c *Concern) AddLiveRoom(groupCode int64, mid int64, roomId int64) error {
	db, err := localdb.GetClient()
	if err != nil {
		return err
	}
	err = db.Update(func(tx *buntdb.Tx) error {
		stateKey := c.ConcernStateKey(groupCode, mid)
		val, err := tx.Get(stateKey)
		if err == buntdb.ErrNotFound {
			tx.Set(stateKey, concern.BibiliLive.ToString(), nil)
		} else if err == nil {
			newVal := concern.FromString(val) | concern.BibiliLive
			tx.Set(stateKey, newVal.ToString(), nil)
		} else {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (c *Concern) Remove(groupCode int64, mid int64, ctype concern.Type) error {
	db, err := localdb.GetClient()
	if err != nil {
		return err
	}
	err = db.Update(func(tx *buntdb.Tx) error {
		stateKey := c.ConcernStateKey(groupCode, mid)
		_, err := tx.Delete(stateKey)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (c *Concern) ListLiving(groupCode int64, all bool) ([]*ConcernLiveNotify, error) {
	result := make([]*ConcernLiveNotify, 0)

	db, err := localdb.GetClient()
	if err != nil {
		return result, err
	}
	err = db.View(func(tx *buntdb.Tx) error {
		var iterErr error
		tx.Ascend(c.CurrentLiveKey(), func(key, value string) bool {
			liveInfo := &LiveInfo{}
			if iterErr = json.Unmarshal([]byte(value), liveInfo); err != nil {
				logger.WithField("key", key).
					WithField("value", value).
					Errorf("json unmarshal liveLnfo failed %v", iterErr)
				return false
			}
			if all || liveInfo.Status == LiveStatus_Living {
				cln := &ConcernLiveNotify{
					GroupCode: groupCode,
					LiveInfo:  *liveInfo,
				}
				result = append(result, cln)
			}
			return true
		})
		return iterErr
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (c *Concern) Load() error {
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
			logger.WithField("mid", event.Mid).
				WithField("roomid", event.RoomId).
				WithField("title", event.LiveTitle).
				WithField("status", event.Status.String()).
				Debugf("event debug")
			db, err := localdb.GetClient()
			if err != nil {
				logger.Errorf("get db failed %v", err)
				continue
			}
			err = db.View(func(tx *buntdb.Tx) error {
				var iterErr error
				tx.Ascend(c.ConcernStateKey(), func(key, value string) bool {
					var (
						groupCode, mid int64
						notify         *ConcernLiveNotify
					)
					groupCode, mid, iterErr = c.ParseConcernStateKey(key)
					if iterErr != nil {
						return false
					}
					if mid != event.Mid || concern.FromString(value)&concern.BibiliLive == 0 {
						return true
					}
					if event.Status == LiveStatus_Living {
						logger.WithField("mid", event.Mid).
							WithField("name", event.Name).
							Debugf("living notify")
						notify = NewConcernLiveNotify(groupCode, event.Mid, event.RoomId, event.RoomUrl,
							event.LiveTitle, event.Name, event.Cover, event.Status,
						)
					} else if event.Status == LiveStatus_NoLiving {
						logger.WithField("mid", event.Mid).
							WithField("name", event.Name).
							Debugf("noliving notify")
						notify = NewConcernLiveNotify(groupCode, event.Mid, event.RoomId, "",
							"", "", "", event.Status,
						)
					} else {
						logger.Errorf("unknown live status %v", event.Status.String())
					}
					if notify != nil {
						c.notify <- notify
					}
					return true
				})
				return iterErr
			})
			if err != nil {
				logger.WithField("mid", event.Mid).
					WithField("name", event.Name).
					Errorf("notify failed err %v", err)
			}
		case News:
			// TODO
			logger.Errorf("concern event news not supported")
		}

	}
}

func (c *Concern) Fresh() {
	logger.Debugf("fresh start")
	defer logger.Debugf("fresh end")

	db, err := localdb.GetClient()
	if err != nil {
		logger.Errorf("get db failed %v", err)
		return
	}

	var (
		concernMidSet = make(map[int64]bool)
	)

	err = db.View(func(tx *buntdb.Tx) error {
		var iterErr error
		tx.Ascend(c.ConcernStateKey(), func(key, value string) bool {
			var mid int64
			_, mid, iterErr = c.ParseConcernStateKey(key)
			if iterErr != nil {
				logger.WithField("key", key).
					WithField("value", value).
					Debugf("ParseConcernStateKey err %v", err)
				return false
			}
			concernMidSet[mid] = true
			return true
		})
		return iterErr
	})

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

func (c *Concern) NamedKey(name string, keys []interface{}) string {
	newkey := []interface{}{name}
	for _, key := range keys {
		newkey = append(newkey, key)
	}
	return localdb.Key(newkey...)
}

func (c *Concern) ConcernStateKey(keys ...interface{}) string {
	return c.NamedKey("ConcernState", keys)
}
func (c *Concern) CurrentLiveKey(keys ...interface{}) string {
	return c.NamedKey("CurrentLive", keys)
}
func (c *Concern) ParseConcernStateKey(key string) (groupCode int64, mid int64, err error) {
	keys := strings.Split(key, ":")
	if len(keys) != 3 || keys[0] != "ConcernState" {
		return 0, 0, errors.New("invalid concern state key")
	}
	groupCode, err = strconv.ParseInt(keys[1], 10, 64)
	if err != nil {
		return 0, 0, err
	}
	mid, err = strconv.ParseInt(keys[2], 10, 64)
	if err != nil {
		return 0, 0, err
	}
	return groupCode, mid, nil
}
func (c *Concern) ParseCurrentLiveKey(key string) (mid int64, err error) {
	keys := strings.Split(key, ":")
	if len(keys) != 2 || keys[0] != "CurrentLive" {
		return 0, errors.New("invalid current live key")
	}
	mid, err = strconv.ParseInt(keys[1], 10, 64)
	if err != nil {
		return 0, err
	}
	return mid, nil
}

func (c *Concern) findUserLiving(mid int64, load bool) (*LiveInfo, error) {
	db, err := localdb.GetClient()
	if err != nil {
		return nil, err
	}
	if load {
		resp, err := XSpaceAccInfo(mid)
		if err != nil {
			return nil, err
		}
		newInfo := NewLiveInfo(mid,
			resp.GetData().GetLiveRoom().GetRoomid(),
			resp.GetData().GetLiveRoom().GetUrl(),
			resp.GetData().GetLiveRoom().GetTitle(),
			resp.GetData().GetName(),
			resp.GetData().GetLiveRoom().GetCover(),
			LiveStatus(resp.GetData().GetLiveRoom().GetLiveStatus()),
		)
		err = db.Update(func(tx *buntdb.Tx) error {
			tx.Set(c.CurrentLiveKey(mid), newInfo.ToString(), nil)
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	var liveInfo = &LiveInfo{}
	err = db.View(func(tx *buntdb.Tx) error {
		val, err := tx.Get(c.CurrentLiveKey(mid))
		if err != nil {
			return err
		}
		err = json.Unmarshal([]byte(val), liveInfo)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return liveInfo, nil
}
