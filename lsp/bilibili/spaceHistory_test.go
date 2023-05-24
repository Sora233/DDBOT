package bilibili

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDynamicSrvSpaceHistory(t *testing.T) {
	resp, err := DynamicSrvSpaceHistory(97505)
	assert.Nil(t, err)
	assert.Zero(t, resp.GetCode())
}
