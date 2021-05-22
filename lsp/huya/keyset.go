package huya

import "github.com/Sora233/DDBOT/lsp/buntdb"

type keySet struct {
}

func (l *keySet) GroupConcernStateKey(keys ...interface{}) string {
	return buntdb.HuyaGroupConcernStateKey(keys...)
}

func (l *keySet) FreshKey(keys ...interface{}) string {
	return buntdb.HuyaFreshKey(keys...)
}

func (l *keySet) ParseGroupConcernStateKey(key string) (int64, interface{}, error) {
	return buntdb.ParseConcernStateKeyWithString(key)
}

type extraKey struct{}

func (k extraKey) CurrentLiveKey(keys ...interface{}) string {
	return buntdb.HuyaCurrentLiveKey(keys...)
}

func NewExtraKey() *extraKey {
	return &extraKey{}
}

func NewKeySet() *keySet {
	return &keySet{}
}