package template

import (
	"fmt"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Sora233/DDBOT/lsp/cfg"
	"github.com/Sora233/DDBOT/lsp/mmsg"
	"math/rand"
	"os"
	"path/filepath"
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
		if f, _ := os.Open(uri); f != nil {
			if dirs, _ := f.ReadDir(-1); dirs != nil {
				var result []os.DirEntry
				for _, file := range dirs {
					if file.IsDir() || (!strings.HasSuffix(file.Name(), ".jpg") &&
						!strings.HasSuffix(file.Name(), ".png")) {
						continue
					}
					result = append(result, file)
				}
				if len(result) > 0 {
					e = mmsg.NewImageByLocal(filepath.Join(uri, result[rand.Intn(len(result))].Name()))
				} else {
					logger.WithField("uri", uri).Errorf("template: pic uri can not find any images")
				}
			}
		}
		if e == nil {
			e = mmsg.NewImageByLocal(uri)
		}
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
