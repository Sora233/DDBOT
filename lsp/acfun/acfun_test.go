package acfun

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Sora233/DDBOT/v2/internal/test"
)

func TestAcfun(t *testing.T) {
	assert.NotEmpty(t, APath(PathApiChannelList))
	assert.NotEmpty(t, APath("api/channel/list"))
	assert.NotEmpty(t, LiveUrl(test.UID1))
}
