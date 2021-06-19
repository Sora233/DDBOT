package lsp

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewStatus(t *testing.T) {
	s := NewStatus()
	assert.NotNil(t, s)
}
