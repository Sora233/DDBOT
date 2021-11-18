package bilibili

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetAttentionList(t *testing.T) {
	_, err := GetAttentionList()
	assert.NotNil(t, err)
}
