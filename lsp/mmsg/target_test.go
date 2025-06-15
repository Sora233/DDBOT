package mmsg

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Sora233/DDBOT/v2/internal/test"
)

func TestTarget(t *testing.T) {
	pt := NewPrivateTarget(test.ID1)
	assert.True(t, pt.TargetType().IsPrivate())
	assert.False(t, pt.TargetType().IsGroup())
	assert.EqualValues(t, test.ID1, pt.TargetCode())

	gt := NewGroupTarget(test.ID2)
	assert.True(t, gt.TargetType().IsGroup())
	assert.False(t, gt.TargetType().IsPrivate())
	assert.EqualValues(t, test.ID2, gt.TargetCode())
}
