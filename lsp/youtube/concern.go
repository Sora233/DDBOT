package youtube

import (
	"errors"
	"fmt"
	"github.com/Logiase/MiraiGo-Template/utils"
	"github.com/Sora233/DDBOT/lsp/concern"
	"github.com/Sora233/DDBOT/lsp/concern_type"
	"github.com/Sora233/DDBOT/lsp/msg"
	localutils "github.com/Sora233/DDBOT/utils"
	"runtime"
	"sync"
)

var logger = utils.GetModuleLogger("youtube")

type Concern struct {
	*StateManager

	eventChan chan concernEvent
	notify    chan<- concern.Notify
	stop      chan interface{}
	wg        sync.WaitGroup
}

func (c *Concern) Site() string {
	return "youtube"
}

func (c *Concern) ParseId(s string) (interface{}, error) {
	return s, nil
}

func (c *Concern) GetStateManager() concern.IStateManager {
	return c.StateManager
}

func (c *Concern) Add(ctx msg.IMsgCtx, groupCode int64, _id interface{}, ctype concern_type.Type) (concern.IdentityInfo, error) {
	id := _id.(string)
	log := logger.WithFields(localutils.GroupLogFields(groupCode)).WithField("id", id)

	err := c.StateManager.CheckGroupConcern(groupCode, id, ctype)
	if err != nil {
		if err == concern.ErrAlreadyExists {
			return nil, errors.New("已经watch过了")
		}
		return nil, err
	}
	info, err := c.FindOrLoad(id)
	if err != nil {
		log.Errorf("FindOrLoad error %v", err)
		return nil, fmt.Errorf("查询channel信息失败 %v - %v", id, err)
	}
	_, err = c.StateManager.AddGroupConcern(groupCode, id, ctype)
	if err != nil {
		return nil, err
	}
	return concern.NewIdentity(info.ChannelId, info.ChannelName), nil
}

func (c *Concern) Remove(ctx msg.IMsgCtx, groupCode int64, _id interface{}, ctype concern_type.Type) (concern.IdentityInfo, error) {
	id := _id.(string)
	identity, _ := c.Get(id)
	_, err := c.StateManager.RemoveGroupConcern(groupCode, id, ctype)
	return identity, err
}

func (c *Concern) Get(id interface{}) (concern.IdentityInfo, error) {
	info, err := c.FindInfo(id.(string), false)
	if err != nil {
		return nil, err
	}
	return concern.NewIdentity(info.ChannelId, info.ChannelName), nil
}

func (c *Concern) List(groupCode int64, ctype concern_type.Type) ([]concern.IdentityInfo, []concern_type.Type, error) {
	log := logger.WithFields(localutils.GroupLogFields(groupCode))

	_, ids, ctypes, err := c.StateManager.List(func(_groupCode int64, id interface{}, p concern_type.Type) bool {
		return groupCode == _groupCode && p.ContainAny(ctype)
	})
	if err != nil {
		return nil, nil, err
	}
	var result = make([]concern.IdentityInfo, 0, len(ids))
	var resultTypes = make([]concern_type.Type, 0, len(ids))
	for index, id := range ids {
		info, err := c.FindOrLoad(id.(string))
		if err != nil {
			log.WithField("id", id.(string)).Errorf("FindInfo failed %v", err)
			continue
		}
		result = append(result, concern.NewIdentity(info.ChannelId, info.ChannelName))
		resultTypes = append(resultTypes, ctypes[index])
	}
	return result, resultTypes, nil
}

func (c *Concern) Stop() {
	logger.Trace("正在停止youtube StateManager")
	c.StateManager.Stop()
	logger.Trace("youtube StateManager已停止")
	if c.stop != nil {
		close(c.stop)
	}
	close(c.eventChan)
	logger.Trace("正在停止youtube concern")
	c.wg.Wait()
	logger.Trace("youtube concern已停止")
}

func (c *Concern) Start() error {
	if err := c.StateManager.Start(); err != nil {
		logger.Errorf("state manager start err %v", err)
	}

	if runtime.NumCPU() >= 3 {
		for i := 0; i < 3; i++ {
			go c.notifyLoop()
		}
	} else {
		go c.notifyLoop()
	}

	go c.EmitFreshCore("youtube", func(ctype concern_type.Type, id interface{}) error {
		if ctype.ContainAny(Live.Add(Video)) {
			channelId, ok := id.(string)
			if !ok {
				return errors.New("canst fresh id to string failed")
			}
			c.freshInfo(channelId)
		}
		return nil
	})
	return nil
}

func (c *Concern) notifyLoop() {
	c.wg.Add(1)
	defer c.wg.Done()
	for ievent := range c.eventChan {
		event := ievent.(*VideoInfo)
		log := event.Logger()
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
		groups, _, idTypes, err := c.StateManager.List(func(groupCode int64, id interface{}, p concern_type.Type) bool {
			return id.(string) == event.ChannelId && p.ContainAny(Live.Add(Video))
		})
		if err != nil {
			logger.Errorf("list id failed %v", err)
			continue
		}
		for index, groupCode := range groups {
			var doNotify bool
			ctype := idTypes[index]
			if ctype.ContainAny(Live) && event.IsLive() {
				notify := NewConcernNotify(groupCode, event)
				c.notify <- notify
				doNotify = true
			}
			if ctype.ContainAny(Video) && event.IsVideo() {
				notify := NewConcernNotify(groupCode, event)
				c.notify <- notify
				doNotify = true
			}
			if doNotify {
				if event.IsVideo() {
					log.WithFields(localutils.GroupLogFields(groupCode)).Debugf("video notify")
				} else if event.IsLive() {
					if event.IsWaiting() {
						log.WithFields(localutils.GroupLogFields(groupCode)).Debugf("live waiting notify")
					} else if event.IsLiving() {
						log.WithFields(localutils.GroupLogFields(groupCode)).Debugf("living notify")
					}
				}
			}
		}
	}
}

func (c *Concern) freshInfo(channelId string) {
	log := logger.WithField("channel_id", channelId)
	oldInfo, _ := c.FindInfo(channelId, false)
	newInfo, err := c.FindInfo(channelId, true)
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
					if newV.IsVideo() && oldV.IsLive() {
						// 应该是下播了吧？
						log.Debug("offline notify")
						c.eventChan <- newV
					}
					if newV.IsLive() && oldV.IsLive() {
						if newV.IsWaiting() && oldV.IsWaiting() && newV.VideoTimestamp != oldV.VideoTimestamp {
							// live time changed, notify
							c.eventChan <- newV
							log.Debugf("live time change notify")
						} else if newV.IsLiving() && oldV.IsWaiting() {
							// live begin
							newV.LiveStatusChanged = true
							c.eventChan <- newV
							log.Debugf("live begin notify")
						} else if newV.VideoTitle != oldV.VideoTitle {
							newV.LiveTitleChanged = true
							c.eventChan <- newV
						}
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

func (c *Concern) FindInfo(channelId string, load bool) (*Info, error) {
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

func (c *Concern) FindOrLoad(channelId string) (*Info, error) {
	info, _ := c.FindInfo(channelId, false)
	if info == nil {
		return c.FindInfo(channelId, true)
	} else {
		return info, nil
	}
}

func NewConcern(notify chan<- concern.Notify) *Concern {
	return &Concern{
		StateManager: NewStateManager(),
		notify:       notify,
		stop:         make(chan interface{}),
		eventChan:    make(chan concernEvent, 500),
	}
}
