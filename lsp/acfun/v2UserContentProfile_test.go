package acfun

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestV2UserContentProfile(t *testing.T) {
	resp, err := V2UserContentProfile(1)
	assert.Nil(t, err)
	assert.Zero(t, resp.GetErrorid())
	assert.EqualValues(t, 1, resp.GetVdata().GetUserId())
	assert.EqualValues(t, "admin", resp.GetVdata().GetUsername())
}
