package migration

import (
	"github.com/Sora233/DDBOT/internal/test"
	localdb "github.com/Sora233/DDBOT/lsp/buntdb"
	"github.com/Sora233/DDBOT/lsp/version"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMigrationV2(t *testing.T) {
	logrus.SetLevel(logrus.WarnLevel)
	test.InitBuntdb(t)
	defer test.CloseBuntdb(t)

	var err error

	_, err = version.SetVersion(testName, 1)
	assert.Nil(t, err)

	var kv = [][2]string{
		{"DouyuConcernState:123456:222", "live"},
		{"weiboGroupConcernState:223456:333", "news"},
		{"AcfunGroupConcernState:323456:3444", "live"},
		{"ConcernState:423456:444", "live"},
		{"ConcernConfig:523456:555", `{"group_concern_at":{"at_all":"","at_someone":null},"group_concern_notify":{"title_change_notify":"","offline_notify":""},"group_concern_filter":{"type":"not_type","config":"{\"type\":[\"转发\"]}"}}`},
		{"Permission:623456:6666:watch", ""},
		{"Permission:723456:admin", ""},
		{"GroupAtAll:823456:222", ""},
		{"GroupPermission:923456:333:GroupAdmin", ""},
		{"GroupEnable:125212:xx", "enable"},
		{"GroupSilence:123456", ""},
		{"GroupMute:123456:0", ""},
		{"GroupMute:123456:123", ""},
		{"NotifyMsg:123456:659594770670157856", ""},
	}
	var expected = [][2]string{
		{"DouyuConcernState:Group_123456:222", "live"},
		{"weiboConcernState:Group_223456:333", "news"},
		{"AcfunConcernState:Group_323456:3444", "live"},
		{"ConcernState:Group_423456:444", "live"},
		{"ConcernConfig:Group_523456:555", `{"concern_at":{"at_all":"","at_someone":null},"concern_notify":{"title_change_notify":"","offline_notify":""},"concern_filter":{"type":"not_type","config":"{\"type\":[\"转发\"]}"}}`},
		{"TargetPermission:Group_623456:6666:watch", ""},
		{"Permission:723456:admin", ""},
		{"AtAllMark:Group_823456:222", ""},
		{"TargetPermission:Group_923456:333:TargetAdmin", ""},
		{"TargetEnable:Group_125212:xx", "enable"},
		{"TargetSilence:Group_123456", ""},
		{"TargetMute:Group_123456:0", ""},
		{"TargetMute:Group_123456:123", ""},
		{"NotifyMsg:Group_123456:659594770670157856", ""},
	}
	err = localdb.RWCover(func() error {
		var err error
		for _, p := range kv {
			err = localdb.Set(p[0], p[1])
			if err != nil {
				return err
			}
		}
		return err
	})
	assert.Nil(t, err)

	err = version.DoMigration(testName, version.NewMigrationMapFromMap(map[int64]version.Migration{1: new(V2)}))
	assert.Nil(t, err)

	assert.Equal(t, len(kv), len(expected))

	//localdb.RCoverTx(func(tx *buntdb.Tx) error {
	//	return tx.Ascend("", func(key, value string) bool {
	//		fmt.Printf("%v %v\n", key, value)
	//		return true
	//	})
	//})

	for idx := range kv {
		if kv[idx][0] != expected[idx][0] {
			assert.False(t, localdb.Exist(kv[idx][0]), idx)
			assert.Truef(t, localdb.Exist(expected[idx][0]), "%v - %v", idx, expected[idx][0])
		} else {
			assert.Truef(t, localdb.Exist(expected[idx][0]), "%v - %v", idx, expected[idx][0])
		}
		v, err := localdb.Get(expected[idx][0])
		assert.Nil(t, err, idx)
		assert.EqualValuesf(t, expected[idx][1], v, "%v - %v", idx, expected[idx][1])
	}
}
