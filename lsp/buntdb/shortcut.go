package buntdb

import "github.com/tidwall/buntdb"

type ShortCut struct{}

func (*ShortCut) RWTxCover(f func(tx *buntdb.Tx) error) error {
	db, err := GetClient()
	if err != nil {
		return err
	}
	return db.Update(func(tx *buntdb.Tx) error {
		return f(tx)
	})
}

func (*ShortCut) RTxCover(f func(tx *buntdb.Tx) error) error {
	db, err := GetClient()
	if err != nil {
		return err
	}
	return db.View(func(tx *buntdb.Tx) error {
		return f(tx)
	})
}
