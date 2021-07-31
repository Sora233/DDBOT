package youtube

import (
	"encoding/json"
	localdb "github.com/Sora233/DDBOT/lsp/buntdb"
	"github.com/Sora233/DDBOT/lsp/concern_manager"
	"github.com/tidwall/buntdb"
	"time"
)

type StateManager struct {
	*concern_manager.StateManager
	*extraKey
}

func (s *StateManager) AddInfo(info *Info) error {
	return s.RWCoverTx(func(tx *buntdb.Tx) error {
		infoKey := s.InfoKey(info.ChannelId)
		_, _, err := tx.Set(infoKey, info.ToString(), localdb.ExpireOption(time.Hour*24*7))
		return err
	})
}

func (s *StateManager) GetInfo(channelId string) (info *Info, err error) {
	err = s.RCoverTx(func(tx *buntdb.Tx) error {
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

func (s *StateManager) GetVideo(channelId string, videoId string) (*VideoInfo, error) {
	var v = new(VideoInfo)
	err := s.RCoverTx(func(tx *buntdb.Tx) error {
		key := s.VideoKey(channelId, videoId)
		val, err := tx.Get(key)
		if err != nil {
			return err
		}
		err = json.Unmarshal([]byte(val), v)
		return err
	})
	if err != nil {
		return nil, err
	}
	return v, nil
}

func (s *StateManager) AddVideo(v *VideoInfo) error {
	return s.RWCoverTx(func(tx *buntdb.Tx) error {
		key := s.VideoKey(v.ChannelId, v.VideoId)
		b, err := json.Marshal(v)
		if err != nil {
			return err
		}
		_, _, err = tx.Set(key, string(b), localdb.ExpireOption(time.Hour*24))
		return err
	})
}

func (s *StateManager) Start() error {
	db, err := localdb.GetClient()
	if err == nil {
		db.CreateIndex(s.GroupConcernStateKey(), s.GroupConcernStateKey("*"), buntdb.IndexString)
		db.CreateIndex(s.FreshKey(), s.FreshKey("*"), buntdb.IndexString)
		db.CreateIndex(s.UserInfoKey(), s.UserInfoKey("*"), buntdb.IndexString)
	}
	return s.StateManager.Start()
}

func NewStateManager() *StateManager {
	sm := new(StateManager)
	sm.extraKey = NewExtraKey()
	sm.StateManager = concern_manager.NewStateManager(NewKeySet(), true)
	return sm
}
