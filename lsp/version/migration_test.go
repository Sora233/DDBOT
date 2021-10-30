package version

import (
	"github.com/Sora233/DDBOT/internal/test"
	localdb "github.com/Sora233/DDBOT/lsp/buntdb"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/buntdb"
	"testing"
)

const testName = "test-mig"

type testMigV1 struct {
}

func (t *testMigV1) Func() MigrationFunc {
	return ChainMigration(
		func() error {
			return localdb.RWCoverTx(func(tx *buntdb.Tx) error {
				_, _, err := tx.Set(localdb.BilibiliGroupConcernStateKey(test.G1, test.UID1), "3", nil)
				if err != nil {
					return err
				}
				_, _, err = tx.Set(localdb.BilibiliGroupConcernStateKey(test.G1, test.UID2), "1", nil)
				return err
			})
		},
		func() error {
			return nil
		},
	)
}

func (t *testMigV1) TargetVersion() int64 {
	return 1
}

type testMigV2 struct {
}

func (t *testMigV2) Func() MigrationFunc {
	return MigrationByKey(localdb.BilibiliGroupConcernStateKey, func(key, value string) string {
		if value == "1" {
			return "live"
		}
		if value == "3" {
			return "live/news"
		}
		return value
	})
}

func (t *testMigV2) TargetVersion() int64 {
	return 99
}

func TestDoMigration(t *testing.T) {
	test.InitBuntdb(t)
	defer test.CloseBuntdb(t)

	assert.NotNil(t, DoMigration(testName, nil))

	assert.Zero(t, GetCurrentVersion(testName))

	var migMap = map[int64]Migration{
		0: new(testMigV1),
		1: new(testMigV2),
	}
	m := NewMigrationMapFromMap(migMap)
	assert.Nil(t, DoMigration(testName, m))

	assert.EqualValues(t, 99, GetCurrentVersion(testName))

	err := localdb.RCoverTx(func(tx *buntdb.Tx) error {
		val, err := tx.Get(localdb.BilibiliGroupConcernStateKey(test.G1, test.UID1))
		if err != nil {
			return err
		}
		assert.EqualValues(t, "live/news", val)

		val, err = tx.Get(localdb.BilibiliGroupConcernStateKey(test.G1, test.UID2))
		if err != nil {
			return err
		}
		assert.EqualValues(t, "live", val)
		return nil
	})
	assert.Nil(t, err)
}
