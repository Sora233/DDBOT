package youtube

import (
	"errors"
	"fmt"
	"github.com/Sora233/DDBOT/lsp/concern"
	"github.com/Sora233/DDBOT/lsp/concern_type"
	"github.com/Sora233/DDBOT/lsp/mmsg"
	localutils "github.com/Sora233/DDBOT/utils"
	"github.com/Sora233/MiraiGo-Template/utils"
)

var logger = utils.GetModuleLogger("youtube")

type Concern struct {
	*StateManager
}

func (c *Concern) Site() string {
	return "youtube"
}

func (c *Concern) Types() []concern_type.Type {
	return []concern_type.Type{Live, Video}
}

func (c *Concern) ParseId(s string) (interface{}, error) {
	return s, nil
}

func (c *Concern) GetStateManager() concern.IStateManager {
	return c.StateManager
}

func (c *Concern) Add(ctx mmsg.IMsgCtx, groupCode int64, _id interface{}, ctype concern_type.Type) (concern.IdentityInfo, error) {
	id := _id.(string)
	log := logger.WithFields(localutils.GroupLogFields(groupCode)).WithField("id", id)

	err := c.StateManager.CheckGroupConcern(groupCode, id, ctype)
	if err != nil {
		return nil, err
	}
	info, err := c.FindOrLoad(id)
	if err != nil {
		log.Errorf("FindOrLoad error %v", err)
		return nil, fmt.Errorf("查询channel信息失败 %v - %v", id, err)
	}
	for _, v := range info.VideoInfo {
		c.StateManager.AddVideo(v)
	}
	_, err = c.StateManager.AddGroupConcern(groupCode, id, ctype)
	if err != nil {
		return nil, err
	}
	return concern.NewIdentity(info.ChannelId, info.ChannelName), nil
}

func (c *Concern) Remove(ctx mmsg.IMsgCtx, groupCode int64, _id interface{}, ctype concern_type.Type) (concern.IdentityInfo, error) {
	id := _id.(string)
	identity, _ := c.Get(id)
	_, err := c.StateManager.RemoveGroupConcern(groupCode, id, ctype)
	if identity == nil {
		identity = concern.NewIdentity(_id, "unknown")
	}
	return identity, err
}

func (c *Concern) Get(id interface{}) (concern.IdentityInfo, error) {
	info, err := c.FindInfo(id.(string), false)
	if err != nil {
		return nil, err
	}
	return concern.NewIdentity(info.ChannelId, info.ChannelName), nil
}

func (c *Concern) Stop() {
	logger.Trace("正在停止youtube concern")
	logger.Trace("正在停止youtube StateManager")
	c.StateManager.Stop()
	logger.Trace("youtube StateManager已停止")
	logger.Trace("youtube concern已停止")
}

func (c *Concern) Start() error {
	c.UseEmitQueue()
	c.UseFreshFunc(c.fresh())
	c.UseNotifyGeneratorFunc(c.notifyGenerator())
	return c.StateManager.Start()
}

func (c *Concern) fresh() concern.FreshFunc {
	return c.EmitQueueFresher(func(ctype concern_type.Type, id interface{}) ([]concern.Event, error) {
		if ctype.ContainAny(Live.Add(Video)) {
			channelId, ok := id.(string)
			if !ok {
				return nil, errors.New("canst fresh id to string failed")
			}
			infos, err := c.freshInfo(channelId)
			if err != nil {
				return nil, err
			}
			var result []concern.Event
			for _, event := range infos {
				prev, getErr := c.StateManager.GetVideo(event.ChannelId, event.VideoId)
				if err := c.StateManager.AddVideo(event); err != nil {
					event.Logger().Errorf("add video err %v", err)
				}
				if getErr == nil {
					if prev.VideoStatus == event.VideoStatus && prev.VideoType == event.VideoType &&
						prev.VideoTimestamp == event.VideoTimestamp && prev.VideoTitle == event.VideoTitle {
						continue
					}
				}
				result = append(result, event)
			}
			return result, nil
		}
		return nil, fmt.Errorf("unknown concern_type %v", ctype.String())
	})
}

func (c *Concern) notifyGenerator() concern.NotifyGeneratorFunc {
	return func(groupCode int64, ievent concern.Event) []concern.Notify {
		switch event := ievent.(type) {
		case *VideoInfo:
			log := event.Logger()
			if event.IsVideo() {
				log.WithFields(localutils.GroupLogFields(groupCode)).Debugf("video notify")
			} else if event.IsLive() {
				if event.IsWaiting() {
					log.WithFields(localutils.GroupLogFields(groupCode)).Debugf("live waiting notify")
				} else if event.IsLiving() {
					log.WithFields(localutils.GroupLogFields(groupCode)).Debugf("living notify")
				}
			}
			return []concern.Notify{NewConcernNotify(groupCode, event)}
		default:
			logger.Errorf("unknown EventType %+v", event)
			return nil
		}
	}
}

func (c *Concern) freshInfo(channelId string) (result []*VideoInfo, err error) {
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
				if newV.IsLiving() {
					newV.liveStatusChanged = true
				}
				result = append(result, newV)
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
						result = append(result, newV)
					}
					if newV.IsLive() && oldV.IsLive() {
						if newV.IsWaiting() && oldV.IsWaiting() && newV.VideoTimestamp != oldV.VideoTimestamp {
							// live time changed, notify
							result = append(result, newV)
						} else if newV.IsLiving() && oldV.IsWaiting() {
							// live begin
							newV.liveStatusChanged = true
							result = append(result, newV)
						} else if newV.VideoTitle != oldV.VideoTitle {
							newV.liveTitleChanged = true
							result = append(result, newV)
						}
					}
				}
			}
			if !found {
				// new video
				if notifyCount == 0 {
					result = append(result, newV)
					notifyCount += 1
					// notify video most once
				}
			}
		}
	}
	return
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
		StateManager: NewStateManager(notify),
	}
}
