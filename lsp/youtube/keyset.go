package youtube

import "github.com/Sora233/DDBOT/v2/lsp/buntdb"

type KeySet struct {
}

func (k *KeySet) GroupAtAllMarkKey(keys ...interface{}) string {
	return buntdb.YoutubeGroupAtAllMarkKey(keys...)
}

func (k *KeySet) GroupConcernConfigKey(keys ...interface{}) string {
	return buntdb.YoutubeGroupConcernConfigKey(keys...)
}

func (k *KeySet) GroupConcernStateKey(keys ...interface{}) string {
	return buntdb.YoutubeGroupConcernStateKey(keys...)
}

func (k *KeySet) FreshKey(keys ...interface{}) string {
	return buntdb.YoutubeFreshKey(keys...)
}

func (k *KeySet) ParseGroupConcernStateKey(key string) (uint32, interface{}, error) {
	return buntdb.ParseConcernStateKeyWithString(key)
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
