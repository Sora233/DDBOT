package permission

import "github.com/Sora233/Sora233-MiraiGo/lsp/buntdb"

type KeySet struct{}

func (k *KeySet) PermissionKey(keys ...interface{}) string {
	return buntdb.PermissionKey(keys...)
}

func (k *KeySet) GroupPermissionKey(keys ...interface{}) string {
	return buntdb.GroupPermissionKey(keys...)
}

func (k *KeySet) GroupEnabledKey(keys ...interface{}) string {
	return buntdb.GroupEnabledKey(keys...)
}

func NewKeySet() *KeySet {
	return &KeySet{}
}
