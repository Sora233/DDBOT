package acfun

import (
	jsoniter "github.com/json-iterator/go"
	"strconv"
	"strings"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

const (
	Site     = "acfun"
	BaseHost = "https://live.acfun.cn"
)

var BasePath = map[string]string{
	PathApiChannelList: BaseHost,
}

func APath(path string) string {
	if strings.HasPrefix(path, "/") {
		return BasePath[path] + path
	} else {
		return BasePath[path] + "/" + path
	}
}

func LiveUrl(uid int64) string {
	return "https://live.acfun.cn/live/" + strconv.FormatInt(uid, 10)
}
