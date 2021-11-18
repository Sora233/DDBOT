package mmsg

import (
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Sora233/DDBOT/internal/test"
	"github.com/Sora233/DDBOT/proxy_pool"
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
	m.Image(nil, "")
	m.Image(nil, "[image]")
	m.ImageByUrl(test.FakeImage(150), "[url]", proxy_pool.PreferAny)
	m.NormImageByUrl(test.FakeImage(150), "[img]", proxy_pool.PreferAny)
	m.Append(NewTypedElement().OnPrivate(message.NewText("test")))
	m.Append(NewTypedElement())
	m.Append(NewTypedElement().OnGroup(NewImage(nil)))
	assert.Len(t, m.Elements(), 11)

	m.ToMessage(NewGroupTarget(test.G1))
	m.ToMessage(NewPrivateTarget(1))

	m = NewText("1")
	m = NewTextf("asd %v", "xxx")

}
