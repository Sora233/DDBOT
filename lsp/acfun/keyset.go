package acfun

import localdb "github.com/Sora233/DDBOT/lsp/buntdb"

type extraKey struct{}

func (e *extraKey) UserInfoKey(keys ...interface{}) string {
	return localdb.AcfunUserInfoKey(keys...)
}

func (e *extraKey) LiveInfoKey(keys ...interface{}) string {
	return localdb.AcfunLiveInfoKey(keys...)
}

func (e *extraKey) NotLiveKey(keys ...interface{}) string {
	return localdb.AcfunNotLiveKey(keys...)
}

func (e *extraKey) UidFirstTimestamp(keys ...interface{}) string {
	return localdb.AcfunUidFirstTimestampKey(keys...)
}
