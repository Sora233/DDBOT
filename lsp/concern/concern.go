package concern

import (
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Sora233/DDBOT/lsp/concern_type"
	"github.com/Sora233/DDBOT/lsp/msg"
	"github.com/sirupsen/logrus"
)

type Notify interface {
	Site() string
	Type() concern_type.Type
	GetGroupCode() int64
	GetUid() interface{}
	ToMessage() []message.IMessageElement
	Logger() *logrus.Entry
}

type Concern interface {
	Site() string
	Start() error
	Stop()
	ParseId(string) (interface{}, error)

	Add(ctx msg.IMsgCtx, groupCode int64, id interface{}, ctype concern_type.Type) (IdentityInfo, error)
	Remove(ctx msg.IMsgCtx, groupCode int64, id interface{}, ctype concern_type.Type) (IdentityInfo, error)
	List(groupCode int64, ctype concern_type.Type) ([]IdentityInfo, []concern_type.Type, error)
	Get(id interface{}) (IdentityInfo, error)

	GetStateManager() IStateManager
	FreshIndex(groupCode ...int64)
}

// IdentityInfo 表示订阅对象的信息，包括名字，ID，type
type IdentityInfo interface {
	GetUid() interface{}
	GetName() string
}
