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

// InitBuntDB 初始化buntdb，正常情况下框架会负责初始化
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

// GetClient 获取 buntdb.DB 对象，如果没有初始化会返回 ErrNotInitialized
func GetClient() (*buntdb.DB, error) {
	if db == nil {
		return nil, ErrNotInitialized
	}
	return db, nil
}

// MustGetClient 获取 buntdb.DB 对象，如果没有初始化会panic，在编写订阅组件时可以放心调用
func MustGetClient() *buntdb.DB {
	if db == nil {
		panic(ErrNotInitialized)
	}
	return db
}

// Close 关闭buntdb，正常情况下框架会负责关闭
func Close() error {
	if db != nil {
		if itx := gls.Get(txKey); itx != nil {
			itx.(*buntdb.Tx).Rollback()
		}
		if err := db.Close(); err != nil {
			return err
		}
		db = nil
	}
	return nil
}
