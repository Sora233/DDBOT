package bilibili

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDynamicSvrDynamicHistory(t *testing.T) {
	_, err := DynamicSvrDynamicHistory("0")
	assert.NotNil(t, err)
}
