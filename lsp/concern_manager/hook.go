package concern_manager

import "github.com/Sora233/DDBOT/concern"

type Hook interface {
	AtBeforeHook(notify concern.Notify) bool
	ShouldSendHook(notify concern.Notify) bool
}

type defaultHook struct {
}

func (d defaultHook) AtBeforeHook(notify concern.Notify) bool {
	return false
}

func (d defaultHook) ShouldSendHook(notify concern.Notify) bool {
	return false
}
