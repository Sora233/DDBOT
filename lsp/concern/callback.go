package concern

import "github.com/Mrs4s/MiraiGo/message"

type ICallback interface {
	NotifyBeforeCallback(notify Notify)
	NotifyAfterCallback(notify Notify, message *message.GroupMessage)
}

type DefaultCallback struct {
}

func (d DefaultCallback) NotifyBeforeCallback(notify Notify) {
}

func (d DefaultCallback) NotifyAfterCallback(notify Notify, message *message.GroupMessage) {
}
