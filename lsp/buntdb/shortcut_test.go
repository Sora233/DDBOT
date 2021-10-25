package buntdb

import (
	"github.com/stretchr/testify/assert"
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
			return JsonSave(keys[i], expected[i], true)
		})
		assert.Nil(t, err)
		var a = reflect.New(reflect.TypeOf(expected[i]).Elem()).Interface()
		err = RCover(func() error {
			return JsonGet(keys[i], a)
		})
		assert.Nil(t, err)
		assert.EqualValues(t, expected[i], a)
	}
}
