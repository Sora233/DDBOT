package buntdb

import (
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/buntdb"
	"os"
	"testing"
	"time"
)

func TestIsRollback(t *testing.T) {
	assert.True(t, IsRollback(ErrRollback))
	assert.False(t, IsRollback(os.ErrNotExist))
}

func TestExpireOption(t *testing.T) {
	e := ExpireOption(time.Hour * 60)
	assert.NotNil(t, e)
	assert.EqualValues(t, time.Hour*60, e.TTL)
	assert.True(t, e.Expires)
	assert.Nil(t, ExpireOption(0))
}

func TestGetClient(t *testing.T) {
	_, err := GetClient()
	assert.EqualValues(t, ErrNotInitialized, err)
	assert.Nil(t, InitBuntDB(MEMORYDB))
	db, err := GetClient()
	assert.Nil(t, err)
	assert.NotNil(t, db)
	assert.NotNil(t, MustGetClient())
	assert.Nil(t, Close())
}

func TestGetClient2(t *testing.T) {
	defer func() {
		e := recover()
		assert.NotNil(t, e)
		assert.Equal(t, ErrNotInitialized, e)
	}()
	MustGetClient()
}

func TestNamedKey(t *testing.T) {
	var testName = []string{
		"t1", "t2",
	}
	var testKey = [][]interface{}{
		{
			"s1", "s2", int32(3), int64(4),
		},
		{
			"s3", uint32(5), false,
		},
	}
	var expected = []string{
		"t1:s1:s2:3:4",
		"t2:s3:5:false",
	}

	assert.Equal(t, len(expected), len(testName))
	assert.Equal(t, len(expected), len(testKey))
	for i := 0; i < len(expected); i++ {
		assert.Equal(t, expected[i], NamedKey(testName[i], testKey[i]))
	}
}

func TestRTxCover(t *testing.T) {
	err := RWTxCover(func(tx *buntdb.Tx) error {
		return nil
	})
	assert.Equal(t, ErrNotInitialized, err)
	err = RTxCover(func(tx *buntdb.Tx) error {
		return nil
	})
	assert.Equal(t, ErrNotInitialized, err)

	err = InitBuntDB(MEMORYDB)
	assert.Nil(t, err)
	defer Close()
	err = RTxCover(func(tx *buntdb.Tx) error {
		_, _, err := tx.Set("a", "b", nil)
		return err
	})
	assert.Equal(t, buntdb.ErrTxNotWritable, err)
	err = RWTxCover(func(tx *buntdb.Tx) error {
		_, _, err := tx.Set("a", "b", nil)
		return err
	})
	assert.Nil(t, err)
	_ = RTxCover(func(tx *buntdb.Tx) error {
		val, err := tx.Get("a")
		assert.Equal(t, "b", val)
		assert.Nil(t, err)
		return nil
	})
}

func TestRWTxCover(t *testing.T) {
	var err error
	err = InitBuntDB(MEMORYDB)
	assert.Nil(t, err)
	defer Close()

	err = RWTxCover(func(tx *buntdb.Tx) error {
		_, _, err := tx.Set("a", "b", ExpireOption(time.Hour*48))
		return err
	})
	assert.Nil(t, err)
	err = RWTxCover(func(tx *buntdb.Tx) error {
		tx.Set("a", "c", ExpireOption(time.Second*1))
		return ErrRollback
	})
	assert.EqualValues(t, ErrRollback, err)
	var ttl time.Duration
	err = RTxCover(func(tx *buntdb.Tx) error {
		var err error
		ttl, err = tx.TTL("a")
		return err
	})
	assert.Nil(t, err)
	assert.Greater(t, ttl, time.Hour*47)
}

func TestNestedCover(t *testing.T) {
	var err error
	err = InitBuntDB(MEMORYDB)
	assert.Nil(t, err)
	defer Close()

	setAfn := func() error {
		return RWTxCover(func(tx *buntdb.Tx) error {
			_, _, err := tx.Set("a", "b", nil)
			return err
		})
	}
	setBfn := func() error {
		return RWTxCover(func(tx *buntdb.Tx) error {
			_, _, err := tx.Set("b", "c", nil)
			return err
		})
	}
	setCfn := func() error {
		return RWTxCover(func(tx *buntdb.Tx) error {
			_, _, err := tx.Set("c", "d", nil)
			return err
		})
	}
	readBfn := func() (string, error) {
		var result string
		err := RTxCover(func(tx *buntdb.Tx) error {
			val, err := tx.Get("b", false)
			result = val
			return err
		})
		return result, err
	}

	var val string

	err = RWTxCover(func(tx *buntdb.Tx) error {
		_, _, err := tx.Set("d", "e", nil)
		if err != nil {
			return err
		}
		err = setBfn()
		if err != nil {
			return err
		}
		err = setAfn()
		if err != nil {
			return err
		}
		val, err = readBfn()
		if err != nil {
			return err
		}
		return nil
	})
	assert.Nil(t, err)
	assert.Equal(t, "c", val)

	err = RTxCover(func(tx *buntdb.Tx) error {
		val, err := tx.Get("a")
		assert.Nil(t, err)
		assert.Equal(t, "b", val)
		val, err = tx.Get("b")
		assert.Nil(t, err)
		assert.Equal(t, "c", val)
		val, err = tx.Get("d")
		assert.Nil(t, err)
		assert.Equal(t, "e", val)
		return nil
	})

	err = RTxCover(func(tx *buntdb.Tx) error {
		val, err := readBfn()
		assert.Nil(t, err)
		assert.Equal(t, "c", val)
		err = setCfn()
		assert.EqualValues(t, buntdb.ErrTxNotWritable, err)
		return nil
	})
	assert.Nil(t, err)
	err = RTxCover(func(tx *buntdb.Tx) error {
		_, err := tx.Get("c")
		assert.EqualValues(t, buntdb.ErrNotFound, err)
		return nil
	})
	assert.Nil(t, err)
}

func TestClose(t *testing.T) {
	var err error
	err = InitBuntDB(MEMORYDB)
	assert.Nil(t, err)

	err = RWTxCover(func(tx *buntdb.Tx) error {
		return Close()
	})
	assert.Equal(t, buntdb.ErrTxClosed, err)
}
