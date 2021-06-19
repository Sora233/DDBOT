package concern_manager

import "github.com/Sora233/DDBOT/concern"

// HookResult Pass是false的时候，要把具体失败的地方填入Reason
type HookResult struct {
	Pass   bool
	Reason string
}

func (h *HookResult) PassOrReason(pass bool, reason string) {
	if pass {
		h.Pass = pass
	} else {
		h.Reason = reason
	}
}

type Hook interface {
	AtBeforeHook(notify concern.Notify) *HookResult
	ShouldSendHook(notify concern.Notify) *HookResult
}

type defaultHook struct {
}

func (d defaultHook) AtBeforeHook(notify concern.Notify) *HookResult {
	return &HookResult{
		Pass:   false,
		Reason: "default hook",
	}
}

func (d defaultHook) ShouldSendHook(notify concern.Notify) *HookResult {
	return &HookResult{
		Pass:   false,
		Reason: "default hook",
	}
}
