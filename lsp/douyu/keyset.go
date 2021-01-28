package douyu

import "github.com/Sora233/Sora233-MiraiGo/lsp/buntdb"

type keySet struct {
}

func (l *keySet) GroupConcernStateKey(keys ...interface{}) string {
	return buntdb.DouyuGroupConcernStateKey(keys...)
}

func (l *keySet) ConcernStateKey(keys ...interface{}) string {
	return buntdb.DouyuAllConcernStateKey(keys...)
}

func (l *keySet) FreshKey(keys ...interface{}) string {
	return buntdb.DouyuFreshKey(keys...)
}

func (l *keySet) ParseGroupConcernStateKey(key string) (int64, interface{}, error) {
	return buntdb.ParseConcernStateKey(key)
}

type extraKey struct {
}

func (l *extraKey) CurrentLiveKey(keys ...interface{}) string {
	return buntdb.DouyuCurrentLiveKey(keys...)
}

func NewExtraKey() *extraKey {
	return &extraKey{}
}

func NewKeySet() *keySet {
	return &keySet{}
}
