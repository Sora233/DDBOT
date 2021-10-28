package buntdb

import (
	jsoniter "github.com/json-iterator/go"
	"github.com/modern-go/gls"
	"github.com/tidwall/buntdb"
)

var db *buntdb.DB

const MEMORYDB = ":memory:"
const LSPDB = ".lsp.db"

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func InitBuntDB(dbpath string) error {
	if dbpath == "" {
		dbpath = LSPDB
	}
	buntDB, err := buntdb.Open(dbpath)
	if err != nil {
		return err
	}
	if dbpath != MEMORYDB {
		buntDB.SetConfig(buntdb.Config{
			SyncPolicy:           buntdb.EverySecond,
			AutoShrinkPercentage: 10,
			AutoShrinkMinSize:    1 * 1024 * 1024,
		})
	}
	db = buntDB
	return nil
}

func GetClient() (*buntdb.DB, error) {
	if db == nil {
		return nil, ErrNotInitialized
	}
	return db, nil
}

func MustGetClient() *buntdb.DB {
	if db == nil {
		panic(ErrNotInitialized)
	}
	return db
}

func Close() error {
	if db != nil {
		if itx := gls.Get(TxKey); itx != nil {
			itx.(*buntdb.Tx).Rollback()
		}
		if err := db.Close(); err != nil {
			return err
		}
		db = nil
	}
	return nil
}
