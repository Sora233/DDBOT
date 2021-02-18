package buntdb

import (
	"errors"
	"reflect"
	"strconv"
	"strings"
)

func Key(keys ...interface{}) string {
	var _keys []string
	for _, ikey := range keys {
		switch key := ikey.(type) {
		case string:
			_keys = append(_keys, key)
		case int, int8, int16, int32, int64:
			_keys = append(_keys, strconv.FormatInt(reflect.ValueOf(key).Int(), 10))
		case uint, uint8, uint16, uint32, uint64:
			_keys = append(_keys, strconv.FormatUint(reflect.ValueOf(key).Uint(), 10))
		case bool:
			_keys = append(_keys, strconv.FormatBool(key))
		default:
			panic("unsupported key type " + reflect.ValueOf(ikey).Type().Name())
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
func BilibiliCurrentNewsKey(keys ...interface{}) string {
	return NamedKey("CurrentNews", keys)
}
func BilibiliUserInfoKey(keys ...interface{}) string {
	return NamedKey("UserInfo", keys)
}
func DouyuGroupConcernStateKey(keys ...interface{}) string {
	return NamedKey("DouyuConcernState", keys)
}
func DouyuAllConcernStateKey(keys ...interface{}) string {
	return NamedKey("DouyuAllConcernState", keys)
}
func DouyuFreshKey(keys ...interface{}) string {
	return NamedKey("douyuFresh", keys)
}
func DouyuCurrentLiveKey(keys ...interface{}) string {
	return NamedKey("DouyuCurrentLive", keys)
}
func YoutubeGroupConcernStateKey(keys ...interface{}) string {
	return NamedKey("YoutubeConcernState", keys)
}
func YoutubeAllConcernStateKey(keys ...interface{}) string {
	return NamedKey("YoutubeAllConcernState", keys)
}
func YoutubeFreshKey(keys ...interface{}) string {
	return NamedKey("youtubeFresh", keys)
}
func YoutubeUserInfoKey(keys ...interface{}) string {
	return NamedKey("YoutubeUserInfo", keys)
}
func YoutubeInfoKey(keys ...interface{}) string {
	return NamedKey("YoutubeInfo", keys)
}
func PermissionKey(keys ...interface{}) string {
	return NamedKey("Permission", keys)
}
func GroupPermissionKey(keys ...interface{}) string {
	return NamedKey("GroupPermission", keys)
}
func GroupEnabledKey(keys ...interface{}) string {
	return NamedKey("GroupEnable", keys)
}
func GroupMessageImageKey(keys ...interface{}) string {
	return NamedKey("GroupMessageImage", keys)
}
func GroupMuteKey(keys ...interface{}) string {
	return NamedKey("GroupMute", keys)
}
func GroupInvitorKey(keys ...interface{}) string {
	return NamedKey("GroupInventor", keys)
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
func ParseYoutubeConcernStateKey(key string) (groupCode int64, id string, err error) {
	keys := strings.Split(key, ":")
	if len(keys) != 3 {
		return 0, "", errors.New("invalid key")
	}
	groupCode, err = strconv.ParseInt(keys[1], 10, 64)
	if err != nil {
		return 0, "", err
	}
	return groupCode, keys[2], nil

}
