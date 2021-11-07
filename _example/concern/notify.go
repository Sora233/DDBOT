package example_concern

import (
	"github.com/Sora233/DDBOT/lsp/concern_type"
	"github.com/Sora233/DDBOT/lsp/mmsg"
	localutils "github.com/Sora233/DDBOT/utils"
	"github.com/sirupsen/logrus"
)

// Event 实现 concern.Event 接口
type Event struct {
	id string
}

func (e *Event) Site() string {
	return Site
}

func (e *Event) Type() concern_type.Type {
	return Example
}

func (e *Event) GetUid() interface{} {
	return e.id
}

func (e *Event) Logger() *logrus.Entry {
	return logger.WithField("Id", e.id)
}

// Notify 实现 concern.Notify 接口
type Notify struct {
	groupCode int64
	*Event
}

func (n *Notify) GetGroupCode() int64 {
	return n.groupCode
}

func (n *Notify) ToMessage() *mmsg.MSG {
	return mmsg.NewTextf("EXAMPLE推送：%v", n.id)
}

func (n *Notify) Logger() *logrus.Entry {
	return n.Event.Logger().WithFields(localutils.GroupLogFields(n.groupCode))
}
