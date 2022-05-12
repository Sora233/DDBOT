package huya

import (
	"github.com/Sora233/DDBOT/lsp/buntdb"
	"github.com/Sora233/DDBOT/lsp/concern"
	"github.com/Sora233/DDBOT/lsp/mmsg/mt"
)

type keySet struct {
}

func (l *keySet) AtAllMarkKey(keys ...interface{}) string {
	return buntdb.HuyaGroupAtAllMarkKey(keys...)
}

func (l *keySet) ConcernConfigKey(keys ...interface{}) string {
	return buntdb.HuyaGroupConcernConfigKey(keys...)
}

func (l *keySet) ConcernStateKey(keys ...interface{}) string {
	return buntdb.HuyaGroupConcernStateKey(keys...)
}

func (l *keySet) FreshKey(keys ...interface{}) string {
	return buntdb.HuyaFreshKey(keys...)
}

func (l *keySet) ParseConcernStateKey(key string) (mt.Target, interface{}, error) {
	return concern.ParseConcernStateKeyWithString(key)
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
