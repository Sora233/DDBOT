package template

import (
	"github.com/Sora233/DDBOT/lsp/cfg"
	"github.com/Sora233/DDBOT/lsp/mmsg"
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
