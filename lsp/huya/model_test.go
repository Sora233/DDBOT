package huya

import (
	"github.com/Sora233/DDBOT/internal/test"
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
	notify := NewConcernLiveNotify(test.G1, l)
	assert.NotNil(t, notify)
	assert.NotNil(t, notify.Logger())
	assert.Equal(t, test.G1, notify.GetGroupCode())
	assert.Equal(t, test.NAME1, notify.GetUid())
	assert.Equal(t, Live, notify.Type())

	m := notify.ToMessage()
	assert.NotNil(t, m)

	notify.IsLiving = true
	m = notify.ToMessage()
	assert.NotNil(t, m)
}
