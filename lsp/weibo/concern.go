package weibo

import (
	"github.com/Logiase/MiraiGo-Template/utils"
	"github.com/Sora233/DDBOT/lsp/concern"
	"github.com/Sora233/DDBOT/lsp/concern_type"
	"github.com/Sora233/DDBOT/lsp/mmsg"
	"strconv"
)

var logger = utils.GetModuleLogger("weibo-concern")

type Concern struct {
	*StateManager
}

func (c *Concern) Start() error {
	c.StateManager.UseFreshFunc(c.EmitQueueFresher(func(p concern_type.Type, id interface{}) ([]concern.Event, error) {
		// TODO
		return nil, nil
	}))
	c.StateManager.UseNotifyGeneratorFunc(func(groupCode int64, event concern.Event) []concern.Notify {
		// TODO
		return nil
	})
	return c.StateManager.Start()
}

func (c *Concern) Stop() {
	logger.Tracef("正在停止%v concern", Site)
	logger.Tracef("正在停止%v StateManager", Site)
	c.StateManager.Stop()
	logger.Tracef("%v StateManager已停止", Site)
	logger.Tracef("%v concern已停止", Site)
}

func (c *Concern) Site() string {
	return Site
}

func (c *Concern) ParseId(s string) (interface{}, error) {
	return strconv.ParseInt(s, 10, 64)
}

func (c *Concern) Add(ctx mmsg.IMsgCtx, groupCode int64, id interface{}, ctype concern_type.Type) (concern.IdentityInfo, error) {
	panic("implement me")
}

func (c *Concern) Remove(ctx mmsg.IMsgCtx, groupCode int64, id interface{}, ctype concern_type.Type) (concern.IdentityInfo, error) {
	panic("implement me")
}

func (c *Concern) List(groupCode int64, ctype concern_type.Type) ([]concern.IdentityInfo, []concern_type.Type, error) {
	panic("implement me")
}

func (c *Concern) Get(id interface{}) (concern.IdentityInfo, error) {
	panic("implement me")
}

func (c *Concern) GetStateManager() concern.IStateManager {
	return c.StateManager
}

func NewConcern(notify chan<- concern.Notify) *Concern {
	c := &Concern{
		StateManager: NewStateManager(notify),
	}
	c.UseEmitQueue()
	return c
}
