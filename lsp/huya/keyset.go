package huya

import "github.com/Sora233/DDBOT/v2/lsp/buntdb"

type keySet struct {
}

func (l *keySet) GroupAtAllMarkKey(keys ...interface{}) string {
	return buntdb.HuyaGroupAtAllMarkKey(keys...)
}

func (l *keySet) GroupConcernConfigKey(keys ...interface{}) string {
	return buntdb.HuyaGroupConcernConfigKey(keys...)
}

func (l *keySet) GroupConcernStateKey(keys ...interface{}) string {
	return buntdb.HuyaGroupConcernStateKey(keys...)
}

func (l *keySet) FreshKey(keys ...interface{}) string {
	return buntdb.HuyaFreshKey(keys...)
}

func (l *keySet) ParseGroupConcernStateKey(key string) (uint32, interface{}, error) {
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
