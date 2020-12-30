package bilibili

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Logiase/MiraiGo-Template/utils"
	"github.com/Sora233/Sora233-MiraiGo/concern"
	localdb "github.com/Sora233/Sora233-MiraiGo/lsp/buntdb"
	"github.com/Sora233/Sora233-MiraiGo/lsp/concern_manager"
	"github.com/forestgiant/sliceutil"
	"github.com/tidwall/buntdb"
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

func NewConcernLiveNotify(groupCode int64, liveInfo *LiveInfo) *ConcernLiveNotify {
	if liveInfo == nil {
		return nil
	}
	return &ConcernLiveNotify{
		GroupCode: groupCode,
		LiveInfo:  *liveInfo,
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

func NewUserInfo(mid, roomId int64, name, url string) *UserInfo {
	return &UserInfo{
		Mid:     mid,
		RoomId:  roomId,
		Name:    name,
		RoomUrl: url,
	}
}

type LiveInfo struct {
	UserInfo
	Status    LiveStatus `json:"status"`
	LiveTitle string     `json:"live_title"`
	Cover     string     `json:"cover"`
}

func NewLiveInfo(userInfo *UserInfo, liveTitle string, cover string, status LiveStatus) *LiveInfo {
	return &LiveInfo{
		UserInfo:  *userInfo,
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

type Concern struct {
	*StateManager

	eventChan chan ConcernEvent
	notify    chan<- concern.Notify
	stopped   bool
	stop      chan interface{}
}

func NewConcern(notify chan<- concern.Notify) *Concern {
	c := &Concern{
		StateManager: NewStateManager(),
		eventChan:    make(chan ConcernEvent, 500),
		notify:       notify,
		stop:         make(chan interface{}),
	}
	return c
}

func (c *Concern) Start() {
	db, err := localdb.GetClient()
	if err == nil {
		db.CreateIndex(c.ConcernStateKey(), c.ConcernStateKey("*"), buntdb.IndexString)
		db.CreateIndex(c.CurrentLiveKey(), c.CurrentLiveKey("*"), buntdb.IndexString)
		db.CreateIndex(c.FreshKey(), c.FreshKey("*"), buntdb.IndexString)
		db.CreateIndex(c.UserInfoKey(), c.UserInfoKey("*", buntdb.IndexString))
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

func (c *Concern) Stop() {
	c.stopped = true
	if c.stop != nil {
		close(c.stop)
	}
}

func (c *Concern) Add(groupCode int64, mid int64, ctype concern.Type) (*UserInfo, error) {
	var err error
	log := logger.WithField("GroupCode", groupCode)

	err = c.StateManager.Check(groupCode, mid, ctype)
	if err != nil {
		if err == concern_manager.ErrAlreadyExists {
			return nil, errors.New("已经watch过了")
		}
		return nil, err
	}

	infoResp, err := XSpaceAccInfo(mid)
	if err != nil {
		log.WithField("mid", mid).Error(err)
		return nil, fmt.Errorf("查询用户信息失败 %v - %v", mid, err)
	}
	if infoResp.Code != 0 {
		log.WithField("mid", mid).WithField("code", infoResp.Code).Errorf(infoResp.Message)
		return nil, fmt.Errorf("查询用户信息失败 %v - %v %v", mid, infoResp.Code, infoResp.Message)
	}

	name := infoResp.GetData().GetName()

	if sliceutil.Contains([]int64{491474049}, mid) {
		return nil, fmt.Errorf("用户 %v 禁止watch", name)
	}
	if ctype.ContainAll(concern.BibiliLive) {
		if RoomStatus(infoResp.GetData().GetLiveRoom().GetRoomStatus()) == RoomStatus_NonExist {
			return nil, fmt.Errorf("用户 %v 暂未开通直播间", name)
		}
	}

	err = c.StateManager.Add(groupCode, mid, ctype)
	if err != nil {
		return nil, err
	}

	userInfo := NewUserInfo(
		mid,
		infoResp.GetData().GetLiveRoom().GetRoomid(),
		infoResp.GetData().GetName(),
		infoResp.GetData().GetLiveRoom().GetUrl(),
	)

	_ = c.StateManager.AddUserInfo(userInfo)
	return userInfo, nil
}

func (c *Concern) ListLiving(groupCode int64, all bool) ([]*ConcernLiveNotify, error) {
	log := logger.WithField("group_code", groupCode).WithField("all", all)
	var result []*ConcernLiveNotify

	mids, _, err := c.StateManager.ListIdByGroup(groupCode, func(id int64, p concern.Type) bool {
		return p.ContainAny(concern.BibiliLive)
	})
	if err != nil {
		return nil, err
	}
	for _, mid := range mids {
		liveInfo, err := c.StateManager.GetLiveInfo(mid)
		if err != nil {
			log.WithField("mid", mid).Errorf("get LiveInfo err %v", err)
			continue
		}
		if all || liveInfo.Status == LiveStatus_Living {
			result = append(result, NewConcernLiveNotify(groupCode, liveInfo))
		}
	}
	return result, nil
}

func (c *Concern) notifyLoop() {
	for ievent := range c.eventChan {
		if c.stopped {
			return
		}

		switch ievent.Type() {
		case Live:
			event := (ievent).(*LiveInfo)
			log := logger.WithField("mid", event).
				WithField("name", event.Name).
				WithField("roomid", event.RoomId).
				WithField("title", event.LiveTitle).
				WithField("status", event.Status.String())
			log.Debugf("debug event")

			groups, _, _, err := c.StateManager.ListId(func(groupCode int64, id int64, p concern.Type) bool {
				return id == event.Mid && p.ContainAny(concern.BibiliLive)
			})
			if err != nil {
				log.Errorf("list id failed %v", err)
				continue
			}

			for _, groupCode := range groups {
				notify := NewConcernLiveNotify(groupCode, event)
				c.notify <- notify
				if event.Status == LiveStatus_Living {
					log.Debug("living notify")
				} else if event.Status == LiveStatus_NoLiving {
					log.Debug("noliving notify")
				} else {
					log.Error("unknown live status")
				}
			}
		case News:
			// TODO
			logger.Errorf("concern event news not supported")
		}

	}
}

func (c *Concern) FreshConcern() {
	_, mids, ctypes, err := c.StateManager.ListId(func(groupCode int64, id int64, p concern.Type) bool {
		return p.ContainAll(concern.BibiliLive | concern.BilibiliNews)
	})
	if err != nil {
		logger.Errorf("list id failed %v", err)
		return
	}

	mids, ctypes, err = c.StateManager.GroupTypeById(mids, ctypes)
	if err != nil {
		logger.Errorf("GroupTypeById failed %v", err)
	}

	var freshConcern []struct {
		Mid         int64
		ConcernType concern.Type
	}

	for index := range mids {
		mid := mids[index]
		ctype := ctypes[index]
		ok, err := c.StateManager.FreshCheck(mid, true)
		if err != nil {
			logger.WithField("mid", mid).Errorf("FreshCheck failed %v", err)
			continue
		}
		if ok {
			freshConcern = append(freshConcern, struct {
				Mid         int64
				ConcernType concern.Type
			}{Mid: mid, ConcernType: ctype})
		}
	}

	for _, item := range freshConcern {
		if item.ConcernType.ContainAll(concern.BibiliLive) {
			oldInfo, _ := c.findUserLiving(item.Mid, false)
			liveInfo, err := c.findUserLiving(item.Mid, true)
			if err != nil {
				logger.WithField("mid", item.Mid).Errorf("load liveinfo failed %v", err)
				continue
			}
			if oldInfo == nil || oldInfo.Status != liveInfo.Status || oldInfo.LiveTitle != liveInfo.LiveTitle {
				c.eventChan <- liveInfo
			}
		}
		if item.ConcernType.ContainAll(concern.BilibiliNews) {
			// TODO
		}
	}
}

func (c *Concern) findUser(mid int64, load bool) (*UserInfo, error) {
	if load {
		resp, err := XSpaceAccInfo(mid)
		if err != nil {
			return nil, err
		}
		if resp.Code != 0 {
			return nil, fmt.Errorf("code:%v %v", resp.Code, resp.Message)
		}
		newUserInfo := NewUserInfo(mid,
			resp.GetData().GetLiveRoom().GetRoomid(),
			resp.GetData().GetName(),
			resp.GetData().GetLiveRoom().GetUrl(),
		)
		err = c.StateManager.AddUserInfo(newUserInfo)
		if err != nil {
			return nil, err
		}
	}
	return c.StateManager.GetUserInfo(mid)
}

func (c *Concern) findUserLiving(mid int64, load bool) (*LiveInfo, error) {
	userInfo, err := c.findUser(mid, load)
	if err != nil {
		return nil, err
	}

	var liveInfo *LiveInfo

	if load {
		roomInfo, err := GetRoomInfoOld(userInfo.Mid)
		if err != nil {
			return nil, err
		}
		if roomInfo.Code != 0 {
			return nil, fmt.Errorf("code:%v %v", roomInfo.Code, roomInfo.Message)
		}
		liveInfo = NewLiveInfo(userInfo,
			roomInfo.GetData().GetTitle(),
			roomInfo.GetData().GetCover(),
			roomInfo.GetData().GetLiveStatus(),
		)
		_ = c.StateManager.AddLiveInfo(liveInfo)
	}

	if liveInfo != nil {
		return liveInfo, nil
	}
	return c.StateManager.GetLiveInfo(mid)
}
