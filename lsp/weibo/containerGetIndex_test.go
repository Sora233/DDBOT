package weibo

import (
	localutils "github.com/Sora233/DDBOT/utils"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestApiContainerGetIndexCards(t *testing.T) {
	var resp *ApiContainerGetIndexCardsResponse
	var err error
	localutils.Retry(5, time.Second, func() bool {
		resp, err = ApiContainerGetIndexCards(5462373877)
		return err == nil
	})
	assert.Nil(t, err)
	assert.NotZero(t, resp.GetOk())
}

func TestApiContainerGetIndexProfile(t *testing.T) {
	var resp *ApiContainerGetIndexProfileResponse
	var err error
	localutils.Retry(5, time.Second, func() bool {
		resp, err = ApiContainerGetIndexProfile(5462373877)
		return err == nil
	})
	assert.Nil(t, err)
	assert.NotZero(t, resp.GetOk())
}
