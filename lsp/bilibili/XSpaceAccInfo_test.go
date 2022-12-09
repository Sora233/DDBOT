package bilibili

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestXSpaceAccInfo(t *testing.T) {
	resp, err := XSpaceAccInfo(97505)
	assert.Nil(t, err)
	assert.Zero(t, resp.GetCode())
}
