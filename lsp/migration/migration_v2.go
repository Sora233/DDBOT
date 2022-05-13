package migration

import (
	"errors"
	localdb "github.com/Sora233/DDBOT/lsp/buntdb"
	"github.com/Sora233/DDBOT/lsp/concern"
	"github.com/Sora233/DDBOT/lsp/mmsg/mt"
	"github.com/Sora233/DDBOT/lsp/version"
	"strconv"
	"strings"
)

type V2 struct {
}

var errInvalidKey = errors.New("invalid key")

func (v *V2) v2ParsePrefixGroupPattern(key string) (groupCode int64, remains string, err error) {
	keys := strings.Split(key, ":")
	groupCode, err = strconv.ParseInt(keys[1], 10, 64)
	if err != nil {
		return 0, "", err
	}
	return groupCode, strings.Join(keys[2:], ":"), nil
}

func (v *V2) genPrefixGroupPatternMigrate(pattern localdb.KeyPatternFunc) func(string, string) (string, string) {
	return func(key string, value string) (string, string) {
		groupCode, remains, err := v.v2ParsePrefixGroupPattern(key)
		if err != nil {
			logger.WithField("key", key).Errorf("v2ParsePrefixGroupPattern error")
			return key, value
		}
		return pattern(mt.NewGroupTarget(groupCode), remains), value
	}
}

func (v *V2) Func() version.MigrationFunc {

	var f []version.MigrationFunc
	var acfunKeyset = concern.NewPrefixKeySetWithInt64ID("Acfun")
	var weiboKeyset = concern.NewPrefixKeySetWithStringID("weibo")
	var tcKeyset = concern.NewPrefixKeySetWithStringID("twitcasting")
	for _, pattern := range []localdb.KeyPatternFunc{
		localdb.BilibiliConcernStateKey,
		localdb.BilibiliConcernConfigKey,
		localdb.BilibiliAtAllMarkKey,
		localdb.BilibiliNotifyMsgKey,

		localdb.DouyuConcernStateKey,
		localdb.DouyuConcernConfigKey,
		localdb.DouyuAtAllMarkKey,

		localdb.YoutubeConcernStateKey,
		localdb.YoutubeConcernConfigKey,
		localdb.YoutubeAtAllMarkKey,

		localdb.HuyaConcernStateKey,
		localdb.HuyaConcernConfigKey,
		localdb.HuyaAtAllMarkKey,

		acfunKeyset.AtAllMarkKey,
		acfunKeyset.ConcernStateKey,
		acfunKeyset.ConcernConfigKey,

		weiboKeyset.AtAllMarkKey,
		weiboKeyset.ConcernStateKey,
		weiboKeyset.ConcernConfigKey,

		tcKeyset.AtAllMarkKey,
		tcKeyset.ConcernStateKey,
		tcKeyset.ConcernConfigKey,

		localdb.TargetPermissionKey,
		localdb.TargetEnabledKey,
		localdb.TargetSilenceKey,
		localdb.TargetMuteKey,
	} {
		f = append(f, version.MigrationKeyValueByPattern(pattern, v.genPrefixGroupPatternMigrate(pattern)))
	}
	f = append(f, version.MigrationKeyValueByRaw(func(key, value string) (string, string) {
		if strings.HasSuffix(key, "GroupAdmin") {
			return strings.TrimSuffix(key, "GroupAdmin") + "TargetAdmin", value
		}
		return key, value
	}))

	f = append(f, version.MigrationKeyValueByPattern(localdb.PermissionKey, func(key, value string) (string, string) {
		spts := strings.Split(key, ":")
		if len(spts) != 4 {
			return key, value
		}
		if strings.HasSuffix(key, "Admin") {
			return key, value
		}
		var g []interface{}
		for _, s := range spts {
			g = append(g, s)
		}
		return localdb.TargetPermissionKey(g...), value
	}))

	return version.ChainMigration(f...)
}

func (v *V2) TargetVersion() int64 {
	return 2
}
