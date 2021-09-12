package concern

type Hook interface {
	FilterHook(notify Notify) *HookResult
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

type DefaultHook struct {
}

func (d DefaultHook) FilterHook(notify Notify) *HookResult {
	return HookResultPass
}

func (d DefaultHook) AtBeforeHook(notify Notify) *HookResult {
	return defaultHookResult
}

func (d DefaultHook) ShouldSendHook(notify Notify) *HookResult {
	return HookResultPass
}

var defaultHookResult = &HookResult{
	Pass:   false,
	Reason: "default hook",
}

var HookResultPass = &HookResult{
	Pass: true,
}
