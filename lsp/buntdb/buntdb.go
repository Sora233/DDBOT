package buntdb

import (
	"github.com/tidwall/buntdb"
)

var db *buntdb.DB

func InitBuntDB() error {
	buntDB, err := buntdb.Open(".lsp.db")
	if err != nil {
		return err
	}
	buntDB.SetConfig(buntdb.Config{
		SyncPolicy:           buntdb.Always,
		AutoShrinkPercentage: 100,
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
