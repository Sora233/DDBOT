package youtube

import (
	"github.com/Sora233/DDBOT/lsp/buntdb"
	"github.com/Sora233/DDBOT/lsp/concern"
	"github.com/Sora233/DDBOT/lsp/mmsg/mt"
)

type KeySet struct {
}

func (k *KeySet) AtAllMarkKey(keys ...interface{}) string {
	return buntdb.YoutubeAtAllMarkKey(keys...)
}

func (k *KeySet) ConcernConfigKey(keys ...interface{}) string {
	return buntdb.YoutubeConcernConfigKey(keys...)
}

func (k *KeySet) ConcernStateKey(keys ...interface{}) string {
	return buntdb.YoutubeConcernStateKey(keys...)
}

func (k *KeySet) FreshKey(keys ...interface{}) string {
	return buntdb.YoutubeFreshKey(keys...)
}

func (k *KeySet) ParseConcernStateKey(key string) (mt.Target, interface{}, error) {
	return concern.ParseConcernStateKeyWithString(key)
}

func NewKeySet() *KeySet {
	return new(KeySet)
}

type extraKey struct {
}

func (e *extraKey) UserInfoKey(keys ...interface{}) string {
	return buntdb.YoutubeUserInfoKey(keys...)
}

func (e *extraKey) InfoKey(keys ...interface{}) string {
	return buntdb.YoutubeInfoKey(keys...)
}

func (e *extraKey) VideoKey(keys ...interface{}) string {
	return buntdb.YoutubeVideoKey(keys...)
}

func NewExtraKey() *extraKey {
	return &extraKey{}
}
