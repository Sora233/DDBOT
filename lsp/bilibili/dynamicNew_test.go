package bilibili

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDynamicSrvDynamicNew(t *testing.T) {
	_, err := DynamicSrvDynamicNew()
	assert.NotNil(t, err)
}
