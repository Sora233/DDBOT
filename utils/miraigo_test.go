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