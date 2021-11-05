package weibo

import "testing"

func TestExtraKey(t *testing.T) {
	var e extraKeySet
	e.NewsInfoKey()
	e.MarkMblogIdKey()
	e.UserInfoKey()
}
