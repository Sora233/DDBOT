package lsp

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCheckValidCommand(t *testing.T) {
	assert.True(t, CheckValidCommand(WatchCommand))
	assert.False(t, CheckValidCommand(WatchCommand+"false"))
}

func TestCheckOperateableCommand(t *testing.T) {
	assert.True(t, CheckOperateableCommand(WatchCommand))
	assert.False(t, CheckOperateableCommand(EnableCommand))
}

func TestCombineCommand(t *testing.T) {
	assert.EqualValues(t, WatchCommand, CombineCommand(UnwatchCommand))
	assert.EqualValues(t, ListCommand, CombineCommand(ListCommand))
}
