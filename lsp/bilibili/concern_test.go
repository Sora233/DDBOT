package bilibili

import (
	"github.com/Sora233/DDBOT/internal/test"
	"github.com/Sora233/DDBOT/lsp/concern"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/buntdb"
	"testing"
	"time"
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
	_, err := c.AddGroupConcern(test.G1, test.UID1, test.BibiliLive)
	assert.Nil(t, err)

	_, err = c.Remove(nil, test.G1, test.UID1, test.BibiliLive)
	assert.Nil(t, err)
}

func TestConcern_List(t *testing.T) {
	test.InitBuntdb(t)
	defer test.CloseBuntdb(t)

	c := initConcern(t)

	origUserInfo := NewUserInfo(test.UID1, test.ROOMID1, test.NAME1, "")
	assert.NotNil(t, origUserInfo)
	assert.Nil(t, c.AddUserInfo(origUserInfo))
	_, err := c.AddGroupConcern(test.G1, test.UID1, test.BibiliLive)
	assert.Nil(t, err)

	_, err = c.AddGroupConcern(test.G1, test.UID1, test.BibiliLive)
	assert.Nil(t, err)

	userInfos, ctypes, err := c.List(test.G1, test.BibiliLive)
	assert.Nil(t, err)
	assert.Len(t, userInfos, 1)
	assert.Len(t, ctypes, 1)
	assert.Equal(t, concern.NewIdentity(origUserInfo.Mid, origUserInfo.Name), userInfos[0])
	assert.Equal(t, test.BibiliLive, ctypes[0])

	userInfos, ctypes, err = c.List(test.G1, test.BilibiliNews)
	assert.Nil(t, err)
	assert.Len(t, userInfos, 0)
	assert.Len(t, ctypes, 0)

	userInfos, ctypes, err = c.List(test.G2, test.BibiliLive)
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

	const testMid int64 = 2

	userInfo, err := c.FindOrLoadUser(testMid)
	assert.Nil(t, err)
	assert.Equal(t, testMid, userInfo.Mid)
	assert.Equal(t, "碧诗", userInfo.Name)

	_, err = c.FindUserLiving(testMid, false)
	assert.Equal(t, buntdb.ErrNotFound, err)

	liveInfo, err = c.FindUserLiving(testMid, true)
	assert.Nil(t, err)
	assert.NotNil(t, liveInfo)
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

	const testMid int64 = 2

	userInfo, err := c.FindOrLoadUser(testMid)
	assert.Nil(t, err)
	assert.Equal(t, testMid, userInfo.Mid)
	assert.Equal(t, "碧诗", userInfo.Name)

	userInfo2, err := c.FindOrLoadUser(testMid)
	assert.Nil(t, err)
	assert.NotNil(t, userInfo2)
	assert.EqualValues(t, userInfo, userInfo2)

	newsInfo, err = c.FindUserNews(testMid, false)
	assert.Equal(t, buntdb.ErrNotFound, err)

	newsInfo, err = c.FindUserNews(testMid, true)
	assert.Nil(t, err)
	assert.NotNil(t, newsInfo)
}

func TestConcern_StatUserWithCache(t *testing.T) {
	test.InitBuntdb(t)
	defer test.CloseBuntdb(t)

	c := initConcern(t)

	const testMid int64 = 2

	userInfo, err := c.FindOrLoadUser(testMid)
	assert.Nil(t, err)
	assert.Equal(t, testMid, userInfo.Mid)
	assert.Equal(t, "碧诗", userInfo.Name)

	stat, err := c.StatUserWithCache(testMid, time.Hour)
	assert.Nil(t, err)
	assert.NotNil(t, stat)
	assert.Equal(t, testMid, stat.Mid)
	assert.True(t, stat.Follower > 0)

	stat2, err := c.StatUserWithCache(testMid, time.Hour)
	assert.Nil(t, err)
	assert.NotNil(t, stat2)
	assert.EqualValues(t, stat, stat2)
}
