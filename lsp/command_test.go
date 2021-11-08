package lsp

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCheckValidCommand(t *testing.T) {
	assert.True(t, CheckValidCommand("watch"))
	assert.False(t, CheckValidCommand("watchfalse"))
}

func TestCheckOperateableCommand(t *testing.T) {
	assert.True(t, CheckOperateableCommand("watch"))
	assert.False(t, CheckOperateableCommand("enable"))
}

func TestCombineCommand(t *testing.T) {
	assert.EqualValues(t, "watch", CombineCommand("unwatch"))
	assert.EqualValues(t, "list", CombineCommand("list"))
}
