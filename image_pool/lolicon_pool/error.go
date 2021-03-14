package lolicon_pool

import "fmt"

var (
	ErrNotFound    = fmt.Errorf("没有符合条件的图片")
	ErrAPIKeyError = fmt.Errorf("APIKEY 不存在或被封禁")
	ErrQuotaExceed = fmt.Errorf("达到调用额度限制")
)
