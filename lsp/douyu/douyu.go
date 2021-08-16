package douyu

import (
	jsoniter "github.com/json-iterator/go"
	"strconv"
)

const (
	Site = "douyu"
	Host = "https://www.douyu.com"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func DouyuPath(path string) string {
	return Host + path
}

func ParseUid(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}
