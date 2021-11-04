package weibo

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestApiContainerGetIndexCards(t *testing.T) {
	resp, err := ApiContainerGetIndexCards(5462373877)
	assert.Nil(t, err)
	assert.NotZero(t, resp.GetOk())
}
