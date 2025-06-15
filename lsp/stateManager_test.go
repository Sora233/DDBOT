package lsp

import (
	"cmp"
	"slices"
	"sort"
	"strconv"
	"testing"

	"github.com/LagrangeDev/LagrangeGo/client/event"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/buntdb"

	"github.com/Sora233/DDBOT/v2/internal/test"
	localdb "github.com/Sora233/DDBOT/v2/lsp/buntdb"
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
	assert.Nil(t, sm.Muted(test.G1, test.UID1, -1))
}

func TestStateManager_GetCurrentMode(t *testing.T) {
	test.InitBuntdb(t)
	defer test.CloseBuntdb(t)

	sm := newStateManager(t)
	assert.NotNil(t, sm)

	assert.True(t, sm.IsPublicMode())
	assert.False(t, sm.IsPrivateMode())
	assert.False(t, sm.IsProtectMode())
	assert.Equal(t, PublicMode, sm.GetCurrentMode())

	assert.Nil(t, sm.SetMode(PublicMode))
	assert.True(t, sm.IsPublicMode())
	assert.False(t, sm.IsPrivateMode())
	assert.False(t, sm.IsProtectMode())
	assert.Equal(t, PublicMode, sm.GetCurrentMode())

	assert.Nil(t, sm.SetMode(PrivateMode))
	assert.False(t, sm.IsPublicMode())
	assert.True(t, sm.IsPrivateMode())
	assert.False(t, sm.IsProtectMode())
	assert.Equal(t, PrivateMode, sm.GetCurrentMode())

	assert.Nil(t, sm.SetMode(ProtectMode))
	assert.False(t, sm.IsPublicMode())
	assert.False(t, sm.IsPrivateMode())
	assert.True(t, sm.IsProtectMode())
	assert.Equal(t, ProtectMode, sm.GetCurrentMode())

	assert.Nil(t, sm.SetMode(PrivateMode))
	assert.NotNil(t, sm.SetMode("unknown"))
	assert.True(t, sm.IsPrivateMode())
	assert.Equal(t, PrivateMode, sm.GetCurrentMode())

	err := localdb.RWCoverTx(func(tx *buntdb.Tx) error {
		key := localdb.ModeKey()
		_, _, err := tx.Set(key, "wrong", nil)
		return err
	})
	assert.Nil(t, err)

	assert.Equal(t, PublicMode, sm.GetCurrentMode())
	assert.True(t, sm.IsPublicMode())
	assert.False(t, sm.IsPrivateMode())
	assert.False(t, sm.IsProtectMode())
}

func TestStateManager_GetNewFriendRequest(t *testing.T) {
	test.InitBuntdb(t)
	defer test.CloseBuntdb(t)

	sm := newStateManager(t)
	assert.NotNil(t, sm)

	_, err := sm.GetNewFriendRequest("0")
	assert.NotNil(t, err)

	var expected = []*event.NewFriendRequest{
		{
			SourceUin:  test.UID1,
			SourceNick: "uid1",
			Msg:        "test1",
			Source:     strconv.Itoa(test.ID1),
		},
		{
			SourceUin:  test.UID2,
			SourceNick: "uid2",
			Msg:        "test2",
			Source:     strconv.Itoa(test.ID2),
		},
	}

	for _, exp := range expected {
		err := sm.SaveNewFriendRequest(exp)
		assert.Nil(t, err)
		act, err := sm.GetNewFriendRequest(exp.Source)
		assert.Nil(t, err)
		assert.EqualValues(t, exp, act)
	}

	act, err := sm.ListNewFriendRequest()
	assert.Nil(t, err)
	slices.SortFunc(act, func(i, j *event.NewFriendRequest) int {
		return cmp.Compare(i.Source, j.Source)
	})
	assert.EqualValues(t, expected, act)

	assert.Nil(t, sm.DeleteNewFriendRequest(strconv.Itoa(test.ID1)))
	_, err = sm.GetNewFriendRequest(strconv.Itoa(test.ID1))
	assert.NotNil(t, err)

	act, err = sm.ListNewFriendRequest()
	assert.Nil(t, err)
	assert.Len(t, act, 1)
	assert.EqualValues(t, expected[1], act[0])

	assert.NotNil(t, sm.DeleteNewFriendRequest(strconv.Itoa(test.ID1)))
	assert.Nil(t, sm.DeleteNewFriendRequest(strconv.Itoa(test.ID2)))

	act, err = sm.ListNewFriendRequest()
	assert.Nil(t, err)
	assert.Empty(t, act)
}

func TestStateManager_GetGroupInvitedRequest(t *testing.T) {
	test.InitBuntdb(t)
	defer test.CloseBuntdb(t)

	sm := newStateManager(t)
	assert.NotNil(t, sm)

	_, err := sm.GetGroupInvitedRequest(0)
	assert.NotNil(t, err)

	var expected = []*event.GroupInvite{
		{
			RequestSeq:  test.ID1,
			InvitorUin:  test.UID1,
			InvitorNick: "uid1",
			GroupUin:    test.G1,
			GroupName:   "test1",
		},
		{
			RequestSeq:  test.ID2,
			InvitorUin:  test.UID2,
			InvitorNick: "uid2",
			GroupUin:    test.G2,
			GroupName:   "test2",
		},
	}
	for _, exp := range expected {
		err := sm.SaveGroupInvitedRequest(exp)
		assert.Nil(t, err)
		act, err := sm.GetGroupInvitedRequest(exp.RequestSeq)
		assert.Nil(t, err)
		assert.EqualValues(t, exp, act)
	}

	act, err := sm.ListGroupInvitedRequest()
	assert.Nil(t, err)
	sort.Slice(act, func(i, j int) bool {
		return act[i].RequestSeq < act[j].RequestSeq
	})
	assert.EqualValues(t, expected, act)

	assert.Nil(t, sm.DeleteGroupInvitedRequest(test.ID1))
	_, err = sm.GetGroupInvitedRequest(test.ID1)
	assert.NotNil(t, err)

	act, err = sm.ListGroupInvitedRequest()
	assert.Nil(t, err)
	assert.Len(t, act, 1)
	assert.EqualValues(t, expected[1], act[0])

	assert.NotNil(t, sm.DeleteGroupInvitedRequest(test.ID1))
	assert.Nil(t, sm.DeleteGroupInvitedRequest(test.ID2))

	act, err = sm.ListGroupInvitedRequest()
	assert.Nil(t, err)
	assert.Empty(t, act)
}
