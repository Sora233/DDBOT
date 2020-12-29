package buntdb

import (
	"errors"
	"github.com/tidwall/buntdb"
	"strconv"
	"strings"
)

var db *buntdb.DB

func InitBuntDB() error {
	buntDB, err := buntdb.Open(".lsp.db")
	if err != nil {
		return err
	}
	db = buntDB
	return nil
}

func GetClient() (*buntdb.DB, error) {
	if db == nil {
		return nil, errors.New("not initialized")
	}
	return db, nil
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

func Key(keys ...interface{}) string {
	var _keys []string
	for _, key := range keys {
		switch key.(type) {
		case string:
			_keys = append(_keys, key.(string))
		case int, int64, int32:
			_keys = append(_keys, strconv.FormatInt(key.(int64), 10))
		}
	}
	return strings.Join(_keys, ":")
}
