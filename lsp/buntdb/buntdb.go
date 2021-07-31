package buntdb

import (
	"github.com/tidwall/buntdb"
	"sync/atomic"
)

var db *buntdb.DB

// tx is a *buntdb.Tx or nilTx
var aTx atomic.Value
var nilTx = new(buntdb.Tx)

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

func withinTransactionNested(writable bool, f func(tx *buntdb.Tx) error) error {
	db, err := GetClient()
	if err != nil {
		return err
	}
	if x := aTx.Load(); x != nil && x != nilTx {
		return f(x.(*buntdb.Tx))
	}
	// db.Begin 有锁
	tx, err := db.Begin(writable)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	aTx.Store(tx)
	// 如果是复用嵌套的事务，只会在执行到f这一行之后才会启动嵌套事务，上面atomic Store之后可以安全启动f
	err = f(tx)
	aTx.Store(nilTx)
	if err == nil && writable {
		err = tx.Commit()
	}
	return err
}

func Close() error {
	if db != nil {
		if x := aTx.Load(); x != nil && x != nilTx {
			x.(*buntdb.Tx).Rollback()
		}
		if err := db.Close(); err != nil {
			return err
		}
		db = nil
	}
	return nil
}
