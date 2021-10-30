package bilibili

import (
	"github.com/Sora233/DDBOT/internal/test"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConcernLiveNotify_ToMessage(t *testing.T) {
	notify := newLiveInfo(test.UID1, true, false, false)
	m := notify.ToMessage()
	assert.NotNil(t, m)

	notify.Status = LiveStatus_NoLiving
	m = notify.ToMessage()
	assert.NotNil(t, m)
}
