package example_concern

import (
	"github.com/Sora233/DDBOT/lsp/concern_type"
	"github.com/Sora233/DDBOT/lsp/mmsg"
	localutils "github.com/Sora233/DDBOT/utils"
	"github.com/sirupsen/logrus"
)

type notify struct {
	groupCode int64
	id        string
}

func (n *notify) Site() string {
	return Site
}

func (n *notify) Type() concern_type.Type {
	return Example
}

func (n *notify) GetGroupCode() int64 {
	return n.groupCode
}

func (n *notify) GetUid() interface{} {
	return n.id
}

func (n *notify) ToMessage() *mmsg.MSG {
	return mmsg.NewTextf("EXAMPLE推送：%v", n.id)
}

func (n *notify) Logger() *logrus.Entry {
	return logger.WithField("Id", n.id).WithFields(localutils.GroupLogFields(n.groupCode))
}
