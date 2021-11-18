package concern

// Hook 定义了一些对 Notify 进行推送过程中的拦截器
type Hook interface {
	// FilterHook 是更细粒度的过滤，可以根据这条推送的内容、文案来判断是否应该推送
	// 例如b站的动态类型过滤就是使用了 FilterHook
	FilterHook(notify Notify) *HookResult
	// AtBeforeHook 控制是否应该执行@操作，注意即使通过了也并不代表一定会@，还需要配置@才可以
	// 如果没有配置，则没有@的对象，也就不会进行@
	AtBeforeHook(notify Notify) *HookResult
	// ShouldSendHook 根据 Notify 本身的状态（而非推送的文案、内容）决定是否进行推送
	// 例如上播推送，下播推送
	// 如果要根据推送的文案、内容判断，则应该使用 FilterHook
	ShouldSendHook(notify Notify) *HookResult
}

// HookResult 定义了 Hook 的结果，Pass是false的时候，要把具体失败的地方填入Reason
type HookResult struct {
	Pass   bool
	Reason string
}

// PassOrReason 如果pass为true，则 HookResult 为true，否则设置 HookResult 的 Reason
func (h *HookResult) PassOrReason(pass bool, reason string) {
	if pass {
		h.Pass = pass
	} else {
		h.Reason = reason
	}
}

var defaultHookResult = &HookResult{
	Pass:   false,
	Reason: "default hook",
}

// HookResultPass 预定义Pass状态的 HookResult
var HookResultPass = &HookResult{
	Pass: true,
}
