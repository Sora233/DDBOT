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
