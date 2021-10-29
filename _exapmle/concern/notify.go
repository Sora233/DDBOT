package example_concern

import (
	"github.com/Logiase/MiraiGo-Template/bot"
	"github.com/Mrs4s/MiraiGo/message"
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

func (n *notify) ToMessage() []message.IMessageElement {
	m := mmsg.NewMSG()
	m.Textf("EXAMPLE推送：%v", n.id)
	sending := m.ToMessage(bot.Instance.QQClient, mmsg.NewGroupTarget(n.groupCode))
	return sending.Elements
}

func (n *notify) Logger() *logrus.Entry {
	return logger.WithField("Id", n.id).WithFields(localutils.GroupLogFields(n.groupCode))
}
