package buntdb

import (
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/buntdb"
	"os"
	"sync"
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
	err := RWCoverTx(func(tx *buntdb.Tx) error {
		return nil
	})
	assert.Equal(t, ErrNotInitialized, err)
	err = RCoverTx(func(tx *buntdb.Tx) error {
		return nil
	})
	assert.Equal(t, ErrNotInitialized, err)

	err = RCover(func() error {
		return nil
	})
	assert.Equal(t, ErrNotInitialized, err)

	err = RWCover(func() error {
		return nil
	})
	assert.Equal(t, ErrNotInitialized, err)

	err = InitBuntDB(MEMORYDB)
	assert.Nil(t, err)
	defer Close()
	err = RCoverTx(func(tx *buntdb.Tx) error {
		_, _, err := tx.Set("a", "b", nil)
		return err
	})
	assert.Equal(t, buntdb.ErrTxNotWritable, err)
	err = RWCoverTx(func(tx *buntdb.Tx) error {
		_, _, err := tx.Set("a", "b", nil)
		return err
	})
	assert.Nil(t, err)
	_ = RCoverTx(func(tx *buntdb.Tx) error {
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

	err = RWCoverTx(func(tx *buntdb.Tx) error {
		_, _, err := tx.Set("a", "b", ExpireOption(time.Hour*48))
		return err
	})
	assert.Nil(t, err)
	err = RWCoverTx(func(tx *buntdb.Tx) error {
		tx.Set("a", "c", ExpireOption(time.Second*1))
		return ErrRollback
	})
	assert.EqualValues(t, ErrRollback, err)
	var ttl time.Duration
	err = RCoverTx(func(tx *buntdb.Tx) error {
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
		return RWCoverTx(func(tx *buntdb.Tx) error {
			_, _, err := tx.Set("a", "b", nil)
			return err
		})
	}
	setBfn := func() error {
		return RWCoverTx(func(tx *buntdb.Tx) error {
			_, _, err := tx.Set("b", "c", nil)
			return err
		})
	}
	setCfn := func() error {
		return RWCoverTx(func(tx *buntdb.Tx) error {
			_, _, err := tx.Set("c", "d", nil)
			return err
		})
	}
	readBfn := func() (string, error) {
		var result string
		err := RCoverTx(func(tx *buntdb.Tx) error {
			val, err := tx.Get("b", false)
			result = val
			return err
		})
		return result, err
	}

	var val string
	err = RWCover(func() error {
		return RWCover(func() error {
			return RWCoverTx(func(tx *buntdb.Tx) error {
				return RWCoverTx(func(tx *buntdb.Tx) error {
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
			})
		})
	})
	assert.Nil(t, err)
	assert.Equal(t, "c", val)
	err = RCover(func() error {
		return RCover(func() error {
			return RCoverTx(func(tx *buntdb.Tx) error {
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
		})
	})

	err = RCoverTx(func(tx *buntdb.Tx) error {
		val, err := readBfn()
		assert.Nil(t, err)
		assert.Equal(t, "c", val)
		err = setCfn()
		assert.EqualValues(t, buntdb.ErrTxNotWritable, err)
		return nil
	})
	assert.Nil(t, err)
	err = RCoverTx(func(tx *buntdb.Tx) error {
		_, err := tx.Get("c")
		assert.True(t, IsNotFound(err))
		return nil
	})
	assert.Nil(t, err)
}

func TestRWTxCover2(t *testing.T) {
	var err error
	err = InitBuntDB(MEMORYDB)
	assert.Nil(t, err)
	defer Close()

	testFn := func(tx *buntdb.Tx, key, exp string) {
		val, err := tx.Get(key)
		assert.Nil(t, err)
		assert.Equal(t, exp, val)
	}

	set1Fn := func(tx *buntdb.Tx) error {
		_, _, err := tx.Set("a", "a", nil)
		if err != nil {
			return err
		}
		err = RWCoverTx(func(tx *buntdb.Tx) error {
			_, _, err = tx.Set("b", "b", nil)
			return err
		})
		if err != nil {
			return err
		}
		err = RCoverTx(func(tx *buntdb.Tx) error {
			testFn(tx, "a", "a")
			testFn(tx, "b", "b")
			return nil
		})
		return err
	}
	set2Fn := func(tx *buntdb.Tx) error {
		_, _, err := tx.Set("d", "d", nil)
		if err != nil {
			return err
		}
		err = RWCoverTx(func(tx *buntdb.Tx) error {
			_, _, err = tx.Set("c", "c", nil)
			return err
		})
		err = RCoverTx(func(tx *buntdb.Tx) error {
			testFn(tx, "c", "c")
			testFn(tx, "d", "d")
			return nil
		})
		return err
	}

	c := make(chan interface{}, 16)
	const b = 100000
	var wg sync.WaitGroup
	wg.Add(b*2 + 2)
	go func() {
		defer wg.Done()
		for i := 0; i < b; i++ {
			c <- struct{}{}
			go func() {
				defer wg.Done()
				assert.Nil(t, RWCoverTx(set1Fn))
				<-c
			}()
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < b; i++ {
			c <- struct{}{}
			go func() {
				defer wg.Done()
				assert.Nil(t, RWCoverTx(set2Fn))
				<-c
			}()
		}
	}()
	wg.Wait()
}

func TestSeqNext(t *testing.T) {
	var err error
	err = InitBuntDB(MEMORYDB)
	assert.Nil(t, err)
	defer Close()

	seq1 := "seq1"
	seq2 := "seq2"

	for i := 0; i < 100000; i++ {
		s, err := SeqNext(seq1)
		assert.Nil(t, err)
		assert.EqualValuesf(t, i+1, s, "Seq %v times must eq %v", i+1, i+1)
	}

	_, err = Delete(seq1)
	assert.Nil(t, err)

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 100000; i++ {
				_, err := SeqNext(seq1)
				assert.Nil(t, err)
			}
		}()
	}
	wg.Wait()
	s, err := SeqNext(seq1)
	assert.Nil(t, err)
	assert.EqualValues(t, 1000001, s)

	s, err = SeqNext(seq2)
	assert.Nil(t, err)
	assert.EqualValues(t, 1, s)

	_, err = Delete(seq1)
	assert.Nil(t, err)

	_, err = Delete(seq1)
	assert.True(t, IsNotFound(err))
	_, err = Delete(seq1, IgnoreNotFoundOpt())
	assert.Nil(t, err)

	s, err = SeqNext(seq2)
	assert.Nil(t, err)
	assert.EqualValues(t, 2, s)

	s, err = SeqNext(seq1)
	assert.Nil(t, err)
	assert.EqualValues(t, 1, s)

	err = RWCoverTx(func(tx *buntdb.Tx) error {
		_, _, err := tx.Set(seq1, "wrong", nil)
		return err
	})
	assert.Nil(t, err)

	_, err = SeqNext(seq1)
	assert.NotNil(t, err)

	_, err = Delete(seq1)
	assert.Nil(t, err)

	s, err = IncInt64(seq1, 1001)
	assert.Nil(t, err)
	assert.EqualValues(t, 1001, s)

	s, err = IncInt64(seq1, 1001)
	assert.Nil(t, err)
	assert.EqualValues(t, 2002, s)

	s, err = DeleteInt64(seq1)
	assert.Nil(t, err)
	assert.EqualValues(t, 2002, s)

	s, err = GetInt64(seq1, IgnoreNotFoundOpt())
	assert.Nil(t, err)
	assert.EqualValues(t, 0, s)
}

type testJson struct {
	A string   `json:"a"`
	B int      `json:"b"`
	C []string `json:"c"`
}

func TestJsonGet(t *testing.T) {
	var err error
	err = InitBuntDB(MEMORYDB)
	assert.Nil(t, err)
	defer Close()

	var expected = []*testJson{
		{
			A: "a1",
			B: 1,
			C: []string{"c1", "c1"},
		},
		{
			A: "a2",
			B: 2,
			C: []string{"c2", "c2"},
		},
	}
	var keys = []string{
		"key1",
		"key2",
	}
	err = RWCover(func() error {
		for index, exp := range expected {
			var tmp = new(testJson)
			assert.Nil(t, SetJson(keys[index], exp))
			assert.Nil(t, GetJson(keys[index], tmp))
			assert.EqualValues(t, exp, tmp)
		}
		return nil
	})
	assert.Nil(t, err)

	err = RCover(func() error {
		return SetJson(keys[0], expected[0])
	})
	assert.NotNil(t, err)
}

func TestRemoveByPrefixAndIndex(t *testing.T) {
	var err error
	err = InitBuntDB(MEMORYDB)
	assert.Nil(t, err)
	defer Close()

	err = RWCoverTx(func(tx *buntdb.Tx) error {
		_, _, err := tx.Set(BilibiliGroupConcernStateKey("1"), "", nil)
		assert.Nil(t, err)
		_, _, err = tx.Set(BilibiliGroupConcernStateKey("2"), "", nil)
		assert.Nil(t, err)
		_, _, err = tx.Set(BilibiliGroupConcernStateKey("3"), "", nil)
		assert.Nil(t, err)
		_, _, err = tx.Set(DouyuGroupConcernStateKey("4"), "", nil)
		assert.Nil(t, err)
		_, _, err = tx.Set(HuyaGroupConcernStateKey("5"), "", nil)
		assert.Nil(t, err)
		createIndex := func(patternFunc KeyPatternFunc) {
			assert.Nil(t, tx.CreateIndex(patternFunc(), patternFunc("*"), buntdb.IndexString))
		}
		for _, pattern := range []KeyPatternFunc{BilibiliGroupConcernStateKey, DouyuGroupConcernStateKey, HuyaGroupConcernStateKey} {
			createIndex(pattern)
		}
		return nil
	})
	assert.Nil(t, err)
	deletedKeys, err := RemoveByPrefixAndIndex([]string{BilibiliGroupConcernStateKey(), DouyuGroupConcernStateKey()}, []string{BilibiliGroupConcernStateKey(), DouyuGroupConcernStateKey()})
	assert.Nil(t, err)
	assert.Len(t, deletedKeys, 4)
	err = RCoverTx(func(tx *buntdb.Tx) error {
		assertNotExist := func(key string) {
			_, err := tx.Get(key)
			assert.True(t, IsNotFound(err))
		}
		assertNotExist(BilibiliGroupConcernStateKey("1"))
		assertNotExist(BilibiliGroupConcernStateKey("2"))
		assertNotExist(BilibiliGroupConcernStateKey("3"))
		assertNotExist(DouyuGroupConcernStateKey("4"))
		_, err := tx.Get(HuyaGroupConcernStateKey("5"))
		assert.Nil(t, err)
		return nil
	})
	assert.Nil(t, err)
}

func TestSetIfNotExist(t *testing.T) {
	var err error
	err = InitBuntDB(MEMORYDB)
	assert.Nil(t, err)
	defer Close()

	const (
		key1 = "key1"
		key2 = "key2"
	)

	assert.Nil(t, Set(key1, "1", SetNoOverWriteOpt()))
	assert.True(t, IsRollback(Set(key1, "1", SetNoOverWriteOpt())))
	assert.Nil(t, Set(key2, "1", SetExpireOpt(time.Hour)))
}

func TestCreatePatternIndex(t *testing.T) {
	var err error
	err = InitBuntDB(MEMORYDB)
	assert.Nil(t, err)
	defer Close()

	assert.Nil(t, CreatePatternIndex(BilibiliGroupConcernStateKey, nil))
	err = RCoverTx(func(tx *buntdb.Tx) error {
		indexes, err := tx.Indexes()
		assert.Nil(t, err)
		assert.Len(t, indexes, 1)
		assert.Contains(t, indexes, BilibiliGroupConcernStateKey())
		return nil
	})
	assert.Nil(t, err)

	var suffix = []interface{}{"a", "1"}

	assert.Nil(t, CreatePatternIndex(BilibiliGroupConcernStateKey, suffix, buntdb.IndexBinary))
	err = RCoverTx(func(tx *buntdb.Tx) error {
		indexes, err := tx.Indexes()
		assert.Nil(t, err)
		assert.Len(t, indexes, 2)
		assert.Contains(t, indexes, BilibiliGroupConcernStateKey(suffix...))
		return nil
	})
	assert.Nil(t, err)

	assert.EqualValues(t, []interface{}{"a", "1"}, suffix)
}

func TestGetInt64(t *testing.T) {
	var err error
	err = InitBuntDB(MEMORYDB)
	assert.Nil(t, err)
	defer Close()

	const key = "test1"

	result, err := GetInt64(key)
	assert.True(t, IsNotFound(err))

	result, err = GetInt64(key, IgnoreNotFoundOpt())
	assert.Nil(t, err)
	assert.Zero(t, result)

	err = SetInt64(key, 1, SetGetPreviousValueInt64Opt(&result))
	assert.Nil(t, err)
	assert.Zero(t, result)

	err = SetInt64(key, 10, SetExpireOpt(time.Hour), SetGetPreviousValueInt64Opt(&result))
	assert.Nil(t, err)
	assert.EqualValues(t, 1, result)

	result, err = GetInt64(key)
	assert.Nil(t, err)
	assert.EqualValues(t, 10, result)

	time.Sleep(time.Second * 3)

	var lastTTL time.Duration
	assert.Nil(t, SetInt64(key, 12, SetKeepLastExpireOpt()))
	assert.True(t, Exist(key, GetTTLOpt(&lastTTL)))
	assert.True(t, lastTTL < time.Hour)
	assert.True(t, lastTTL > (time.Hour-time.Second*4))
	assert.Nil(t, err)

	assert.Nil(t, SetInt64(key, 13, SetExpireOpt(time.Millisecond*10)))
	time.Sleep(time.Millisecond * 50)

	result, err = GetInt64(key)
	assert.True(t, IsNotFound(err))
	assert.Zero(t, result)

	result, err = GetInt64(key, GetIgnoreExpireOpt())
	assert.Nil(t, err)
	assert.EqualValues(t, 13, result)
}

func TestGet(t *testing.T) {
	var err error
	err = InitBuntDB(MEMORYDB)
	assert.Nil(t, err)
	defer Close()

	const key = "test1"

	result, err := Get(key)
	assert.True(t, IsNotFound(err))

	result, err = Get(key, IgnoreNotFoundOpt())
	assert.Nil(t, err)

	var prev string
	var replaced bool
	assert.Nil(t, Set(key, "1", SetGetPreviousValueStringOpt(&prev), SetGetIsOverwriteOpt(&replaced)))
	assert.Equal(t, "", prev)
	assert.False(t, replaced)
	assert.Nil(t, Set(key, "2", SetGetPreviousValueStringOpt(nil), SetGetIsOverwriteOpt(nil)))
	assert.Nil(t, Set(key, "3", SetGetPreviousValueStringOpt(&prev), SetGetIsOverwriteOpt(&replaced)))
	assert.True(t, replaced)
	assert.EqualValues(t, "2", prev)

	var prevInt64 int64
	assert.Nil(t, SetInt64(key, 2, SetGetPreviousValueInt64Opt(&prevInt64)))
	assert.EqualValues(t, 3, prevInt64)
	assert.Nil(t, SetInt64(key, 5, SetGetPreviousValueInt64Opt(nil)))
	assert.Nil(t, SetInt64(key, 10, SetGetPreviousValueInt64Opt(&prevInt64)))
	assert.EqualValues(t, 5, prevInt64)

	result, err = Get(key)
	assert.Nil(t, err)
	assert.Equal(t, "10", result)

	_, err = Delete(key, IgnoreNotFoundOpt())
	assert.Nil(t, err)

	assert.Nil(t, Set(key, "1", SetExpireOpt(time.Millisecond*50)))

	time.Sleep(time.Millisecond * 60)

	result, err = Get(key, GetIgnoreExpireOpt())
	assert.Nil(t, err)
	assert.EqualValues(t, "1", result)

	var ttl time.Duration
	assert.Nil(t, Set(key, "2", SetExpireOpt(time.Hour)))
	_, err = Get(key, GetTTLOpt(&ttl))
	assert.Nil(t, err)
	assert.NotZero(t, ttl)

	assert.True(t, Exist(key, GetTTLOpt(&ttl)))
	assert.NotZero(t, ttl)

	assert.False(t, Exist("wrong", GetTTLOpt(&ttl)))
}
