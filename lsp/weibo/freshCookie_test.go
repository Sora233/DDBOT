package weibo

import (
	localutils "github.com/Sora233/DDBOT/utils"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
	"time"
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
