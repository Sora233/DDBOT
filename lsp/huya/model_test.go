package huya

import (
	"github.com/Sora233/DDBOT/internal/test"
	"github.com/Sora233/DDBOT/lsp/mmsg/mt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLiveInfo(t *testing.T) {
	l := &LiveInfo{
		RoomId:   test.NAME1,
		Name:     test.NAME2,
		RoomName: test.NAME2,
	}
	assert.Equal(t, Site, l.Site())
	assert.Equal(t, test.NAME2, l.GetName())
	assert.Equal(t, Live, l.Type())
	notify := NewConcernLiveNotify(mt.NewGroupTarget(test.G1), l)
	assert.NotNil(t, notify)
	assert.NotNil(t, notify.Logger())
	assert.True(t, notify.GetTarget().Equal(mt.NewGroupTarget(test.G1)))
	assert.Equal(t, test.NAME1, notify.GetUid())
	assert.Equal(t, Live, notify.Type())

	m := notify.ToMessage()
	assert.NotNil(t, m)

	notify.IsLiving = true
	m = notify.ToMessage()
	assert.NotNil(t, m)
}
