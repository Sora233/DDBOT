package msgstringer

import (
	"github.com/Mrs4s/MiraiGo/message"
	"testing"
)

func TestMsgToString(t *testing.T) {
	var m = []message.IMessageElement{
		message.NewFace(1),
		message.NewText("q"),
		&message.GroupImageElement{},
		&message.GroupImageElement{Flash: true},
		&message.FriendImageElement{},
		&message.FriendImageElement{Flash: true},
		message.AtAll(),
		&message.RedBagElement{},
		&message.GroupFileElement{},
		&message.ShortVideoElement{},
		&message.ForwardElement{},
		&message.MusicShareElement{},
		&message.LightAppElement{},
		&message.ServiceElement{},
		&message.VoiceElement{},
		&message.ReplyElement{ReplySeq: 199},
		nil,
	}
	MsgToString(m)
}
