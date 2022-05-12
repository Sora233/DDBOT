package mmsg

import (
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Sora233/DDBOT/internal/test"
	"github.com/Sora233/DDBOT/lsp/mmsg/mt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTyped(t *testing.T) {
	tp := NewTypedElement()
	assert.EqualValues(t, Typed, tp.Type())

	assert.Panics(t, func() {
		tp.OnPrivate(tp)
	})
	assert.Panics(t, func() {
		tp.OnGroup(tp)
	})

	m := NewMSG()
	m.Append(NewTypedElement().OnGroup(nil))
	m.Append(NewTypedElement().OnPrivate(nil))

	pt := NewPrivateElement(message.NewText("testpe"))
	assert.Nil(t, pt.PackToElement(mt.NewGroupTarget(test.ID1)))
	assert.NotNil(t, pt.PackToElement(mt.NewPrivateTarget(test.ID1)))

	gt := NewGroupElement(message.NewText("testge"))
	assert.Nil(t, gt.PackToElement(mt.NewPrivateTarget(test.ID2)))
	assert.NotNil(t, gt.PackToElement(mt.NewGroupTarget(test.ID2)))
}
