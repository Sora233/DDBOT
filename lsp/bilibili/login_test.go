package bilibili

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLogin(t *testing.T) {
	resp, err := Login("stupidusername", "wrong")
	assert.Nil(t, err)
	assert.NotZero(t, resp.GetCode())
}
