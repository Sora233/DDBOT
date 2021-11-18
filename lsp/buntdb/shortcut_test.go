package buntdb

import (
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/buntdb"
	"reflect"
	"testing"
)

type test1 struct {
	A1 string `json:"a_1"`
	B1 string `json:"b_1"`
}

type test2 struct {
	A2 string `json:"a_2"`
	B2 string `json:"b_2"`
}

const (
	key1 = "test1"
	key2 = "test2"
)

func TestShortCut_JsonGet(t *testing.T) {
	var err error
	err = InitBuntDB(MEMORYDB)
	assert.Nil(t, err)
	defer Close()

	var expected = []interface{}{
		&test1{"A1A1", "B1B1"},
		&test1{"A1A1A1", "B1B1B1"},
		&test2{"A2A2", "B2B2"},
		&test2{"A2A2A2", "B2B2B2"},
	}

	var keys = []string{
		key1, key1, key2, key2,
	}

	assert.Equal(t, len(expected), len(keys))

	for i := 0; i < len(expected); i++ {
		err = RWCover(func() error {
			return SetJson(keys[i], expected[i])
		})
		assert.Nil(t, err)
		var a = reflect.New(reflect.TypeOf(expected[i]).Elem()).Interface()
		err = RCover(func() error {
			return GetJson(keys[i], a)
		})
		assert.Nil(t, err)
		assert.EqualValues(t, expected[i], a)
	}

	var notfound = &test1{}

	assert.EqualValues(t, buntdb.ErrNotFound, GetJson("not_found", notfound))
	assert.Nil(t, GetJson("not_found", notfound, IgnoreNotFoundOpt()))

	assert.NotNil(t, GetJson("nil", nil))
	assert.NotNil(t, SetJson("nil", nil))

	_, err = Delete(key1, IgnoreNotFoundOpt())
	assert.Nil(t, err)

	var lastS *test1
	var s1 = &test1{"A1", "B1"}
	var s2 = &test1{"A2", "B2"}
	assert.Nil(t, SetJson(key1, s1, SetGetPreviousValueJsonObjectOpt(nil)))
	assert.Nil(t, SetJson(key1, s2, SetGetPreviousValueJsonObjectOpt(&lastS)))
	assert.EqualValues(t, s1, lastS)
}
