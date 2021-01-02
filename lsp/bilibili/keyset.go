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

func (k *keySet) CurrentLiveKey(keys ...interface{}) string {
	return buntdb.BilibiliCurrentLiveKey(keys...)
}

func (k *keySet) ParseGroupConcernStateKey(key string) (int64, int64, error) {
	return buntdb.ParseConcernStateKey(key)
}
func (k *keySet) ParseCurrentLiveKey(key string) (int64, error) {
	return buntdb.ParseCurrentLiveKey(key)
}

type extraKey struct {
}

func (k *extraKey) UserInfoKey(keys ...interface{}) string {
	return buntdb.BilibiliUserInfoKey(keys...)
}

func NewKeySet() *keySet {
	return &keySet{}
}
func NewExtraKey() *extraKey {
	return &extraKey{}
}
