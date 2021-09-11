package concern

import "github.com/Mrs4s/MiraiGo/message"

type ICallback interface {
	NotifyBeforeCallback(notify Notify)
	NotifyAfterCallback(notify Notify, message *message.GroupMessage)
}

type DefaultCallback struct {
}

func (d DefaultCallback) NotifyBeforeCallback(notify Notify) {
	if notify == nil {
		return
	}
	notify.Logger().Trace("default NotifyBeforeCallback")
}

func (d DefaultCallback) NotifyAfterCallback(notify Notify, message *message.GroupMessage) {
	if notify == nil {
		return
	}
	notify.Logger().Trace("default NotifyAfterCallback")
}
