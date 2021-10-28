package version

import (
	localdb "github.com/Sora233/DDBOT/lsp/buntdb"
	"github.com/tidwall/buntdb"
)

func ChainMigration(fn ...MigrationFunc) MigrationFunc {
	return func() error {
		for _, f := range fn {
			err := f()
			if err != nil {
				return err
			}
		}
		return nil
	}
}

func MigrationByKey(patternFunc localdb.KeyPatternFunc, operator func(key, value string) string) MigrationFunc {
	return func() error {
		err := localdb.CreatePatternIndex(patternFunc, nil)
		if err != nil {
			return err
		}
		return localdb.RWCoverTx(func(tx *buntdb.Tx) error {
			var data [][2]string
			err := tx.Ascend(patternFunc(), func(key, value string) bool {
				data = append(data, [2]string{key, value})
				return true
			})
			if err != nil {
				return err
			}
			var key, value string
			for _, kv := range data {
				key = kv[0]
				value = operator(kv[0], kv[1])
				expire, err := tx.TTL(key)
				if err == nil && expire > 0 {
					_, _, err = tx.Set(key, value, localdb.ExpireOption(expire))
				} else {
					_, _, err = tx.Set(key, value, nil)
				}
				if err != nil {
					return err
				}
			}
			return nil
		})
	}
}

type migrationMap struct {
	m map[int64]Migration
}

func (mm *migrationMap) From(v int64) Migration {
	return mm.m[v]
}

func NewMigrationMapFromMap(m map[int64]Migration) MigrationMap {
	return &migrationMap{m}
}
