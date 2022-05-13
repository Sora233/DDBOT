package version

import (
	"github.com/Sora233/DDBOT/internal/test"
	localdb "github.com/Sora233/DDBOT/lsp/buntdb"
	"github.com/Sora233/DDBOT/lsp/concern"
	"github.com/Sora233/DDBOT/lsp/mmsg/mt"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/buntdb"
	"testing"
)

const testName = "test-mig"

var (
	g1 = mt.NewGroupTarget(test.G1)
	g2 = mt.NewGroupTarget(test.G2)
)

func v1() MigrationFunc {
	return ChainMigration(
		func() error {
			return localdb.RWCoverTx(func(tx *buntdb.Tx) error {
				_, _, err := tx.Set(localdb.BilibiliConcernStateKey(g1, test.UID1), "3", nil)
				if err != nil {
					return err
				}
				_, _, err = tx.Set(localdb.BilibiliConcernStateKey(g1, test.UID2), "1", nil)
				return err
			})
		},
		func() error {
			return nil
		},
	)
}

func v2() MigrationFunc {
	return MigrationValueByPattern(localdb.BilibiliConcernStateKey, func(key, value string) string {
		if value == "1" {
			return "live"
		}
		if value == "3" {
			return "live/news"
		}
		return value
	})
}

func v3() MigrationFunc {
	return ChainMigration(
		MigrationKeyValueByPattern(localdb.BilibiliConcernStateKey, func(key, value string) (string, string) {
			groupCode, id, err := concern.ParseConcernStateKeyWithString(key)
			if err != nil {
				return key, value
			}
			return localdb.DouyuConcernStateKey(groupCode, id), value
		}),
		CopyKeyValueByPattern(localdb.DouyuConcernStateKey, func(key, value string) (string, string) {
			groupCode, id, err := concern.ParseConcernStateKeyWithString(key)
			if err != nil {
				return key, value
			}
			return localdb.HuyaConcernStateKey(groupCode, id), value
		}),
	)
}

func TestDoMigration(t *testing.T) {
	test.InitBuntdb(t)
	defer test.CloseBuntdb(t)

	assert.NotNil(t, DoMigration(testName, nil))
	assert.Zero(t, GetCurrentVersion(testName))

	var migMap = map[int64]Migration{
		0: CreateSimpleMigration(1, v1()),
		1: CreateSimpleMigration(99, v2()),
	}
	m := NewMigrationMapFromMap(migMap)
	assert.Nil(t, DoMigration(testName, m))

	assert.EqualValues(t, 99, GetCurrentVersion(testName))

	err := localdb.RCoverTx(func(tx *buntdb.Tx) error {
		val, err := tx.Get(localdb.BilibiliConcernStateKey(g1, test.UID1))
		if err != nil {
			return err
		}
		assert.EqualValues(t, "live/news", val)

		val, err = tx.Get(localdb.BilibiliConcernStateKey(g1, test.UID2))
		if err != nil {
			return err
		}
		assert.EqualValues(t, "live", val)
		return nil
	})
	assert.Nil(t, err)

	migMap[99] = CreateSimpleMigration(100, v3())

	m = NewMigrationMapFromMap(migMap)
	assert.Nil(t, DoMigration(testName, m))

	assert.EqualValues(t, 100, GetCurrentVersion(testName))

	err = localdb.RCoverTx(func(tx *buntdb.Tx) error {
		assert.False(t, localdb.Exist(localdb.BilibiliConcernStateKey(g1, test.UID1)))
		assert.False(t, localdb.Exist(localdb.BilibiliConcernStateKey(g1, test.UID2)))

		val, err := tx.Get(localdb.DouyuConcernStateKey(g1, test.UID1))
		if err != nil {
			return err
		}
		assert.EqualValues(t, "live/news", val)

		val, err = tx.Get(localdb.DouyuConcernStateKey(g1, test.UID2))
		if err != nil {
			return err
		}
		assert.EqualValues(t, "live", val)

		val, err = tx.Get(localdb.HuyaConcernStateKey(g1, test.UID1))
		if err != nil {
			return err
		}
		assert.EqualValues(t, "live/news", val)

		val, err = tx.Get(localdb.HuyaConcernStateKey(g1, test.UID2))
		if err != nil {
			return err
		}
		assert.EqualValues(t, "live", val)

		return nil

	})
}
