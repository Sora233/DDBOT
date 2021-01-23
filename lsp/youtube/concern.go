package youtube

import (
	"errors"
	"fmt"
	"github.com/Logiase/MiraiGo-Template/utils"
	"github.com/Sora233/Sora233-MiraiGo/concern"
	"github.com/Sora233/Sora233-MiraiGo/lsp/concern_manager"
)

var logger = utils.GetModuleLogger("youtube")

type Concern struct {
	*StateManager

	eventChan chan ConcernEvent
	notify    chan<- concern.Notify
}

func (c *Concern) Add(groupCode int64, id string, ctype concern.Type) (info *Info, err error) {
	log := logger.WithField("group_code", groupCode)

	err = c.StateManager.CheckConcern(id, ctype)
	if err != nil {
		if err == concern_manager.ErrAlreadyExists {
			return nil, errors.New("已经watch过了")
		}
		return nil, err
	}
	videoInfo, err := XFetchInfo(id)
	if err != nil {
		log.WithField("id", id).Errorf("XFetchInfo failed %v", err)
		return nil, fmt.Errorf("查询channel信息失败 %v - %v", id, err)
	}
	err = c.StateManager.AddGroupConcern(groupCode, id, ctype)
	if err != nil {
		return nil, err
	}
	return NewInfo(videoInfo), nil
}

func (c *Concern) ListLiving(groupCode int64, all bool) ([]*ConcernNotify, error) {
	log := logger.WithField("group_code", groupCode).WithField("all", all)
	var result []*ConcernNotify

	ids, _, err := c.StateManager.ListByGroup(groupCode, func(id interface{}, p concern.Type) bool {
		return p.ContainAny(concern.Youtube)
	})
	if err != nil {
		return nil, err
	}
	if len(ids) != 0 {
		result = make([]*ConcernNotify, 0)
	}
	for _, id := range ids {
		info, err := c.findInfo(id.(string), false)
		if err != nil {
			log.WithField("id", id.(string)).Errorf("findInfo failed %v", err)
			continue
		}
		if len(info.VideoInfo) > 0 {
			result = append(result, NewConcernNotify(groupCode, info.VideoInfo[0]))
		}
	}
	return result, nil
}

func (c *Concern) Start() {
	if err := c.StateManager.Start(); err != nil {
		logger.Errorf("state manager start err %v", err)
	}

	go c.notifyLoop()
	go c.EmitFreshCore("youtube", func(ctype concern.Type, id interface{}) error {
		if ctype.ContainAll(concern.Youtube) {
			channelId, ok := id.(string)
			if !ok {
				return errors.New("canst fresh id to string failed")
			}
			c.freshInfo(channelId)
		}
		return nil
	})
}

func (c *Concern) notifyLoop() {
	for ievent := range c.eventChan {
		switch ievent.Type() {
		case Video:
			event := ievent.(*VideoInfo)
			log := logger.WithField("channel_id", event.ChannelId).
				WithField("video_id", event.VideoId).
				WithField("video_type", event.VideoType.String()).
				WithField("video_title", event.VideoTitle).
				WithField("video_status", event.VideoStatus.String())
			log.Debugf("debug event")
			groups, _, _, err := c.StateManager.List(func(groupCode int64, id interface{}, p concern.Type) bool {
				return id.(string) == event.ChannelId && p.ContainAny(concern.Youtube)
			})
			if err != nil {
				logger.Errorf("list id failed %v", err)
				continue
			}
			for _, groupCode := range groups {
				notify := NewConcernNotify(groupCode, event)
				c.notify <- notify
				if event.IsVideo() {
					log.Debugf("video notify")
				} else if event.IsLive() {
					if event.IsWaiting() {
						log.Debugf("live waiting notify")
					} else if event.IsLiving() {
						log.Debugf("living notify")
					}
				}
			}
		}
	}
}

func (c *Concern) freshInfo(channelId string) {
	log := logger.WithField("channel_id", channelId)
	oldInfo, _ := c.findInfo(channelId, false)
	newInfo, err := c.findInfo(channelId, true)
	if err != nil {
		log.Errorf("load newInfo failed %v", err)
		return
	}
	if oldInfo == nil {
		// first load, just notify if living
		for _, newV := range newInfo.VideoInfo {
			if newV.IsLive() && newV.IsLiving() {
				c.eventChan <- newV
				log.Debugf("first load notify")
			}
		}
	} else {
		for _, newV := range newInfo.VideoInfo {
			var found bool
			for _, oldV := range oldInfo.VideoInfo {
				if newV.VideoId == oldV.VideoId {
					found = true
					if newV.IsLive() && oldV.IsLive() {
						if newV.IsWaiting() && oldV.IsWaiting() && newV.VideoTimestamp != oldV.VideoTimestamp {
							// live time changed, notify
							c.eventChan <- newV
							log.Debugf("live time change notify")
						} else if newV.IsLiving() && oldV.IsWaiting() {
							// live begin
							c.eventChan <- newV
							log.Debugf("live begin notify")
						}
						// any other case?
					}
				}
			}
			if !found {
				// new video
				c.eventChan <- newV
				log.Debugf("new video notify")
			}
		}
	}
}

func (c *Concern) findInfo(channelId string, load bool) (*Info, error) {
	var info *Info
	if load {
		vi, err := XFetchInfo(channelId)
		if err != nil {
			return nil, err
		}
		info = NewInfo(vi)
		c.StateManager.AddInfo(info)
	}

	if info != nil {
		return info, nil
	}
	return c.GetInfo(channelId)
}

func NewConcern(notify chan<- concern.Notify) *Concern {
	return &Concern{
		StateManager: NewStateManager(),
		notify:       notify,
		eventChan:    make(chan ConcernEvent, 500),
	}
}
