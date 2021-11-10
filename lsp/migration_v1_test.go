package lsp

import (
	"github.com/Sora233/DDBOT/internal/test"
	localdb "github.com/Sora233/DDBOT/lsp/buntdb"
	"github.com/Sora233/DDBOT/lsp/version"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMigrationV1(t *testing.T) {
	test.InitBuntdb(t)
	defer test.CloseBuntdb(t)

	var cfg = groupConcernConfig{
		GroupConcernAt: groupConcernAtConfig{
			AtAll: bibiliLive,
			AtSomeone: []*atSomeone{
				{
					Ctype:  douyuLive,
					AtList: []int64{1, 2, 3, 4, 5},
				},
			},
		},
		GroupConcernNotify: groupConcernNotifyConfig{
			TitleChangeNotify: bilibiliNews | bibiliLive | douyuLive | huyaLive | youtubeLive | youtubeVideo,
			OfflineNotify:     bilibiliNews | bibiliLive | douyuLive | huyaLive | youtubeLive | youtubeVideo,
		},
		GroupConcernFilter: groupConcernFilterConfig{},
	}
	s, err := json.MarshalToString(&cfg)
	assert.Nil(t, err)

	err = localdb.Set(localdb.BilibiliGroupConcernConfigKey(test.G1, test.UID1), s)
	assert.Nil(t, err)

	err = localdb.Set(localdb.BilibiliGroupConcernConfigKey(test.G1, test.UID2), "wrong key")
	assert.Nil(t, err)

	err = localdb.SetInt64(localdb.BilibiliGroupConcernStateKey(test.G1, test.UID1), int64(bilibiliNews|bibiliLive))
	assert.Nil(t, err)

	err = localdb.Set(localdb.BilibiliGroupConcernStateKey(test.G1, test.UID2), "wrong key")
	assert.Nil(t, err)

	err = version.DoMigration(LspVersionName, lspMigrationMap)
	assert.Nil(t, err)

}
