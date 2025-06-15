package douyu

import (
	"errors"
	"time"

	localdb "github.com/Sora233/DDBOT/v2/lsp/buntdb"
	"github.com/Sora233/DDBOT/v2/lsp/concern"
)

type StateManager struct {
	*concern.StateManager
	*extraKey
}

func (c *StateManager) GetLiveInfo(id int64) (*LiveInfo, error) {
	var liveInfo = &LiveInfo{}
	err := c.GetJson(c.CurrentLiveKey(id), liveInfo)
	if err != nil {
		return nil, err
	}
	return liveInfo, nil
}

func (c *StateManager) AddLiveInfo(liveInfo *LiveInfo) error {
	if liveInfo == nil {
		return errors.New("nil LiveInfo")
	}
	return c.SetJson(c.CurrentLiveKey(liveInfo.RoomId), liveInfo, localdb.SetExpireOpt(time.Hour*24*7))
}

func (c *StateManager) DeleteLiveInfo(id int64) error {
	_, err := c.Delete(c.CurrentLiveKey(id))
	return err
}

func (c *StateManager) GetGroupConcernConfig(groupCode uint32, id interface{}) (concernConfig concern.IConfig) {
	return NewGroupConcernConfig(c.StateManager.GetGroupConcernConfig(groupCode, id))
}

func NewStateManager(notify chan<- concern.Notify) *StateManager {
	sm := &StateManager{}
	sm.extraKey = NewExtraKey()
	sm.StateManager = concern.NewStateManagerWithCustomKey(Site, NewKeySet(), notify)
	return sm
}
