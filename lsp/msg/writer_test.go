package msg

import (
	"github.com/Mrs4s/MiraiGo/message"
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

	m = NewText("1")
	m = NewTextf("asd %v", "xxx")

}
