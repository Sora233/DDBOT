package version

import (
	localdb "github.com/Sora233/DDBOT/lsp/buntdb"
	"github.com/tidwall/buntdb"
)

// ChainMigration 将多个 MigrationFunc 组合成一个 MigrationFunc ，每个 MigrationFunc 会按顺序执行
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

// MigrationValueByPattern 遍历一个KeyPattern，对每个key/value执行operator，并把operator的返回值作为新的value，保留TTL
func MigrationValueByPattern(patternFunc localdb.KeyPatternFunc,
	operator func(key, value string) string) MigrationFunc {
	return migrationByPattern(patternFunc, false, func(key, value string) (string, string) {
		return key, operator(key, value)
	})
}

// MigrationKeyValueByPattern 遍历一个KeyPattern，对每个key/value执行operator，并把operator的返回值作为新的key/value，保留TTL
// 如果key变化了，会删除老的key
func MigrationKeyValueByPattern(patternFunc localdb.KeyPatternFunc,
	operator func(key, value string) (string, string)) MigrationFunc {
	return migrationByPattern(patternFunc, false, operator)
}

// CopyKeyValueByPattern 遍历一个KeyPattern，对每个key/value执行operator，并把operator的返回值作为新的key/value，保留TTL
// 如果key变化了，不会删除老的key
func CopyKeyValueByPattern(patternFunc localdb.KeyPatternFunc,
	operator func(key, value string) (string, string)) MigrationFunc {
	return migrationByPattern(patternFunc, true, operator)
}

func MigrationKeyValueByRaw(operator func(key, value string) (string, string)) MigrationFunc {
	return migrationByPattern(nil, false, operator)
}

func MigrationValueByRaw(operator func(key, value string) string) MigrationFunc {
	return migrationByPattern(nil, false, func(key, value string) (string, string) {
		return key, operator(key, value)
	})
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

func migrationByPattern(patternFunc localdb.KeyPatternFunc, keepOldOnChanged bool,
	operator func(key, value string) (string, string)) MigrationFunc {
	return func() error {
		if patternFunc != nil {
			if err := localdb.CreatePatternIndex(patternFunc, nil); err != nil {
				return err
			}
		}
		return localdb.RWCoverTx(func(tx *buntdb.Tx) error {
			var data [][2]string
			var err error
			if patternFunc == nil {
				err = tx.Ascend("", func(key, value string) bool {
					data = append(data, [2]string{key, value})
					return true
				})
			} else {
				err = tx.Ascend(patternFunc(), func(key, value string) bool {
					data = append(data, [2]string{key, value})
					return true
				})
			}
			if err != nil {
				return err
			}
			var key, value string
			var oldKey, oldValue string
			for _, kv := range data {
				oldKey, oldValue = kv[0], kv[1]
				key, value = operator(oldKey, oldValue)
				expire, err := tx.TTL(oldKey)
				if err != nil && localdb.IsNotFound(err) {
					// 处理的时候正好过期了?
					continue
				}
				if err != nil {
					return err
				}
				if expire > 0 {
					_, _, err = tx.Set(key, value, localdb.ExpireOption(expire))
				} else {
					_, _, err = tx.Set(key, value, nil)
				}
				if err != nil {
					return err
				}
				if !keepOldOnChanged && oldKey != key {
					_, err = tx.Delete(oldKey)
					if err != nil && !localdb.IsNotFound(err) {
						return err
					}
				}
			}
			return nil
		})
	}
}
