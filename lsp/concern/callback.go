package concern

import (
	"github.com/Mrs4s/MiraiGo/message"
)

type Callback interface {
	NotifyBeforeCallback(notify Notify)
	NotifyAfterCallback(notify Notify, message *message.GroupMessage)
}

type defaultCallback struct {
}

func (d defaultCallback) NotifyBeforeCallback(notify Notify) {
}

func (d defaultCallback) NotifyAfterCallback(notify Notify, message *message.GroupMessage) {
}
