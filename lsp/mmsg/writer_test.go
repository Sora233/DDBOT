package mmsg

import (
	"testing"

	"github.com/LagrangeDev/LagrangeGo/message"
	"github.com/stretchr/testify/assert"

	"github.com/Sora233/DDBOT/v2/internal/test"
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
	m.AtAll()
	m.Append(message.NewText("test"))
	m.Append(nil)
	m.Append()
	m.Image(nil, "")
	m.Image(nil, "[image]")
	m.ImageWithNorm(nil, "[image]")
	m.ImageWithResize(nil, "[image]", 200, 200)
	m.ImageByUrl(test.FakeImage(150), "[url]")
	m.ImageByUrlWithNorm(test.FakeImage(150), "[img]")
	m.ImageByUrlWithResize(test.FakeImage(150), "[img]", 200, 200)
	m.ImageByLocal("", "[img]")
	m.ImageByLocalWithNorm("", "[img]")
	m.ImageByLocalWithResize("", "[img]", 200, 200)
	m.Append(NewTypedElement().OnPrivate(message.NewText("test")))
	m.Append(NewTypedElement())
	m.Append(NewTypedElement().OnGroup(NewImage(nil)))
	m = m.Clone()
	assert.Len(t, m.Elements(), 17)

	m.ToCombineMessage(NewGroupTarget(test.G1))
	m.ToCombineMessage(NewPrivateTarget(1))

	m = NewText("1")
	m = NewTextf("asd %v", "xxx")

}
