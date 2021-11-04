package acfun

import (
	localutils "github.com/Sora233/DDBOT/utils"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
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
