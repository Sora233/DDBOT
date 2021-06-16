package concern_manager

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewStateManager(t *testing.T) {
	assert.NotNil(t, NewStateManager(nil, false))
}
