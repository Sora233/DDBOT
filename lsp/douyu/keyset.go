package douyu

import (
	"github.com/Sora233/DDBOT/lsp/buntdb"
	"github.com/Sora233/DDBOT/lsp/concern"
	"github.com/Sora233/DDBOT/lsp/mmsg/mt"
)

type keySet struct {
}

func (l *keySet) AtAllMarkKey(keys ...interface{}) string {
	return buntdb.DouyuGroupAtAllMarkKey(keys...)
}

func (l *keySet) ConcernConfigKey(keys ...interface{}) string {
	return buntdb.DouyuGroupConcernConfigKey(keys...)
}

func (l *keySet) ConcernStateKey(keys ...interface{}) string {
	return buntdb.DouyuGroupConcernStateKey(keys...)
}

func (l *keySet) FreshKey(keys ...interface{}) string {
	return buntdb.DouyuFreshKey(keys...)
}

func (l *keySet) ParseConcernStateKey(key string) (mt.Target, interface{}, error) {
	return concern.ParseConcernStateKeyWithInt64(key)
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
