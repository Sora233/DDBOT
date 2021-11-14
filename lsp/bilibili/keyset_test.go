package bilibili

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewKeySet(t *testing.T) {
	s := NewKeySet()
	assert.NotNil(t, s)
	e := NewExtraKey()
	assert.NotNil(t, e)
	s.GroupAtAllMarkKey()
}
