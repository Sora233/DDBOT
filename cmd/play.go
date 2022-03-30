package main

import (
	"fmt"
	"github.com/Sora233/DDBOT/lsp/mmsg"
	"github.com/Sora233/DDBOT/lsp/template"
	"github.com/Sora233/DDBOT/utils/msgstringer"
	"github.com/sirupsen/logrus"
	"time"
)

func play() {
	logrus.SetLevel(logrus.DebugLevel)
	template.InitTemplateLoader()
	for {
		t := template.LoadTemplate("command.private.help.tmpl")
		if t == nil {
			panic("t is nil")
		}
		m := mmsg.NewMSG()
		if err := t.ExecuteTemplate(m, "command.private.help.tmpl", nil); err != nil {
			panic(err)
		}
		fmt.Println(msgstringer.MsgToString(m.Elements()))
		time.Sleep(time.Second * 3)
	}
}
