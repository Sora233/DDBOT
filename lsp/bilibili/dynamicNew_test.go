package bilibili

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDynamicSrvDynamicNew(t *testing.T) {
	_, err := DynamicSvrDynamicNew()
	assert.NotNil(t, err)
}
