package acfun

import (
	"fmt"
	"github.com/Sora233/DDBOT/requests"
	jsoniter "github.com/json-iterator/go"
	"strings"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

const (
	Site         = "acfun"
	BaseHost     = "https://live.acfun.cn"
	AppAcfunHost = "https://apipc.app.acfun.cn"
)

var BasePath = map[string]string{
	PathApiChannelList:       BaseHost,
	PathV2UserContentProfile: AppAcfunHost,
}

func APath(path string) string {
	if strings.HasPrefix(path, "/") {
		return BasePath[path] + path
	} else {
		return BasePath[path] + "/" + path
	}
}

func LiveUrl(uid int64) string {
	return fmt.Sprintf("https://live.acfun.cn/live/%v", uid)
}

func AcfunHeaderOption() []requests.Option {
	return []requests.Option{
		requests.HeaderOption("deviceType", "1"),
		requests.HeaderOption("appVersion", "5.0.0"),
	}
}
