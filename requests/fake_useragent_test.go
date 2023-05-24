package requests

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRandomUA(t *testing.T) {
	ua := RandomUA(Computer)
	assert.NotEmpty(t, ua)
	assert.NotEqual(t, ua, defaultUA)
}
