package buntdb

import (
	"github.com/tidwall/buntdb"
	"time"
)

type ShortCut struct{}

var shortCut = new(ShortCut)

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

func RWTxCover(f func(tx *buntdb.Tx) error) error {
	return shortCut.RWTxCover(f)
}

func RTxCover(f func(tx *buntdb.Tx) error) error {
	return shortCut.RTxCover(f)
}

func ExpireOption(duration time.Duration) *buntdb.SetOptions {
	return &buntdb.SetOptions{
		Expires: true,
		TTL:     duration,
	}
}
