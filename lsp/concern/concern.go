package concern

import (
	"github.com/Sora233/DDBOT/lsp/concern_type"
	"github.com/Sora233/DDBOT/lsp/mmsg"
	"github.com/sirupsen/logrus"
)

// Event 是对事件的一个抽象，它可以表示发表动态，发表微博，发布文章，发布视频，是订阅对象做出的行为
// 通常是由一个爬虫负责产生，例如：当b站主播发布了新动态的时候，爬虫抓到这条动态，就产生了一个 Event
// Event 不应该关联推送的接收方的信息，例如：不应含有qq群号码
type Event interface {
	Site() string
	Type() concern_type.Type
	GetUid() interface{}
	Logger() *logrus.Entry
}

// Notify 是对推送的一个抽象，它在 Event 的基础上还包含了推送的接受方信息，例如：qq群号码
// Event 产生后，通过 Event + 需要推送的QQ群信息，由 Dispatch 和 NotifyGenerator 产生一组 Notify
// 因为可能多个群订阅同一个 Event，所以一个 Event 可以产生多个 Notify
// DDBOT目前只支持向QQ群推送
type Notify interface {
	Event
	GetGroupCode() int64
	ToMessage() *mmsg.MSG
}

// Concern 是DDBOT的一个完整订阅模块，包含一个订阅源的全部信息
// 当一个 Concern 编写完成后，需要使用 concern.RegisterConcern 向DDBOT注册才能生效
type Concern interface {
	// Site 必须全局唯一，不允许注册两个相同的site
	Site() string
	// Types 返回该 Concern 支持的 concern_type.Type，此处返回的每一项必须是单个type，并且第一个type为默认type
	Types() []concern_type.Type
	Start() error
	Stop()
	ParseId(string) (interface{}, error)

	Add(ctx mmsg.IMsgCtx, groupCode int64, id interface{}, ctype concern_type.Type) (IdentityInfo, error)
	Remove(ctx mmsg.IMsgCtx, groupCode int64, id interface{}, ctype concern_type.Type) (IdentityInfo, error)
	Get(id interface{}) (IdentityInfo, error)

	GetStateManager() IStateManager
	FreshIndex(groupCode ...int64)
}

// IdentityInfo 表示订阅对象的信息，包括名字，ID
type IdentityInfo interface {
	GetUid() interface{}
	GetName() string
}
