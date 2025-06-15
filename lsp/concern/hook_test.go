package concern

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Sora233/DDBOT/v2/internal/test"
)

func TestHook(t *testing.T) {
	var b = HookResult{}
	b.PassOrReason(true, "")
	assert.True(t, b.Pass)
	var c = HookResult{}
	c.PassOrReason(false, test.NAME1)
	assert.False(t, c.Pass)
	assert.EqualValues(t, test.NAME1, c.Reason)
}
