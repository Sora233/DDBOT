package youtube

import (
	"encoding/json"
	localdb "github.com/Sora233/Sora233-MiraiGo/lsp/buntdb"
	"github.com/Sora233/Sora233-MiraiGo/lsp/concern_manager"
	"github.com/tidwall/buntdb"
	"time"
)

type StateManager struct {
	*concern_manager.StateManager
	*extraKey
}

func (s *StateManager) AddInfo(info *Info) error {
	return s.RWTxCover(func(tx *buntdb.Tx) error {
		infoKey := s.InfoKey(info.ChannelId)
		_, _, err := tx.Set(infoKey, info.ToString(), localdb.ExpireOption(time.Hour*24))
		return err
	})
}

func (s *StateManager) GetInfo(channelId string) (info *Info, err error) {
	err = s.RTxCover(func(tx *buntdb.Tx) error {
		infoKey := s.InfoKey(channelId)
		val, err := tx.Get(infoKey)
		if err != nil {
			return err
		}
		info = new(Info)
		return json.Unmarshal([]byte(val), info)
	})
	if err != nil {
		return nil, err
	}
	return
}
func (s *StateManager) Start() error {
	db, err := localdb.GetClient()
	if err == nil {
		db.CreateIndex(s.GroupConcernStateKey(), s.GroupConcernStateKey("*"), buntdb.IndexString)
		db.CreateIndex(s.FreshKey(), s.FreshKey("*"), buntdb.IndexString)
		db.CreateIndex(s.UserInfoKey(), s.UserInfoKey("*"), buntdb.IndexString)
		db.CreateIndex(s.ConcernStateKey(), s.ConcernStateKey("*"), buntdb.IndexBinary)
	}
	return s.StateManager.Start()
}

func NewStateManager() *StateManager {
	sm := new(StateManager)
	sm.extraKey = NewExtraKey()
	sm.StateManager = concern_manager.NewStateManager(NewKeySet())
	return sm
}
