package mmsg

import (
	"testing"

	"github.com/LagrangeDev/LagrangeGo/message"
	"github.com/stretchr/testify/assert"
)

func TestImage(t *testing.T) {
	var im *ImageBytesElement
	e := im.PackToElement(NewGroupTarget(0))
	assert.Equal(t, e.(*message.TextElement).Content, "[nil image]\n")

	im = NewImage(nil)
	im.Alternative("test")
	assert.EqualValues(t, ImageBytes, im.Type())
	e = im.PackToElement(NewGroupTarget(0))
	assert.Equal(t, e.(*message.TextElement).Content, "test\n")

	assert.NotPanics(t, func() {
		im.Norm().Resize(100, 100)
	})
}
