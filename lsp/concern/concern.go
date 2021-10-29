package concern

import (
	"github.com/Sora233/DDBOT/lsp/concern_type"
	"github.com/Sora233/DDBOT/lsp/mmsg"
	"github.com/sirupsen/logrus"
)

type Notify interface {
	Site() string
	Type() concern_type.Type
	GetGroupCode() int64
	GetUid() interface{}
	ToMessage() *mmsg.MSG
	Logger() *logrus.Entry
}

type Concern interface {
	Site() string
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
