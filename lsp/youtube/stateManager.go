package youtube

import (
	localdb "github.com/Sora233/Sora233-MiraiGo/lsp/buntdb"
	"github.com/Sora233/Sora233-MiraiGo/lsp/concern_manager"
	"github.com/tidwall/buntdb"
)

// TODO
type StateManager struct {
	*concern_manager.StateManager
	*extraKey
}

func (s *StateManager) AddChannel(channelId string) {

}

func (s *StateManager) AddInfo(info *Info) {

}

func (s *StateManager) GetInfo(channelId string) (*Info, error) {
	return nil, nil
}

func (s *StateManager) FreshIndex() {
	s.StateManager.FreshIndex()
	db, err := localdb.GetClient()
	if err == nil {
		db.CreateIndex(s.GroupConcernStateKey(), s.GroupConcernStateKey("*"), buntdb.IndexString)
		db.CreateIndex(s.FreshKey(), s.FreshKey("*"), buntdb.IndexString)
		db.CreateIndex(s.UserInfoKey(), s.UserInfoKey("*", buntdb.IndexString))
		db.CreateIndex(s.ConcernStateKey(), s.ConcernStateKey("*"), buntdb.IndexBinary)
	}
}

func NewStateManager() *StateManager {
	sm := new(StateManager)
	sm.extraKey = NewExtraKey()
	sm.StateManager = concern_manager.NewStateManager(NewKeySet())
	return sm
}
