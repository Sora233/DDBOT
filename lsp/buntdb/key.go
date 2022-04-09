package buntdb

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type KeyPatternFunc func(...interface{}) string

func Key(keys ...interface{}) string {
	var _keys []string
	for _, ikey := range keys {
		rk := reflect.ValueOf(ikey)
		if !rk.IsValid() {
			panic(fmt.Sprintf("invalid value %T %v", ikey, ikey))
		}
		if rk.Kind() == reflect.Ptr || rk.Kind() == reflect.Interface {
			rk = rk.Elem()
		}
		switch rk.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			_keys = append(_keys, strconv.FormatInt(rk.Int(), 10))
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			_keys = append(_keys, strconv.FormatUint(rk.Uint(), 10))
		case reflect.String:
			_keys = append(_keys, rk.String())
		case reflect.Bool:
			_keys = append(_keys, strconv.FormatBool(rk.Bool()))
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
func BilibiliGroupConcernConfigKey(keys ...interface{}) string {
	return NamedKey("ConcernConfig", keys)
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
func BilibiliDynamicIdKey(keys ...interface{}) string {
	return NamedKey("DynamicId", keys)
}
func BilibiliUidFirstTimestampKey(keys ...interface{}) string {
	return NamedKey("UidFirstTimestamp", keys)
}
func BilibiliUserCookieInfoKey(keys ...interface{}) string {
	return NamedKey("UserCookieInfo", keys)
}
func BilibiliNotLiveCountKey(keys ...interface{}) string {
	return NamedKey("NotLiveCount", keys)
}
func BilibiliUserInfoKey(keys ...interface{}) string {
	return NamedKey("UserInfo", keys)
}
func BilibiliUserStatKey(keys ...interface{}) string {
	return NamedKey("UserStat", keys)
}
func BilibiliGroupAtAllMarkKey(keys ...interface{}) string {
	return NamedKey("GroupAtAll", keys)
}
func BilibiliCompactMarkKey(keys ...interface{}) string {
	return NamedKey("CompactMark", keys)
}
func BilibiliNotifyMsgKey(keys ...interface{}) string {
	return NamedKey("NotifyMsg", keys)
}
func BilibiliActiveTimestampKey(keys ...interface{}) string {
	return NamedKey("ActiveTimestamp", keys)
}
func BilibiliLastFreshKey(keys ...interface{}) string {
	return NamedKey("BilibiliLastFresh", keys)
}
func DouyuGroupConcernStateKey(keys ...interface{}) string {
	return NamedKey("DouyuConcernState", keys)
}
func DouyuGroupConcernConfigKey(keys ...interface{}) string {
	return NamedKey("DouyuConcernConfig", keys)
}
func DouyuFreshKey(keys ...interface{}) string {
	return NamedKey("douyuFresh", keys)
}
func DouyuCurrentLiveKey(keys ...interface{}) string {
	return NamedKey("DouyuCurrentLive", keys)
}
func DouyuGroupAtAllMarkKey(keys ...interface{}) string {
	return NamedKey("DouyuGroupAtAll", keys)
}
func YoutubeGroupConcernStateKey(keys ...interface{}) string {
	return NamedKey("YoutubeConcernState", keys)
}
func YoutubeGroupConcernConfigKey(keys ...interface{}) string {
	return NamedKey("YoutubeConcernConfig", keys)
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
func YoutubeVideoKey(keys ...interface{}) string {
	return NamedKey("YoutubeVideo", keys)
}
func YoutubeGroupAtAllMarkKey(keys ...interface{}) string {
	return NamedKey("YoutubeGroupAtAll", keys)
}
func HuyaGroupConcernStateKey(keys ...interface{}) string {
	return NamedKey("HuyaConcernState", keys)
}
func HuyaGroupConcernConfigKey(keys ...interface{}) string {
	return NamedKey("HuyaConcernConfig", keys)
}
func HuyaFreshKey(keys ...interface{}) string {
	return NamedKey("huyaFresh", keys)
}
func HuyaCurrentLiveKey(keys ...interface{}) string {
	return NamedKey("HuyaCurrentLive", keys)
}
func HuyaGroupAtAllMarkKey(keys ...interface{}) string {
	return NamedKey("HuyaGroupAtAll", keys)
}
func AcfunUserInfoKey(keys ...interface{}) string {
	return NamedKey("AcfunUserInfo", keys)
}
func AcfunLiveInfoKey(keys ...interface{}) string {
	return NamedKey("AcfunLiveInfo", keys)
}
func AcfunNotLiveKey(keys ...interface{}) string {
	return NamedKey("AcfunNotLiveCount", keys)
}
func AcfunUidFirstTimestampKey(keys ...interface{}) string {
	return NamedKey("AcfunUidFirstTimestamp", keys)
}
func WeiboUserInfoKey(keys ...interface{}) string {
	return NamedKey("WeiboUserInfo", keys)
}
func WeiboNewsInfoKey(keys ...interface{}) string {
	return NamedKey("WeiboNewsInfo", keys)
}
func WeiboMarkMblogIdKey(keys ...interface{}) string {
	return NamedKey("WeiboMarkMblogId", keys)
}

func PermissionKey(keys ...interface{}) string {
	return NamedKey("Permission", keys)
}
func BlockListKey(keys ...interface{}) string {
	return NamedKey("BlockList", keys)
}
func GroupPermissionKey(keys ...interface{}) string {
	return NamedKey("GroupPermission", keys)
}
func GroupEnabledKey(keys ...interface{}) string {
	return NamedKey("GroupEnable", keys)
}
func GlobalEnabledKey(keys ...interface{}) string {
	return NamedKey("GlobalEnable", keys)
}
func GroupMessageImageKey(keys ...interface{}) string {
	return NamedKey("GroupMessageImage", keys)
}
func GroupSilenceKey(keys ...interface{}) string {
	return NamedKey("GroupSilence", keys)
}
func GlobalSilenceKey(keys ...interface{}) string {
	return NamedKey("GlobalSilence", keys)
}
func GroupMuteKey(keys ...interface{}) string {
	return NamedKey("GroupMute", keys)
}
func GroupInvitorKey(keys ...interface{}) string {
	return NamedKey("GroupInventor", keys)
}

func LoliconPoolStoreKey(keys ...interface{}) string {
	return NamedKey("LoliconPoolStore", keys)
}

func ImageCacheKey(keys ...interface{}) string {
	return NamedKey("ImageCache", keys)
}

func ModeKey() string {
	return NamedKey("Mode", nil)
}
func NewFriendRequestKey(keys ...interface{}) string {
	return NamedKey("NewFriendRequest", keys)
}
func GroupInvitedKey(keys ...interface{}) string {
	return NamedKey("GroupInvited", keys)
}

func VersionKey(keys ...interface{}) string {
	return NamedKey("Version", keys)
}

func DDBotReleaseKey(keys ...interface{}) string {
	return NamedKey("DDBotReleaseKey", keys)
}

func DDBotNoUpdateKey(keys ...interface{}) string {
	return NamedKey("DDBotNoUpdateKey", keys)
}

func ParseConcernStateKeyWithInt64(key string) (groupCode int64, id int64, err error) {
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
func ParseConcernStateKeyWithString(key string) (groupCode int64, id string, err error) {
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
