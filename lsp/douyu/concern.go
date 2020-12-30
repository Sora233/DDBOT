package douyu

import (
	"encoding/json"
	"errors"
	"fmt"
	miraiBot "github.com/Logiase/MiraiGo-Template/bot"
	"github.com/Logiase/MiraiGo-Template/utils"
	"github.com/Sora233/Sora233-MiraiGo/concern"
	localdb "github.com/Sora233/Sora233-MiraiGo/lsp/buntdb"
	"github.com/tidwall/buntdb"
	"math/rand"
	"time"
)

var logger = utils.GetModuleLogger("douyu-concern")

type EventType int64

const (
	Live EventType = iota
)

type ConcernEvent interface {
	Type() EventType
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

func (m *LiveInfo) Type() EventType {
	return Live
}

type ConcernLiveNotify struct {
	LiveInfo
	GroupCode int64 `json:"group_code"`
}

func (notify *ConcernLiveNotify) Type() concern.Type {
	return concern.DouyuLive
}

type Concern struct {
	eventChan chan ConcernEvent
	notify    chan<- concern.Notify
	stop      chan interface{}
	stopped   bool
}

func (c *Concern) Start() {
	db, err := localdb.GetClient()
	if err == nil {
		db.CreateIndex(c.ConcernStateKey(), c.ConcernStateKey("*"), buntdb.IndexString)
		db.CreateIndex(c.CurrentLiveKey(), c.CurrentLiveKey("*"), buntdb.IndexString)
		db.CreateIndex(c.FreshKey(), c.FreshKey("*"), buntdb.IndexString)
	}
	go c.notifyLoop()
	go func() {
		timer := time.NewTimer(time.Second * 5)
		for {
			select {
			case <-timer.C:
				c.FreshConcern()
				timer.Reset(time.Second * 5)
			}
		}
	}()
}

func (c *Concern) Add(groupCode int64, id int64, ctype concern.Type) (*LiveInfo, error) {
	db, err := localdb.GetClient()
	if err != nil {
		return nil, err
	}
	log := logger.WithField("GroupCode", groupCode)

	err = db.View(func(tx *buntdb.Tx) error {
		val, err := tx.Get(c.ConcernStateKey(groupCode, id))
		if err == nil {
			if concern.FromString(val).ContainAll(ctype) {
				return errors.New("已经watch过了")
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	betardResp, err := Betard(id)
	if err != nil {
		log.WithField("id", id).Error(err)
		return nil, fmt.Errorf("查询房间信息失败 %v - %v", id, err)
	}
	err = db.Update(func(tx *buntdb.Tx) error {
		stateKey := c.ConcernStateKey(groupCode, id)
		val, err := tx.Get(stateKey)
		if err == buntdb.ErrNotFound {
			tx.Set(stateKey, ctype.String(), nil)
		} else if err == nil {
			newVal := concern.FromString(val).Add(ctype)
			tx.Set(stateKey, newVal.String(), nil)
		} else {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	liveInfo := &LiveInfo{
		Nickname:   betardResp.GetRoom().GetNickname(),
		RoomId:     betardResp.GetRoom().GetRoomId(),
		RoomName:   betardResp.GetRoom().GetRoomName(),
		RoomUrl:    betardResp.GetRoom().GetRoomUrl(),
		ShowStatus: betardResp.GetRoom().GetShowStatus(),
		Avatar:     betardResp.GetRoom().GetAvatar(),
	}
	return liveInfo, nil
}

func (c *Concern) Remove(groupCode int64, id int64, ctype concern.Type) error {
	db, err := localdb.GetClient()
	if err != nil {
		return err
	}
	err = db.Update(func(tx *buntdb.Tx) error {
		stateKey := c.ConcernStateKey(groupCode, id)
		val, err := tx.Get(stateKey)
		if err != nil {
			return err
		}
		oldState := concern.FromString(val)
		newState := oldState.Remove(ctype)
		_, _, err = tx.Set(stateKey, newState.String(), nil)
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
	var result []*ConcernLiveNotify
	db, err := localdb.GetClient()
	if err != nil {
		return result, err
	}
	err = db.View(func(tx *buntdb.Tx) error {
		var iterErr error
		var mid int64

		var concernMid []int64

		iterErr = tx.Ascend(c.ConcernStateKey(groupCode), func(concernStateKey, state string) bool {
			_, mid, iterErr = c.ParseConcernStateKey(concernStateKey)
			if iterErr != nil {
				logger.WithField("key", concernStateKey).
					Errorf("ParseConcernStateKey failed %v", iterErr)
				return false
			}
			if concern.FromString(state).ContainAll(concern.DouyuLive) {
				concernMid = append(concernMid, mid)
			}
			return true
		})

		if iterErr != nil {
			return iterErr
		}

		if len(concernMid) != 0 {
			result = make([]*ConcernLiveNotify, 0)
		}

		for _, mid = range concernMid {
			key := c.CurrentLiveKey(mid)
			value, err := tx.Get(key)
			if err != nil {
				if err != buntdb.ErrNotFound {
					logger.WithField("key", key).
						Errorf("get currentLive err %v", err)
				}
				continue
			}
			liveInfo := &LiveInfo{}
			if iterErr = json.Unmarshal([]byte(value), liveInfo); iterErr != nil {
				logger.WithField("key", key).
					WithField("value", value).
					Errorf("json unmarshal liveLnfo failed %v", iterErr)
				continue
			}
			if all || liveInfo.ShowStatus == ShowStatus_Living {
				cln := &ConcernLiveNotify{
					GroupCode: groupCode,
					LiveInfo:  *liveInfo,
				}
				result = append(result, cln)
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (c *Concern) FreshConcern() {
	db, err := localdb.GetClient()
	if err != nil {
		logger.Errorf("get db failed %v", err)
		return
	}

	var (
		concernIdSet = make(map[int64]concern.Type)
		freshConcern []struct {
			Id          int64
			ConcernType concern.Type
		}
	)

	err = db.View(func(tx *buntdb.Tx) error {
		var iterErr error
		iterErr = tx.Ascend(c.ConcernStateKey(), func(key, value string) bool {
			var mid int64
			_, mid, iterErr = c.ParseConcernStateKey(key)
			if iterErr != nil {
				logger.WithField("key", key).
					WithField("value", value).
					Debugf("ParseConcernStateKey err %v", err)
				return false
			}
			concernIdSet[mid] = concern.FromString(value)
			return true
		})
		return iterErr
	})
	if err != nil {
		logger.Errorf("fresh list key failed %v", err)
	}

	for id, concernType := range concernIdSet {
		err = db.Update(func(tx *buntdb.Tx) error {
			var err error
			freshKey := c.FreshKey(id)
			_, err = tx.Get(freshKey)
			if err == buntdb.ErrNotFound {
				ttl := time.Minute + time.Duration(rand.Intn(120))*time.Second
				_, _, err = tx.Set(freshKey, "", &buntdb.SetOptions{Expires: true, TTL: ttl})
				if err != nil {
					return err
				}
				freshConcern = append(freshConcern, struct {
					Id          int64
					ConcernType concern.Type
				}{Id: id, ConcernType: concernType})
			}
			return nil
		})
		if err != nil {
			logger.WithField("mid", id).Errorf("scan concernMidSet failed %v", err)
		}
	}

	for _, item := range freshConcern {
		if item.ConcernType.ContainAll(concern.DouyuLive) {
			oldInfo, _ := c.findRoom(item.Id, false)
			liveInfo, err := c.findRoom(item.Id, true)
			if err != nil {
				logger.WithField("mid", item.Id).Debugf("fresh failed %v", err)
				continue
			}
			if oldInfo == nil {
				c.eventChan <- liveInfo
			} else if oldInfo.VideoLoop == VideoLoopStatus_On {
				continue
			} else if oldInfo.ShowStatus != liveInfo.ShowStatus || oldInfo.RoomName != liveInfo.RoomName {
				c.eventChan <- liveInfo
			}
		}
	}
}

func (c *Concern) notifyLoop() {
	for ievent := range c.eventChan {
		if c.stopped {
			return
		}
		switch ievent.Type() {
		case Live:
			event := ievent.(*LiveInfo)
			logger.WithField("name", event.GetNickname()).
				WithField("roomid", event.GetRoomId()).
				WithField("title", event.GetRoomName()).
				WithField("status", event.GetShowStatus().String()).
				Debugf("event debug")
			db, err := localdb.GetClient()
			if err != nil {
				logger.Errorf("get db failed %v", err)
				continue
			}
			err = db.View(func(tx *buntdb.Tx) error {
				var iterErr error
				iterErr = tx.Ascend(c.ConcernStateKey(), func(key, value string) bool {
					var (
						groupCode, Id int64
						notify        *ConcernLiveNotify
					)
					groupCode, Id, iterErr = c.ParseConcernStateKey(key)
					if iterErr != nil {
						return false
					}
					if Id != event.RoomId || !concern.FromString(value).ContainAll(concern.DouyuLive) {
						return true
					}
					if event.ShowStatus == ShowStatus_Living {
						logger.WithField("roomid", event.RoomId).
							WithField("name", event.Nickname).
							Debugf("living notify")
						notify = NewConcernLiveNotify(groupCode, event)
					} else {
						logger.WithField("roomid", event.RoomId).
							WithField("name", event.Nickname).
							Debugf("noliving notify")
						notify = NewConcernLiveNotify(groupCode, event)
					}
					if notify != nil {
						c.notify <- notify
					}
					return true
				})
				return iterErr
			})
		}
	}
}

func (c *Concern) FreshIndex() {
	db, err := localdb.GetClient()
	if err != nil {
		return
	}
	for _, groupInfo := range miraiBot.Instance.GroupList {
		db.CreateIndex(c.ConcernStateKey(groupInfo.Code), c.ConcernStateKey(groupInfo.Code, "*"), buntdb.IndexString)
	}
}

func (c *Concern) ConcernStateKey(keys ...interface{}) string {
	return localdb.DouyuConcernStateKey(keys...)
}
func (c *Concern) CurrentLiveKey(keys ...interface{}) string {
	return localdb.DouyuCurrentLiveKey(keys...)
}
func (c *Concern) FreshKey(keys ...interface{}) string {
	return localdb.DouyuFreshKey(keys...)
}
func (c *Concern) ParseConcernStateKey(key string) (int64, int64, error) {
	return localdb.ParseConcernStateKey(key)
}
func (c *Concern) ParseCurrentLiveKey(key string) (int64, error) {
	return localdb.ParseCurrentLiveKey(key)
}

func (c *Concern) findRoom(id int64, load bool) (*LiveInfo, error) {
	db, err := localdb.GetClient()
	if err != nil {
		return nil, err
	}

	var liveInfo *LiveInfo
	if load {
		betardResp, err := Betard(id)
		if err != nil {
			return nil, err
		}
		liveInfo = &LiveInfo{
			Nickname:   betardResp.GetRoom().GetNickname(),
			RoomId:     betardResp.GetRoom().GetRoomId(),
			RoomName:   betardResp.GetRoom().GetRoomName(),
			RoomUrl:    betardResp.GetRoom().GetRoomUrl(),
			ShowStatus: betardResp.GetRoom().GetShowStatus(),
			VideoLoop:  betardResp.GetRoom().GetVideoLoop(),
			Avatar:     betardResp.GetRoom().GetAvatar(),
		}
		db.Update(func(tx *buntdb.Tx) error {
			tx.Set(c.CurrentLiveKey(id), liveInfo.ToString(), nil)
			return nil
		})
	}
	if liveInfo != nil {
		return liveInfo, nil
	}
	liveInfo = &LiveInfo{}
	err = db.View(func(tx *buntdb.Tx) error {
		val, err := tx.Get(c.CurrentLiveKey(id))
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

func NewConcernLiveNotify(groupCode int64, l *LiveInfo) *ConcernLiveNotify {
	if l == nil {
		return nil
	}
	return &ConcernLiveNotify{
		*l,
		groupCode,
	}
}

func NewConcern(notify chan<- concern.Notify) *Concern {
	c := &Concern{
		eventChan: make(chan ConcernEvent, 500),
		notify:    notify,
		stop:      make(chan interface{}),
	}
	return c
}
