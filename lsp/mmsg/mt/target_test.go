package mt

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTarget(t *testing.T) {
	pt := NewPrivateTarget(1)
	assert.True(t, pt.IsPrivate())
	assert.False(t, pt.IsGroup())
	assert.False(t, pt.IsGulid())
	assert.EqualValues(t, 1, pt.TargetCode())

	gt := NewGroupTarget(2)
	assert.True(t, gt.IsGroup())
	assert.False(t, gt.IsPrivate())
	assert.False(t, gt.IsGulid())
	assert.EqualValues(t, 2, gt.TargetCode())

	gut := NewGulidTarget(3, 3)
	assert.True(t, gut.IsGulid())
	assert.False(t, gut.IsPrivate())
	assert.False(t, gut.IsGroup())

	assert.False(t, gut.Equal(gt))
	assert.False(t, gut.Equal(pt))
	assert.False(t, gt.Equal(pt))

	assert.False(t, gut.Equal(NewGulidTarget(3, 4)))
	assert.False(t, gut.Equal(NewGulidTarget(4, 3)))
	assert.True(t, gut.Equal(NewGulidTarget(3, 3)))

	assert.False(t, pt.Equal(NewPrivateTarget(2)))
	assert.False(t, pt.Equal(NewPrivateTarget(3)))
	assert.True(t, pt.Equal(NewPrivateTarget(1)))

	assert.False(t, gt.Equal(NewGroupTarget(1)))
	assert.False(t, gt.Equal(NewGroupTarget(3)))
	assert.True(t, gt.Equal(NewGroupTarget(2)))

}
