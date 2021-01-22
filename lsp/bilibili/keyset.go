package bilibili

import "github.com/Sora233/Sora233-MiraiGo/lsp/buntdb"

type keySet struct {
}

func (k *keySet) GroupConcernStateKey(keys ...interface{}) string {
	return buntdb.BilibiliGroupConcernStateKey(keys...)
}

func (k *keySet) ConcernStateKey(keys ...interface{}) string {
	return buntdb.BilibiliAllConcernStateKey(keys...)
}

func (k *keySet) FreshKey(keys ...interface{}) string {
	return buntdb.BilibliFreshKey(keys...)
}

func (k *keySet) ParseGroupConcernStateKey(key string) (int64, interface{}, error) {
	return buntdb.ParseConcernStateKey(key)
}

type extraKey struct {
}

func (k *extraKey) CurrentLiveKey(keys ...interface{}) string {
	return buntdb.BilibiliCurrentLiveKey(keys...)
}

func (k *extraKey) UserInfoKey(keys ...interface{}) string {
	return buntdb.BilibiliUserInfoKey(keys...)
}
func (k *extraKey) CurrentNewsKey(keys ...interface{}) string {
	return buntdb.BilibiliCurrentNewsKey(keys...)
}

func NewKeySet() *keySet {
	return &keySet{}
}
func NewExtraKey() *extraKey {
	return &extraKey{}
}
