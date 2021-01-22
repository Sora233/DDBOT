package youtube

import (
	"errors"
	"github.com/Logiase/MiraiGo-Template/utils"
	"github.com/Sora233/Sora233-MiraiGo/concern"
)

var logger = utils.GetModuleLogger("youtube")

type Concern struct {
	*StateManager

	eventChan chan ConcernEvent
	notify    chan<- concern.Notify
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
						} else if newV.IsLiving() && oldV.IsWaiting() {
							// live begin
							c.eventChan <- newV
						}
						// any other case?
					}
				}
			}
			if !found {
				// new video
				c.eventChan <- newV
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
		info = &Info{VideoInfo: vi}
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
