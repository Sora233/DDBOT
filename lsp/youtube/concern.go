package youtube

import (
	"errors"
	"fmt"
	"github.com/Logiase/MiraiGo-Template/utils"
	"github.com/Sora233/DDBOT/concern"
	"github.com/Sora233/DDBOT/lsp/concern_manager"
)

var logger = utils.GetModuleLogger("youtube")

type Concern struct {
	*StateManager

	eventChan chan ConcernEvent
	notify    chan<- concern.Notify
}

func (c *Concern) Add(groupCode int64, id string, ctype concern.Type) (info *Info, err error) {
	log := logger.WithField("group_code", groupCode)

	err = c.StateManager.CheckGroupConcern(groupCode, id, ctype)
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
	_, err = c.StateManager.AddGroupConcern(groupCode, id, ctype)
	if err != nil {
		return nil, err
	}
	return NewInfo(videoInfo), nil
}

func (c *Concern) ListWatching(groupCode int64, ctype concern.Type) ([]*UserInfo, error) {
	log := logger.WithField("group_code", groupCode)

	ids, _, err := c.StateManager.ListByGroup(groupCode, func(id interface{}, p concern.Type) bool {
		return p.ContainAny(ctype)
	})
	if err != nil {
		return nil, err
	}
	var result = make([]*UserInfo, 0)
	for _, id := range ids {
		info, err := c.findOrLoad(id.(string))
		if err != nil {
			log.WithField("id", id.(string)).Errorf("findInfo failed %v", err)
			continue
		}
		result = append(result, NewUserInfo(info.ChannelId, info.ChannelName))
	}
	return result, nil
}

func (c *Concern) Start() {
	if err := c.StateManager.Start(); err != nil {
		logger.Errorf("state manager start err %v", err)
	}

	go c.notifyLoop()
	go c.EmitFreshCore("youtube", func(ctype concern.Type, id interface{}) error {
		if ctype.ContainAny(concern.YoutubeLive | concern.YoutubeVideo) {
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
		event := ievent.(*VideoInfo)
		log := logger.WithField("channel_id", event.ChannelId).
			WithField("video_id", event.VideoId).
			WithField("video_type", event.VideoType.String()).
			WithField("video_title", event.VideoTitle).
			WithField("video_status", event.VideoStatus.String())
		if prev, err := c.StateManager.GetVideo(event.ChannelId, event.VideoId); err == nil {
			if prev.VideoStatus == event.VideoStatus && prev.VideoType == event.VideoType {
				log.Debugf("duplicate event")
				continue
			}
		}
		log.Debugf("debug event")
		if err := c.StateManager.AddVideo(event); err != nil {
			log.Errorf("add video err %v", err)
		}
		groups, _, idTypes, err := c.StateManager.List(func(groupCode int64, id interface{}, p concern.Type) bool {
			return id.(string) == event.ChannelId && p.ContainAny(concern.YoutubeLive|concern.YoutubeVideo)
		})
		if err != nil {
			logger.Errorf("list id failed %v", err)
			continue
		}
		for index, groupCode := range groups {
			var doNotify bool
			ctype := idTypes[index]
			if ctype.ContainAny(concern.YoutubeLive) && event.IsLive() {
				notify := NewConcernNotify(groupCode, event)
				c.notify <- notify
				doNotify = true
			}
			if ctype.ContainAny(concern.YoutubeVideo) && event.IsVideo() {
				notify := NewConcernNotify(groupCode, event)
				c.notify <- notify
				doNotify = true
			}
			if doNotify {
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
			if newV.IsLive() {
				c.eventChan <- newV
				log.Debugf("first load live notify")
			}
		}
	} else {
		var notifyCount = 0
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
				if notifyCount == 0 {
					c.eventChan <- newV
					log.Debugf("new video notify")
					notifyCount += 1
					// notify video most once
				}
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

func (c *Concern) findOrLoad(channelId string) (*Info, error) {
	info, _ := c.findInfo(channelId, false)
	if info == nil {
		return c.findInfo(channelId, true)
	} else {
		return info, nil
	}
}

func NewConcern(notify chan<- concern.Notify) *Concern {
	return &Concern{
		StateManager: NewStateManager(),
		notify:       notify,
		eventChan:    make(chan ConcernEvent, 500),
	}
}
