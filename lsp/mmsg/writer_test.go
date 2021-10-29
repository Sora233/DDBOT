package mmsg

import (
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Sora233/DDBOT/proxy_pool"
	"testing"
)

func TestMSG(t *testing.T) {
	m := NewMSG()
	m.Text("1")
	m.Text("2")
	m.Textf("3 %v", "xxx")
	m.flushText()
	m.Text("1")
	m.Text("2")
	m.Text("3")
	m.Append(message.NewAt(0))
	m.Append(message.NewText("test"))
	m.Image(nil, "")
	m.ImageByUrl("https://via.placeholder.com/1500", "", proxy_pool.PreferAny)

	m = NewText("1")
	m = NewTextf("asd %v", "xxx")

}
