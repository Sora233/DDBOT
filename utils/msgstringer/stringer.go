package msgstringer

import (
	"strconv"
	"strings"

	"github.com/LagrangeDev/LagrangeGo/message"
	"github.com/davecgh/go-spew/spew"

	"github.com/Sora233/DDBOT/v2/lsp/mmsg"
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
			res.WriteString(strconv.Itoa(int(e.FaceID)))
			res.WriteString("]")
		case *message.ImageElement:
			if e.Flash {
				res.WriteString("[Flash Image]")
			} else {
				res.WriteString("[Image]")
			}
		case *message.AtElement:
			res.WriteString(e.Display)
		case *message.ReplyElement:
			res.WriteString("[Reply:")
			res.WriteString(strconv.FormatInt(int64(e.ReplySeq), 10))
			res.WriteString("]")
		case *message.FileElement:
			res.WriteString("[File]")
			res.WriteString(e.FileName)
		case *message.ShortVideoElement:
			res.WriteString("[Video]")
		case *message.ForwardMessage:
			res.WriteString("[Forward]")
		case *message.LightAppElement:
			res.WriteString("[LightApp]")
			res.WriteString(e.Content)
		case *message.XMLElement:
			res.WriteString("[XML]")
			res.WriteString(e.Content)
		case *message.VoiceElement:
			res.WriteString("[Voice]")
		case *mmsg.ImageBytesElement:
			res.WriteString("[Image]")
		case *mmsg.TypedElement:
			res.WriteString("[Typed]")
		case *mmsg.CutElement:
			res.WriteString("[CUT]")
		case *message.MarketFaceElement:
			res.WriteString(e.Summary)
		default:
			logger.WithField("content", spew.Sdump(elem)).Debug("found new element")
		}
	}
	return res.String()
}
