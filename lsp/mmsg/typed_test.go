package mmsg

import (
	"testing"

	"github.com/LagrangeDev/LagrangeGo/message"
	"github.com/stretchr/testify/assert"

	"github.com/Sora233/DDBOT/v2/internal/test"
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
	assert.Nil(t, pt.PackToElement(NewGroupTarget(test.ID1)))
	assert.NotNil(t, pt.PackToElement(NewPrivateTarget(test.ID1)))

	gt := NewGroupElement(message.NewText("testge"))
	assert.Nil(t, gt.PackToElement(NewPrivateTarget(test.ID2)))
	assert.NotNil(t, gt.PackToElement(NewGroupTarget(test.ID2)))
}
