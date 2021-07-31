package buntdb

import (
	"github.com/tidwall/buntdb"
	"time"
)

type ShortCut struct{}

var shortCut = new(ShortCut)

// RWCoverTx 在一个RW事务中执行f，注意f的返回值不一定是RWCoverTx的返回值
// 有可能f返回nil，但RWTxCover返回non-nil
// 可以忽略error，但不要简单地用f返回值替代RWTxCover返回值，ref: bilibili/MarkDynamicId
func (s *ShortCut) RWCoverTx(f func(tx *buntdb.Tx) error) error {
	return withinTransactionNested(true, f)
}

func (s *ShortCut) RCoverTx(f func(tx *buntdb.Tx) error) error {
	return withinTransactionNested(false, f)
}

func (s *ShortCut) RCover(f func() error) error {
	return withinTransactionNested(false, func(tx *buntdb.Tx) error {
		return f()
	})
}

func RWTxCover(f func(tx *buntdb.Tx) error) error {
	return shortCut.RWCoverTx(f)
}

func RTxCover(f func(tx *buntdb.Tx) error) error {
	return shortCut.RCoverTx(f)
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
