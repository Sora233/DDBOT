package permission

import localdb "github.com/Sora233/DDBOT/lsp/buntdb"

type KeySet struct{}

func (k *KeySet) PermissionKey(keys ...interface{}) string {
	return localdb.PermissionKey(keys...)
}

func (k *KeySet) TargetPermissionKey(keys ...interface{}) string {
	return localdb.GroupPermissionKey(keys...)
}

func (k *KeySet) TargetEnabledKey(keys ...interface{}) string {
	return localdb.GroupEnabledKey(keys...)
}

func (k *KeySet) GlobalEnabledKey(keys ...interface{}) string {
	return localdb.GlobalEnabledKey(keys...)
}

func (k *KeySet) TargetSilenceKey(keys ...interface{}) string {
	return localdb.GroupSilenceKey(keys...)
}

func (k *KeySet) GlobalSilenceKey(keys ...interface{}) string {
	return localdb.GlobalSilenceKey(keys...)
}

func (k *KeySet) BlockListKey(keys ...interface{}) string {
	return localdb.BlockListKey(keys...)
}

func NewKeySet() *KeySet {
	return &KeySet{}
}
