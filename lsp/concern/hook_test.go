package concern

import (
	"github.com/Sora233/DDBOT/internal/test"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestHook(t *testing.T) {
	var a = DefaultHook{}
	assert.True(t, a.ShouldSendHook(nil).Pass)
	assert.True(t, a.FilterHook(nil).Pass)
	assert.False(t, a.AtBeforeHook(nil).Pass)

	var b = HookResult{}
	b.PassOrReason(true, "")
	assert.True(t, b.Pass)
	var c = HookResult{}
	c.PassOrReason(false, test.NAME1)
	assert.False(t, c.Pass)
	assert.EqualValues(t, test.NAME1, c.Reason)
}
