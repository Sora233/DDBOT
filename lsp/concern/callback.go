package concern

import "github.com/Mrs4s/MiraiGo/message"

// ICallback 定义了一些针对 Notify 推送前后的 callback
type ICallback interface {
	// NotifyBeforeCallback 会在 Notify 推送前获取推送文案内容之前最后一刻进行调用
	// 所以在这个callback里还可以对 Notify 推送内容进行修改
	// b站推送使用了这个callback进行缩略推送
	NotifyBeforeCallback(notify Notify)
	// NotifyAfterCallback 会在 Notify 推送后第一时间进行调用
	NotifyAfterCallback(notify Notify, message *message.GroupMessage)
}

// DefaultCallback ICallback 的默认实现，默认为空
type DefaultCallback struct {
}

// NotifyBeforeCallback stub
func (d DefaultCallback) NotifyBeforeCallback(notify Notify) {
}

// NotifyAfterCallback stub
func (d DefaultCallback) NotifyAfterCallback(notify Notify, message *message.GroupMessage) {
}
