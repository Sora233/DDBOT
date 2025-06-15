package youtube

import (
	"time"

	localdb "github.com/Sora233/DDBOT/v2/lsp/buntdb"
	"github.com/Sora233/DDBOT/v2/lsp/concern"
)

type StateManager struct {
	*concern.StateManager
	*extraKey
}

func (s *StateManager) AddInfo(info *Info) error {
	return s.SetJson(s.InfoKey(info.ChannelId), info, localdb.SetExpireOpt(time.Hour*24*7))
}

func (s *StateManager) GetInfo(channelId string) (*Info, error) {
	info := new(Info)
	err := s.GetJson(s.InfoKey(channelId), info)
	if err != nil {
		return nil, err
	}
	return info, nil
}

func (s *StateManager) GetVideo(channelId string, videoId string) (*VideoInfo, error) {
	var v *VideoInfo
	err := s.GetJson(s.VideoKey(channelId, videoId), &v)
	if err != nil {
		return nil, err
	}
	return v, nil
}

func (s *StateManager) AddVideo(v *VideoInfo) error {
	return s.SetJson(s.VideoKey(v.ChannelId, v.VideoId), v)
}

func (s *StateManager) GetGroupConcernConfig(groupCode uint32, id interface{}) (concernConfig concern.IConfig) {
	return NewGroupConcernConfig(s.StateManager.GetGroupConcernConfig(groupCode, id))
}

func NewStateManager(notify chan<- concern.Notify) *StateManager {
	sm := new(StateManager)
	sm.extraKey = NewExtraKey()
	sm.StateManager = concern.NewStateManagerWithCustomKey(Site, NewKeySet(), notify)
	return sm
}
