package youtube

import "github.com/Sora233/Sora233-MiraiGo/lsp/buntdb"

type KeySet struct {
}

func (k *KeySet) GroupConcernStateKey(keys ...interface{}) string {
	return buntdb.YoutubeGroupConcernStateKey(keys...)
}

func (k *KeySet) ConcernStateKey(keys ...interface{}) string {
	return buntdb.YoutubeAllConcernStateKey(keys...)
}

func (k *KeySet) FreshKey(keys ...interface{}) string {
	return buntdb.YoutubeFreshKey(keys...)
}

func (k *KeySet) ParseGroupConcernStateKey(key string) (int64, interface{}, error) {
	return buntdb.ParseYoutubeConcernStateKey(key)
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

func NewExtraKey() *extraKey {
	return &extraKey{}
}
