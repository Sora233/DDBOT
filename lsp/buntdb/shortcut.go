package buntdb

import (
	"github.com/tidwall/buntdb"
	"time"
)

type ShortCut struct{}

var shortCut = new(ShortCut)

// RWTxCover 在一个RW事务中执行f，注意f的返回值不一定是RWTxCover的返回值
// 有可能f返回nil，但RWTxCover返回non-nil
// 可以忽略error，但不要简单地用f返回值替代RWTxCover返回值，ref: bilibili/MarkDynamicId
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
	if duration == 0 {
		return nil
	}
	return &buntdb.SetOptions{
		Expires: true,
		TTL:     duration,
	}
}
