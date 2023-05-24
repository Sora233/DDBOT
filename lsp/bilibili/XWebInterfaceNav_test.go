package bilibili

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestXWebInterfaceNav(t *testing.T) {
	resp, err := XWebInterfaceNav(false)
	assert.Nil(t, err)
	assert.NotNil(t, resp.GetData().GetWbiImg().GetImgUrl())
	assert.NotNil(t, resp.GetData().GetWbiImg().GetSubUrl())
}
