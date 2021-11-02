package acfun

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestApiChannelList(t *testing.T) {
	resp, err := ApiChannelList(100, "")
	assert.Nil(t, err)
	assert.NotNil(t, resp)
}
