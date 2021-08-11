package concern_manager

import (
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Sora233/DDBOT/concern"
)

type Callback interface {
	NotifyBeforeCallback(notify concern.Notify)
	NotifyAfterCallback(notify concern.Notify, message *message.GroupMessage)
}

type defaultCallback struct {
}

func (d defaultCallback) NotifyBeforeCallback(notify concern.Notify) {
}

func (d defaultCallback) NotifyAfterCallback(notify concern.Notify, message *message.GroupMessage) {
}
