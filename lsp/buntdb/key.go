package buntdb

import (
	"errors"
	"strconv"
	"strings"
)

func Key(keys ...interface{}) string {
	var _keys []string
	for _, key := range keys {
		switch key.(type) {
		case string:
			_keys = append(_keys, key.(string))
		case int:
			_keys = append(_keys, strconv.FormatInt(int64(key.(int)), 10))
		case int32:
			_keys = append(_keys, strconv.FormatInt(int64(key.(int32)), 10))
		case int64:
			_keys = append(_keys, strconv.FormatInt(key.(int64), 10))
		}
	}
	return strings.Join(_keys, ":")
}

func NamedKey(name string, keys []interface{}) string {
	newkey := []interface{}{name}
	for _, key := range keys {
		newkey = append(newkey, key)
	}
	return Key(newkey...)
}

func BilibiliGroupConcernStateKey(keys ...interface{}) string {
	return NamedKey("ConcernState", keys)
}
func BilibiliAllConcernStateKey(keys ...interface{}) string {
	return NamedKey("BilibiliAllConcernState", keys)
}
func BilibliFreshKey(keys ...interface{}) string {
	return NamedKey("fresh", keys)
}
func BilibiliCurrentLiveKey(keys ...interface{}) string {
	return NamedKey("CurrentLive", keys)
}
func BilibiliUserInfoKey(keys ...interface{}) string {
	return NamedKey("UserInfo", keys)
}
func DouyuGroupConcernStateKey(keys ...interface{}) string {
	return NamedKey("DouyuConcernState", keys)
}
func DouyuAllConcernStateKey(keys ...interface{}) string {
	return NamedKey("DouyuAllConcernStateKey", keys)
}
func DouyuFreshKey(keys ...interface{}) string {
	return NamedKey("douyuFresh", keys)
}
func DouyuCurrentLiveKey(keys ...interface{}) string {
	return NamedKey("DouyuCurrentLive", keys)
}

func ParseConcernStateKey(key string) (groupCode int64, id int64, err error) {
	keys := strings.Split(key, ":")
	if len(keys) != 3 {
		return 0, 0, errors.New("invalid key")
	}
	groupCode, err = strconv.ParseInt(keys[1], 10, 64)
	if err != nil {
		return 0, 0, err
	}
	id, err = strconv.ParseInt(keys[2], 10, 64)
	if err != nil {
		return 0, 0, err
	}
	return groupCode, id, nil
}
func ParseCurrentLiveKey(key string) (id int64, err error) {
	keys := strings.Split(key, ":")
	if len(keys) != 2 {
		return 0, errors.New("invalid key")
	}
	id, err = strconv.ParseInt(keys[1], 10, 64)
	if err != nil {
		return 0, err
	}
	return id, nil
}
