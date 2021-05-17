package huya

const (
	Site = "huya"
	Host = "https://www.huya.com"
)

func HuyaPath(path string) string {
	return Host + "/" + path
}
