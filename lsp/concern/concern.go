package concern

import (
	"github.com/Sora233/DDBOT/lsp/concern_type"
	"github.com/Sora233/DDBOT/lsp/mmsg"
	"github.com/sirupsen/logrus"
)

type Event interface {
	Site() string
	Type() concern_type.Type
	GetUid() interface{}
	Logger() *logrus.Entry
}

type Notify interface {
	Event
	GetGroupCode() int64
	ToMessage() *mmsg.MSG
}

// NotifyLiveExt 是一个针对直播推送过滤的扩展接口， Notify 可以选择性实现这个接口，如果实现了，则会自动使用默认的推送过滤逻辑
// 默认情况下，如果 IsLive 为 true，则根据以下规则推送：
// Living 为 true 且 LiveStatusChanged 为true（说明是开播了）进行推送
// Living 为 false 且 LiveStatusChanged 为true（说明是下播了）根据 OfflineNotify 配置进行推送推送
// Living 为 true 且 LiveStatusChanged 为false 且 TitleChanged 为true（说明是上播状态更改标题）根据 TitleChangeNotify 配置进行推送推送
// 如果 Notify 没有实现这个接口，则会推送所有内容
type NotifyLiveExt interface {
	IsLive() bool
	Living() bool
	TitleChanged() bool
	LiveStatusChanged() bool
}

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
	List(groupCode int64, ctype concern_type.Type) ([]IdentityInfo, []concern_type.Type, error)
	Get(id interface{}) (IdentityInfo, error)

	GetStateManager() IStateManager
	FreshIndex(groupCode ...int64)
}

// IdentityInfo 表示订阅对象的信息，包括名字，ID
type IdentityInfo interface {
	GetUid() interface{}
	GetName() string
}
