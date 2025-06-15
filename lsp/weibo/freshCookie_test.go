package weibo

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	localutils "github.com/Sora233/DDBOT/v2/utils"
)

func TestFreshCookie(t *testing.T) {
	var cookies []*http.Cookie
	var err error
	localutils.Retry(5, time.Second, func() bool {
		cookies, err = FreshCookie()
		return err == nil
	})
	assert.Nil(t, err)
	assert.NotEmpty(t, cookies)
}
