package concern

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestHook(t *testing.T) {
	var d defaultHook
	r1 := d.ShouldSendHook(nil)
	assert.False(t, r1.Pass)
	r2 := d.AtBeforeHook(nil)
	assert.False(t, r2.Pass)

	var hook = new(HookResult)
	hook.PassOrReason(true, "111")
	assert.True(t, hook.Pass)

	hook = new(HookResult)
	hook.PassOrReason(false, "222")
	assert.False(t, hook.Pass)
	assert.Equal(t, "222", hook.Reason)

	r := d.NewsFilterHook(nil)
	assert.False(t, r.Pass)
}
