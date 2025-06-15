package utils

import (
	"testing"

	"github.com/LagrangeDev/LagrangeGo/message"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"

	"github.com/Sora233/DDBOT/v2/internal/test"
)

func TestSerializationGroupMsg(t *testing.T) {
	msg := &message.GroupMessage{
		ID:       1,
		GroupUin: test.G1,
		Sender: &message.Sender{
			Uin: test.ID1,
		},
		Time: 30,
		Elements: []message.IMessageElement{
			message.NewText("qwe"),
			message.NewText("asd"),
			&message.ImageElement{ImageID: "1231we"},
			&message.ImageElement{ImageID: "qwe"},
		},
	}

	msgString, err := SerializationGroupMsg(msg)
	assert.Nil(t, err)
	msg2, err := DeserializationGroupMsg(msgString)
	assert.Nil(t, err)
	assert.EqualValues(t, msg, msg2)
}

func TestMessageFilter(t *testing.T) {
	var e = []message.IMessageElement{
		message.NewText("asd"),
		&message.ImageElement{},
		&message.XMLElement{},
		&message.AtElement{},
	}
	c := lo.Filter(e, func(element message.IMessageElement, _ int) bool {
		return element.Type() == message.Text
	})
	assert.Len(t, c, 1)

	c = lo.Filter(e, func(element message.IMessageElement, _ int) bool {
		return element.Type() == message.Text || element.Type() == message.Service
	})
	assert.Len(t, c, 2)

	c = lo.Filter(e, func(element message.IMessageElement, _ int) bool {
		return element.Type() == message.At
	})
	assert.Len(t, c, 1)
}

func TestUploadGroupImage(t *testing.T) {
	test.InitMirai()
	defer test.CloseMirai()
	e, err := UploadGroupImage(test.G1, []byte("asdsad"), true)
	assert.NotNil(t, err)
	assert.Nil(t, e)
	e, err = UploadGroupImageByUrl(test.G1, test.FakeImage(10), true)
	assert.NotNil(t, err)
}

func TestUploadPrivateImage(t *testing.T) {
	test.InitMirai()
	defer test.CloseMirai()
	e, err := UploadPrivateImage(1, []byte("asdsad"), true)
	img, err := ImageGet(imageUrl)
	assert.Nil(t, err)
	e, err = UploadPrivateImage(1, img, true)
	assert.NotNil(t, err)
	assert.Nil(t, e)
}
