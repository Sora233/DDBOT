package bilibili

import "github.com/Sora233/DDBOT/v2/lsp/buntdb"

type keySet struct {
}

func (k *keySet) GroupAtAllMarkKey(keys ...interface{}) string {
	return buntdb.BilibiliGroupAtAllMarkKey(keys...)
}

func (k *keySet) GroupConcernConfigKey(keys ...interface{}) string {
	return buntdb.BilibiliGroupConcernConfigKey(keys...)
}

func (k *keySet) GroupConcernStateKey(keys ...interface{}) string {
	return buntdb.BilibiliGroupConcernStateKey(keys...)
}

func (k *keySet) FreshKey(keys ...interface{}) string {
	return buntdb.BilibliFreshKey(keys...)
}

func (k *keySet) ParseGroupConcernStateKey(key string) (uint32, interface{}, error) {
	return buntdb.ParseConcernStateKeyWithInt64(key)
}

type extraKey struct {
}

func (k *extraKey) CurrentLiveKey(keys ...interface{}) string {
	return buntdb.BilibiliCurrentLiveKey(keys...)
}

func (k *extraKey) UserInfoKey(keys ...interface{}) string {
	return buntdb.BilibiliUserInfoKey(keys...)
}
func (k *extraKey) UserStatKey(keys ...interface{}) string {
	return buntdb.BilibiliUserStatKey(keys...)
}
func (k *extraKey) CurrentNewsKey(keys ...interface{}) string {
	return buntdb.BilibiliCurrentNewsKey(keys...)
}

func (k *extraKey) DynamicIdKey(keys ...interface{}) string {
	return buntdb.BilibiliDynamicIdKey(keys...)
}

func (k *extraKey) UidFirstTimestamp(keys ...interface{}) string {
	return buntdb.BilibiliUidFirstTimestampKey(keys...)
}

func (k *extraKey) NotLiveKey(keys ...interface{}) string {
	return buntdb.BilibiliNotLiveCountKey(keys...)
}

func (k *extraKey) LastFreshKey(keys ...interface{}) string {
	return buntdb.BilibiliLastFreshKey(keys...)
}

func (k *extraKey) CompactMarkKey(keys ...interface{}) string {
	return buntdb.BilibiliCompactMarkKey(keys...)
}

func (k *extraKey) NotifyMsgKey(keys ...interface{}) string {
	return buntdb.BilibiliNotifyMsgKey(keys...)
}

func (k *extraKey) ActiveTimestampKey(keys ...interface{}) string {
	return buntdb.BilibiliActiveTimestampKey(keys...)
}

func NewKeySet() *keySet {
	return &keySet{}
}
func NewExtraKey() *extraKey {
	return &extraKey{}
}
