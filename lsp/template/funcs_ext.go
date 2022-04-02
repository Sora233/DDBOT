package template

import (
	"fmt"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Sora233/DDBOT/lsp/cfg"
	"github.com/Sora233/DDBOT/lsp/mmsg"
	"math/rand"
	"strings"
)

var funcsExt = make(FuncMap)

// RegisterExtFunc 在init阶段插入额外的template函数
func RegisterExtFunc(name string, fn interface{}) {
	checkValueFuncs(name, fn)
	funcsExt[name] = fn
}

func cut() *mmsg.CutElement {
	return new(mmsg.CutElement)
}

func prefix() string {
	return cfg.GetCommandPrefix()
}

func reply(msg interface{}) *message.ReplyElement {
	if msg == nil {
		return nil
	}
	switch e := msg.(type) {
	case *message.GroupMessage:
		return message.NewReply(e)
	case *message.PrivateMessage:
		return message.NewPrivateReply(e)
	default:
		panic(fmt.Sprintf("unknown reply message %v", msg))
	}
}

func pic(uri string, alternative ...string) (e *mmsg.ImageBytesElement) {
	if strings.HasPrefix(uri, "http://") || strings.HasPrefix(uri, "https://") {
		e = mmsg.NewImageByUrl(uri)
	} else {
		e = mmsg.NewImageByLocal(uri)
	}
	if len(alternative) > 0 && len(alternative[0]) > 0 {
		e.Alternative(alternative[0])
	}
	return e
}

func roll(from, to int64) int64 {
	return rand.Int63n(to-from+1) + from
}

func choose(items ...string) string {
	return items[rand.Intn(len(items))]
}
