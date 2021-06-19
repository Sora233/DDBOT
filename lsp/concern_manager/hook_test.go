package concern_manager

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
}
