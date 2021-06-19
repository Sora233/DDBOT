package lsp

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewStateManager(t *testing.T) {
	sm := NewStateManager()
	assert.NotNil(t, sm)
}
