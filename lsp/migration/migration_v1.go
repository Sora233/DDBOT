package migration

import (
	"encoding/json"
	localdb "github.com/Sora233/DDBOT/lsp/buntdb"
	"github.com/Sora233/DDBOT/lsp/concern"
	"github.com/Sora233/DDBOT/lsp/concern_type"
	"github.com/Sora233/DDBOT/lsp/version"
	"strconv"
	"strings"
)

type oldType int64

const (
	bibiliLive oldType = 1 << iota
	bilibiliNews
	douyuLive
	youtubeLive
	youtubeVideo
	huyaLive
)

func (o oldType) ToNewType() concern_type.Type {
	var nt concern_type.Type
	if o&bibiliLive != 0 || o&douyuLive != 0 || o&youtubeLive != 0 || o&huyaLive != 0 {
		nt = nt.Add("live")
	}
	if o&bilibiliNews != 0 || o&youtubeVideo != 0 {
		nt = nt.Add("news")
	}
	return nt
}

type atSomeone struct {
	Ctype  oldType `json:"ctype"`
	AtList []int64 `json:"at_list"`
}

type groupConcernAtConfig struct {
	AtAll     oldType      `json:"at_all"`
	AtSomeone []*atSomeone `json:"at_someone"`
}

type groupConcernNotifyConfig struct {
	TitleChangeNotify oldType `json:"title_change_notify"`
	OfflineNotify     oldType `json:"offline_notify"`
}

type groupConcernFilterConfig struct {
	Type   string `json:"type"`
	Config string `json:"config"`
}

type groupConcernConfig struct {
	GroupConcernAt     groupConcernAtConfig     `json:"group_concern_at"`
	GroupConcernNotify groupConcernNotifyConfig `json:"group_concern_notify"`
	GroupConcernFilter groupConcernFilterConfig `json:"group_concern_filter"`
}

func newGroupConcernConfigFromString(s string) (*groupConcernConfig, error) {
	var concernConfig *groupConcernConfig
	decoder := json.NewDecoder(strings.NewReader(s))
	decoder.UseNumber()
	err := decoder.Decode(&concernConfig)
	return concernConfig, err
}

type V1 struct {
}

func newV1GroupConcernConfigFromString(s string) (*v1GroupConcernConfig, error) {
	var concernConfig *v1GroupConcernConfig
	decoder := json.NewDecoder(strings.NewReader(s))
	decoder.UseNumber()
	err := decoder.Decode(&concernConfig)
	return concernConfig, err
}

type v1GroupConcernConfig struct {
	GroupConcernAt     concern.ConcernAtConfig     `json:"group_concern_at"`
	GroupConcernNotify concern.ConcernNotifyConfig `json:"group_concern_notify"`
	GroupConcernFilter concern.ConcernFilterConfig `json:"group_concern_filter"`
}

func (g *v1GroupConcernConfig) ToString() string {
	b, e := json.Marshal(g)
	if e != nil {
		panic(e)
	}
	return string(b)
}

func (v *V1) configMigrate(key, value string) string {
	g, err := newGroupConcernConfigFromString(value)
	if err != nil {
		logger.WithField("key", key).Errorf("configMigrate newGroupConcernConfigFromString <%v> error %v", value, err)
		return value
	}
	var ng v1GroupConcernConfig
	ng.GroupConcernAt.AtAll = g.GroupConcernAt.AtAll.ToNewType()
	for _, atone := range g.GroupConcernAt.AtSomeone {
		ng.GroupConcernAt.MergeAtSomeoneList(atone.Ctype.ToNewType(), atone.AtList)
	}

	ng.GroupConcernFilter.Config = g.GroupConcernFilter.Config
	ng.GroupConcernFilter.Type = g.GroupConcernFilter.Type

	ng.GroupConcernNotify.OfflineNotify = g.GroupConcernNotify.OfflineNotify.ToNewType()
	ng.GroupConcernNotify.TitleChangeNotify = g.GroupConcernNotify.TitleChangeNotify.ToNewType()
	return ng.ToString()
}

func (v *V1) concernMigrate(key, value string) string {
	i, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		logger.WithField("key", key).Errorf("concernMigrate parse <%v> error %v", value, err)
		return value
	}
	return oldType(i).ToNewType().String()

}

func (v *V1) Func() version.MigrationFunc {
	return version.ChainMigration(
		version.MigrationValueByPattern(localdb.BilibiliConcernConfigKey, v.configMigrate),
		version.MigrationValueByPattern(localdb.DouyuConcernConfigKey, v.configMigrate),
		version.MigrationValueByPattern(localdb.HuyaConcernConfigKey, v.configMigrate),
		version.MigrationValueByPattern(localdb.YoutubeConcernConfigKey, v.configMigrate),

		version.MigrationValueByPattern(localdb.BilibiliConcernStateKey, v.concernMigrate),
		version.MigrationValueByPattern(localdb.DouyuConcernStateKey, v.concernMigrate),
		version.MigrationValueByPattern(localdb.HuyaConcernStateKey, v.concernMigrate),
		version.MigrationValueByPattern(localdb.YoutubeConcernStateKey, v.concernMigrate),
	)
}
func (v *V1) TargetVersion() int64 {
	return 1
}
