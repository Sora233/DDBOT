package weibo

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFreshCookie(t *testing.T) {
	cookies, err := FreshCookie()
	assert.Nil(t, err)
	assert.NotEmpty(t, cookies)
}
