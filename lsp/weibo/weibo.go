package weibo

import (
	"github.com/Sora233/DDBOT/requests"
	jsoniter "github.com/json-iterator/go"
	"go.uber.org/atomic"
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
