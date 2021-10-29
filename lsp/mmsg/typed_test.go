package mmsg

import (
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Sora233/DDBOT/internal/test"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTyped(t *testing.T) {
	m := NewMSG()
	m.Append(NewTypedElement().OnGroup(nil))
	m.Append(NewTypedElement().OnPrivate(nil))

	pt := NewPrivateElement(message.NewText("testpe"))
	assert.Nil(t, pt.PackToElement(nil, NewGroupTarget(test.ID1)))
	assert.NotNil(t, pt.PackToElement(nil, NewPrivateTarget(test.ID1)))

	gt := NewGroupElement(message.NewText("testge"))
	assert.Nil(t, gt.PackToElement(nil, NewPrivateTarget(test.ID2)))
	assert.NotNil(t, gt.PackToElement(nil, NewGroupTarget(test.ID2)))
}
