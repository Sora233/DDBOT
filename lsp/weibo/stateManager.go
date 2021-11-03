package weibo

import "github.com/Sora233/DDBOT/lsp/concern"

type StateManager struct {
	*concern.StateManager
}

func NewStateManager(notify chan<- concern.Notify) *StateManager {
	return &StateManager{
		concern.NewStateManagerWithInt64ID(Site, notify),
	}
}
