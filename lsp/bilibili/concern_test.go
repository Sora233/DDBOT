package bilibili

import (
	"github.com/Sora233/DDBOT/concern"
	"github.com/Sora233/DDBOT/lsp/test"
	"github.com/stretchr/testify/assert"
	"testing"
)

func initConcern(t *testing.T) *Concern {
	c := NewConcern(nil)
	assert.NotNil(t, c)
	c.StateManager.FreshIndex(test.G1, test.G2)
	assert.Nil(t, c.StateManager.Start())
	return c
}

func TestNewConcern(t *testing.T) {
	test.InitBuntdb(t)
	defer test.CloseBuntdb(t)

	c := initConcern(t)
	c.Stop()
}

func TestConcern_Remove(t *testing.T) {
	test.InitBuntdb(t)
	defer test.CloseBuntdb(t)

	c := initConcern(t)

	origUserInfo := NewUserInfo(test.UID1, test.ROOMID1, test.NAME1, "")
	assert.NotNil(t, origUserInfo)
	_, err := c.AddGroupConcern(test.G1, test.UID1, concern.BibiliLive)
	assert.Nil(t, err)

	_, err = c.Remove(test.G1, test.UID1, concern.BibiliLive)
	assert.Nil(t, err)
}

func TestConcern_ListWatching(t *testing.T) {
	test.InitBuntdb(t)
	defer test.CloseBuntdb(t)

	c := initConcern(t)

	origUserInfo := NewUserInfo(test.UID1, test.ROOMID1, test.NAME1, "")
	assert.NotNil(t, origUserInfo)
	assert.Nil(t, c.AddUserInfo(origUserInfo))
	_, err := c.AddGroupConcern(test.G1, test.UID1, concern.BibiliLive)
	assert.Nil(t, err)

	_, err = c.AddGroupConcern(test.G1, test.UID1, concern.BibiliLive)
	assert.Nil(t, err)

	userInfos, ctypes, err := c.ListWatching(test.G1, concern.BibiliLive)
	assert.Nil(t, err)
	assert.Len(t, userInfos, 1)
	assert.Len(t, ctypes, 1)
	assert.Equal(t, origUserInfo, userInfos[0])
	assert.Equal(t, concern.BibiliLive, ctypes[0])

	userInfos, ctypes, err = c.ListWatching(test.G1, concern.BilibiliNews)
	assert.Nil(t, err)
	assert.Len(t, userInfos, 0)
	assert.Len(t, ctypes, 0)

	userInfos, ctypes, err = c.ListWatching(test.G2, concern.BibiliLive)
	assert.Nil(t, err)
	assert.Len(t, userInfos, 0)
	assert.Len(t, ctypes, 0)
}

func TestConcern_FindUserLiving(t *testing.T) {
	test.InitBuntdb(t)
	defer test.CloseBuntdb(t)

	c := initConcern(t)

	origUserInfo := NewUserInfo(test.UID1, test.ROOMID1, test.NAME1, "")
	origLiveInfo := NewLiveInfo(origUserInfo, "", "", LiveStatus_Living)
	assert.NotNil(t, origLiveInfo)

	err := c.AddLiveInfo(origLiveInfo)
	assert.Nil(t, err)

	liveInfo, err := c.FindUserLiving(test.UID1, false)
	assert.Nil(t, err)
	assert.EqualValues(t, origLiveInfo, liveInfo)

	liveInfo, err = c.FindUserLiving(test.UID2, false)
	assert.NotNil(t, err)
	assert.Nil(t, liveInfo)
}

func TestConcern_FindUserNews(t *testing.T) {
	test.InitBuntdb(t)
	defer test.CloseBuntdb(t)

	c := initConcern(t)

	origUserInfo := NewUserInfo(test.UID1, test.ROOMID1, test.NAME1, "")
	origNewsInfo := NewNewsInfo(origUserInfo, test.DynamicID1, test.TIMESTAMP1)

	err := c.AddNewsInfo(origNewsInfo)
	assert.Nil(t, err)

	newsInfo, err := c.FindUserNews(test.UID1, false)
	assert.Nil(t, err)
	assert.EqualValues(t, origNewsInfo, newsInfo)

	newsInfo, err = c.FindUserNews(test.UID2, false)
	assert.NotNil(t, err)
	assert.Nil(t, newsInfo)

	assert.Nil(t, c.ClearUserNews(test.UID1))
	assert.NotNil(t, c.ClearUserNews(test.UID2))
}
