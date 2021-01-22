package youtube

import (
	"errors"
	"github.com/Logiase/MiraiGo-Template/utils"
	"github.com/Sora233/Sora233-MiraiGo/concern"
)

var logger = utils.GetModuleLogger("youtube")

type Concern struct {
	*StateManager

	notify chan<- concern.Notify
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

}

func (c *Concern) freshInfo(channelId string) {
	// TODO
}

func (c *Concern) findInfo(channelId string, load bool) (*Info, error) {
	var info *Info
	if load {
		vi, err := Video(channelId)
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
	emitChan := make(chan interface{})
	return &Concern{
		StateManager: NewStateManager(emitChan),
		notify:       notify,
	}
}
