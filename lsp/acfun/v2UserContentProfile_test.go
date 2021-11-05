package acfun

import (
	localutils "github.com/Sora233/DDBOT/utils"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestV2UserContentProfile(t *testing.T) {
	var resp *V2UserContentProfileResponse
	var err error
	localutils.Retry(5, time.Second, func() bool {
		resp, err = V2UserContentProfile(1)
		return err == nil
	})
	assert.Nil(t, err)
	assert.Zero(t, resp.GetErrorid())
	assert.EqualValues(t, 1, resp.GetVdata().GetUserId())
	assert.EqualValues(t, "admin", resp.GetVdata().GetUsername())
}
