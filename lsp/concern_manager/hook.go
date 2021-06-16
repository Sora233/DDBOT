package concern_manager

import "github.com/Sora233/DDBOT/concern"

type Hook interface {
	AtAllBeforeHook(notify concern.Notify) bool
	ShouldSendHook(notify concern.Notify) bool
}

type defaultHook struct {
}

func (d defaultHook) AtAllBeforeHook(notify concern.Notify) bool {
	return false
}

func (d defaultHook) ShouldSendHook(notify concern.Notify) bool {
	return false
}
