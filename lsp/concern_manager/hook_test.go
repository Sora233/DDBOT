package concern_manager

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestHook(t *testing.T) {
	var d defaultHook
	assert.False(t, d.ShouldSendHook(nil))
	assert.False(t, d.AtAllBeforeHook(nil))
}
