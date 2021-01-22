package youtube

import (
	"github.com/Sora233/Sora233-MiraiGo/lsp/concern_manager"
	"github.com/Sora233/Sora233-MiraiGo/utils"
	"time"
)
// TODO
type StateManager struct {
	*concern_manager.StateManager
	emit     *utils.EmitQueue
	emitChan chan interface{}
}

func (s *StateManager) AddChannel(channelId string) {
	
}

func (s *StateManager) AddInfo(info *Info) {

}

func (s *StateManager) GetInfo(channelId string) (*Info, error) {
	return nil, nil
}

func NewStateManager(c chan<- interface{}) *StateManager {
	emitChan := make(chan interface{})
	sm := &StateManager{
		emitChan:     emitChan,
		StateManager: concern_manager.NewStateManager(NewKeySet(), emitChan),
		emit:         utils.NewEmitQueue(c, time.Second*5),
	}
	return sm
}
