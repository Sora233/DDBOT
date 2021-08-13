package lsp

import (
	"github.com/Mrs4s/MiraiGo/message"
	localdb "github.com/Sora233/DDBOT/lsp/buntdb"
	"github.com/Sora233/DDBOT/lsp/test"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/buntdb"
	"testing"
)

func newStateManager(t *testing.T) *StateManager {
	sm := NewStateManager()
	assert.NotNil(t, sm)
	sm.FreshIndex()
	return sm
}

func TestNewStateManager(t *testing.T) {
	test.InitBuntdb(t)
	defer test.CloseBuntdb(t)

	sm := newStateManager(t)
	assert.NotNil(t, sm)
}

func TestStateManager_GetGroupInvitor(t *testing.T) {
	test.InitBuntdb(t)
	defer test.CloseBuntdb(t)

	sm := newStateManager(t)
	assert.NotNil(t, sm)

	_, err := sm.PopGroupInvitor(test.G1)
	assert.EqualValues(t, buntdb.ErrNotFound, err)

	assert.Nil(t, sm.SaveGroupInvitor(test.G1, test.UID1))

	assert.EqualValues(t, localdb.ErrKeyExist, sm.SaveGroupInvitor(test.G1, test.UID2))

	target, err := sm.PopGroupInvitor(test.G1)
	assert.Nil(t, err)
	assert.Equal(t, test.UID1, target)

	_, err = sm.PopGroupInvitor(test.G2)
	assert.EqualValues(t, buntdb.ErrNotFound, err)
}

func TestStateManager_IsMuted(t *testing.T) {
	test.InitBuntdb(t)
	defer test.CloseBuntdb(t)

	sm := newStateManager(t)
	assert.NotNil(t, sm)
	assert.False(t, sm.IsMuted(test.G1, test.UID1))
	assert.Nil(t, sm.Muted(test.G1, test.UID1, 999999))
	assert.True(t, sm.IsMuted(test.G1, test.UID1))
	assert.Nil(t, sm.Muted(test.G1, test.UID1, 0))
	assert.False(t, sm.IsMuted(test.G1, test.UID1))
	assert.Nil(t, sm.Muted(test.G1, 0, -1))
	assert.True(t, sm.IsMuted(test.G1, 0))
}

func TestStateManager_GetMessageImageUrl(t *testing.T) {
	test.InitBuntdb(t)
	defer test.CloseBuntdb(t)

	sm := newStateManager(t)
	assert.NotNil(t, sm)

	assert.Nil(t, sm.SaveMessageImageUrl(test.G1, test.MessageID1, []message.IMessageElement{}))
	assert.Len(t, sm.GetMessageImageUrl(test.G1, test.MessageID1), 0)

	assert.Nil(t, sm.SaveMessageImageUrl(test.G1, test.MessageID1, []message.IMessageElement{
		&message.GroupImageElement{
			Url: "image1",
		},
		&message.GroupImageElement{
			Url: "image2",
		},
	}))

	assert.Len(t, sm.GetMessageImageUrl(test.G1, test.MessageID1), 2)
}
