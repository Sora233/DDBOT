package msgstringer

import (
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Sora233/DDBOT/lsp/mmsg"
	"github.com/davecgh/go-spew/spew"
	"strconv"
	"strings"
)

func MsgToString(elements []message.IMessageElement) string {
	var res strings.Builder
	for _, elem := range elements {
		if elem == nil {
			continue
		}
		switch e := elem.(type) {
		case *message.TextElement:
			res.WriteString(e.Content)
		case *message.FaceElement:
			res.WriteString("[")
			res.WriteString(e.Name)
			res.WriteString("]")
		case *message.GroupImageElement:
			if e.Flash {
				res.WriteString("[Flash Image]")
			} else {
				res.WriteString("[Image]")
			}
		case *message.FriendImageElement:
			if e.Flash {
				res.WriteString("[Flash Image]")
			} else {
				res.WriteString("[Image]")
			}
		case *message.AtElement:
			res.WriteString(e.Display)
		case *message.RedBagElement:
			res.WriteString("[RedBag:")
			res.WriteString(e.Title)
			res.WriteString("]")
		case *message.ReplyElement:
			res.WriteString("[Reply:")
			res.WriteString(strconv.FormatInt(int64(e.ReplySeq), 10))
			res.WriteString("]")
		case *message.GroupFileElement:
			res.WriteString("[File]")
			res.WriteString(e.Name)
		case *message.ShortVideoElement:
			res.WriteString("[Video]")
		case *message.ForwardElement:
			res.WriteString("[Forward]")
		case *message.MusicShareElement:
			res.WriteString("[Music]")
		case *message.LightAppElement:
			res.WriteString("[LightApp]")
			res.WriteString(e.Content)
		case *message.ServiceElement:
			res.WriteString("[Service]")
			res.WriteString(e.Content)
		case *message.VoiceElement, *message.PrivateVoiceElement, *message.GroupVoiceElement:
			res.WriteString("[Voice]")
		case *mmsg.ImageBytesElement:
			res.WriteString("[Image]")
		case *mmsg.TypedElement:
			res.WriteString("[Typed]")
		case *message.MarketFaceElement:
			res.WriteString(e.Name)
		case *message.DiceElement:
			res.WriteString(e.Name)
			res.WriteString(strconv.FormatInt(int64(e.Value), 10))
		case *message.AnimatedSticker:
			res.WriteString("[")
			res.WriteString(e.Name)
			res.WriteString("]")
		default:
			logger.WithField("content", spew.Sdump(elem)).Debug("found new element")
		}
	}
	return res.String()
}
