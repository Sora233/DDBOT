package msg

import (
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestImage(t *testing.T) {
	var im *ImageBytesElement
	e := im.PackToElement(nil, NewGroupTarget(0))
	assert.Equal(t, e.(*message.TextElement).Content, "[nil image]")

	im = NewImage(nil)
	im.Alternative("test")
	e = im.PackToElement(nil, NewGroupTarget(0))
	assert.Equal(t, e.(*message.TextElement).Content, "test")
}
