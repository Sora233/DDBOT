package acfun

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	localutils "github.com/Sora233/DDBOT/v2/utils"
)

func TestApiChannelList(t *testing.T) {
	var resp *ApiChannelListResponse
	var err error
	localutils.Retry(5, time.Second, func() bool {
		resp, err = ApiChannelList(100, "")
		return err == nil
	})
	assert.Nil(t, err)
	assert.NotNil(t, resp)
}
