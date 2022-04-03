package mmsg

import (
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Sora233/DDBOT/internal/test"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMSG(t *testing.T) {
	test.InitMirai()
	defer test.CloseMirai()

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
	m.Append(nil)
	m.Append()
	m.Image(nil, "")
	m.Image(nil, "[image]")
	m.ImageWithNorm(nil, "[image]")
	m.ImageByUrl(test.FakeImage(150), "[url]")
	m.ImageByUrlWithNorm(test.FakeImage(150), "[img]")
	m.Append(NewTypedElement().OnPrivate(message.NewText("test")))
	m.Append(NewTypedElement())
	m.Append(NewTypedElement().OnGroup(NewImage(nil)))
	assert.Len(t, m.Elements(), 12)

	m.ToCombineMessage(NewGroupTarget(test.G1))
	m.ToCombineMessage(NewPrivateTarget(1))

	m = NewText("1")
	m = NewTextf("asd %v", "xxx")

}
