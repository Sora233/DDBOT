package utils

import (
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Sora233/DDBOT/internal/test"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSerializationGroupMsg(t *testing.T) {
	msg := &message.GroupMessage{
		Id:        1,
		GroupCode: test.G1,
		Sender: &message.Sender{
			Uin: test.ID1,
		},
		Time: 30,
		Elements: []message.IMessageElement{
			message.NewText("qwe"),
			message.NewText("asd"),
			&message.GroupImageElement{ImageId: "1231we"},
			&message.FriendImageElement{ImageId: "qwe"},
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
		MessageTextf("asd"),
		&message.GroupImageElement{},
		&message.ServiceElement{},
		&message.AtElement{},
	}
	c := MessageFilter(e, func(element message.IMessageElement) bool {
		return element.Type() == message.Text
	})
	assert.Len(t, c, 1)

	c = MessageFilter(e, func(element message.IMessageElement) bool {
		return element.Type() == message.Text || element.Type() == message.Service
	})
	assert.Len(t, c, 2)

	c = MessageFilter(e, func(element message.IMessageElement) bool {
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
	img, err := ImageGet(test.FakeImage(10))
	assert.Nil(t, err)
	e, err = UploadPrivateImage(1, img, true)
	assert.NotNil(t, err)
	assert.Nil(t, e)
}
