package mt

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTarget(t *testing.T) {
	pt := NewPrivateTarget(1)
	assert.True(t, pt.IsPrivate())
	assert.False(t, pt.IsGroup())
	assert.False(t, pt.IsGuild())
	assert.EqualValues(t, 1, pt.TargetCode())

	gt := NewGroupTarget(2)
	assert.True(t, gt.IsGroup())
	assert.False(t, gt.IsPrivate())
	assert.False(t, gt.IsGuild())
	assert.EqualValues(t, 2, gt.TargetCode())

	gut := NewGuildTarget(3, 3)
	assert.True(t, gut.IsGuild())
	assert.False(t, gut.IsPrivate())
	assert.False(t, gut.IsGroup())

	assert.False(t, gut.Equal(gt))
	assert.False(t, gut.Equal(pt))
	assert.False(t, gt.Equal(pt))

	assert.False(t, gut.Equal(NewGuildTarget(3, 4)))
	assert.False(t, gut.Equal(NewGuildTarget(4, 3)))
	assert.True(t, gut.Equal(NewGuildTarget(3, 3)))

	assert.False(t, pt.Equal(NewPrivateTarget(2)))
	assert.False(t, pt.Equal(NewPrivateTarget(3)))
	assert.True(t, pt.Equal(NewPrivateTarget(1)))

	assert.False(t, gt.Equal(NewGroupTarget(1)))
	assert.False(t, gt.Equal(NewGroupTarget(3)))
	assert.True(t, gt.Equal(NewGroupTarget(2)))

}
