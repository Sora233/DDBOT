package weibo

import (
	jsoniter "github.com/json-iterator/go"
	"go.uber.org/atomic"

	"github.com/Sora233/DDBOT/v2/requests"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

const (
	Site = "weibo"
)

var (
	visitorCookiesOpt atomic.Value
)

func CookieOption() []requests.Option {
	if c := visitorCookiesOpt.Load(); c != nil {
		return c.([]requests.Option)
	}
	return nil
}
