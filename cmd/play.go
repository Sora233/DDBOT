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
		t := template.LoadTemplate("command.group.checkin.tmpl")
		if t == nil {
			panic("t is nil")
		}
		m := mmsg.NewMSG()
		if err := t.ExecuteTemplate(m, "command.group.checkin.tmpl", map[string]interface{}{
			"score":   100,
			"success": true,
		}); err != nil {
			panic(err)
		}
		fmt.Println(msgstringer.MsgToString(m.Elements()))
		time.Sleep(time.Second * 3)
	}
}
