package acfun

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLivePage(t *testing.T) {
	result, err := LivePage(19372766)
	assert.Nil(t, err)
	assert.NotEmpty(t, result.GetLiveInfo().GetUser().GetName())
	assert.NotEmpty(t, result.GetLiveInfo().GetUser().GetHeadUrl())
}
