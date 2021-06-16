package buntdb

import (
	"github.com/tidwall/buntdb"
)

var db *buntdb.DB

const MEMORYDB = ":memory:"

func InitBuntDB(dbpath string) error {
	if dbpath == "" {
		dbpath = ".lsp.db"
	}
	buntDB, err := buntdb.Open(dbpath)
	if err != nil {
		return err
	}
	buntDB.SetConfig(buntdb.Config{
		SyncPolicy:           buntdb.Always,
		AutoShrinkPercentage: 10,
		AutoShrinkMinSize:    1 * 1024 * 1024,
	})
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
		if err := db.Close(); err != nil {
			return err
		}
		db = nil
	}
	return nil
}
