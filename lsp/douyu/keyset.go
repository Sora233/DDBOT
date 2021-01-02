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

func (l *keySet) CurrentLiveKey(keys ...interface{}) string {
	return buntdb.DouyuCurrentLiveKey(keys...)
}

func (l *keySet) ParseCurrentLiveKey(key string) (int64, error) {
	return buntdb.ParseCurrentLiveKey(key)
}

func (l *keySet) ParseGroupConcernStateKey(key string) (int64, int64, error) {
	return buntdb.ParseConcernStateKey(key)
}

func NewKeySet() *keySet {
	return &keySet{}
}
