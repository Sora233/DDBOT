package douyu

const (
	Site = "douyu"
	Host = "https://www.douyu.com"
)

func DouyuPath(path string) string {
	return Host + path
}
