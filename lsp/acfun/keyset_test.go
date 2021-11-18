package acfun

import "testing"

func TestExtraKeySet(t *testing.T) {
	var e extraKey
	e.LiveInfoKey()
	e.NotLiveKey()
	e.UserInfoKey()
	e.UidFirstTimestamp()
}
