package douyu

import "strconv"

const (
	Site = "douyu"
	Host = "https://www.douyu.com"
)

func DouyuPath(path string) string {
	return Host + path
}

func ParseUid(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}
