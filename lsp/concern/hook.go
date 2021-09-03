package concern

type Hook interface {
	NewsFilterHook(notify Notify) *HookResult
	AtBeforeHook(notify Notify) *HookResult
	ShouldSendHook(notify Notify) *HookResult
}

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

type defaultHook struct {
}

func (d defaultHook) NewsFilterHook(notify Notify) *HookResult {
	return defaultHookResult
}

func (d defaultHook) AtBeforeHook(notify Notify) *HookResult {
	return defaultHookResult
}

func (d defaultHook) ShouldSendHook(notify Notify) *HookResult {
	return defaultHookResult
}

var defaultHookResult = &HookResult{
	Pass:   false,
	Reason: "default hook",
}

var HookResultPass = &HookResult{
	Pass: true,
}
