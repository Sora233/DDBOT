package bilibili

import (
	"github.com/Sora233/DDBOT/internal/test"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestModelNotify(t *testing.T) {
	liveNotify := newLiveInfo(test.UID1, true, false, false)
	liveNotify.GroupCode = test.G1
	m := liveNotify.ToMessage()
	assert.NotNil(t, m)

	assert.Equal(t, Site, liveNotify.Site())
	assert.NotNil(t, liveNotify.Logger())
	assert.NotNil(t, Live, liveNotify.Type())
	assert.Equal(t, test.G1, liveNotify.GetGroupCode())
	assert.Equal(t, test.UID1, liveNotify.GetUid())

	liveNotify.Status = LiveStatus_NoLiving
	m = liveNotify.ToMessage()
	assert.NotNil(t, m)

	newsNotify := newNewsInfo(test.UID1, DynamicDescType_TextOnly)[0]
	newsNotify.GroupCode = test.G2
	assert.NotNil(t, newsNotify)
	assert.NotNil(t, newsNotify.Logger())
	assert.Equal(t, Site, newsNotify.Site())
	assert.Equal(t, News, newsNotify.Type())
	assert.Equal(t, test.UID1, newsNotify.GetUid())
	assert.Equal(t, test.G2, newsNotify.GetGroupCode())
	m = newsNotify.ToMessage()
	assert.NotNil(t, m)

	notifies := newNewsInfo(test.UID1, DynamicDescType_WithVideo, DynamicDescType_WithImage,
		DynamicDescType_WithPost, DynamicDescType_WithMusic, DynamicDescType_WithSketch, DynamicDescType_WithLive,
		DynamicDescType_WithLiveV2, DynamicDescType_WithMiss)
	for _, notify := range notifies {
		notify.GroupCode = test.G2
		m = notify.ToMessage()
		assert.NotNil(t, m)
		notify.Card.Card = "{}"
		m = notify.ToMessage()
		assert.NotNil(t, m)
	}
}

func TestNewConcernLiveNotify(t *testing.T) {
	origUserInfo := NewUserInfo(test.UID1, test.ROOMID1, test.NAME1, "")
	origLiveInfo := NewLiveInfo(origUserInfo, "", "", LiveStatus_Living)
	notify := NewConcernLiveNotify(test.G1, origLiveInfo)
	assert.NotNil(t, notify)
}

func TestNewConcernNewsNotify(t *testing.T) {
	origUserInfo := NewUserInfo(test.UID1, test.ROOMID1, test.NAME1, "")
	origNewsInfo := NewNewsInfo(origUserInfo, test.DynamicID1, test.TIMESTAMP1)
	origNewsInfo.Cards = []*Card{{}}
	notify := NewConcernNewsNotify(test.G1, origNewsInfo, nil)
	assert.NotNil(t, notify)
}
