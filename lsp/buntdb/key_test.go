package buntdb

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

const (
	GroupCode1 int64 = 123456
	Uid        int64 = 777
	Sid              = "uid"
)

func TestParseConcernStateKeyWithInt64(t *testing.T) {
	var testCase = []string{
		BilibiliGroupConcernStateKey(GroupCode1, Uid),
		DouyuGroupConcernStateKey(GroupCode1, Uid),
	}
	var expected = [][]int64{
		{
			GroupCode1, Uid,
		},
		{
			GroupCode1, Uid,
		},
	}
	assert.Equal(t, len(expected), len(testCase))
	for index := range testCase {
		a, b, err := ParseConcernStateKeyWithInt64(testCase[index])
		assert.Nil(t, err)
		assert.EqualValues(t, []int64{a, b}, expected[index])
	}
}

func TestParseConcernStateKeyWithInt642(t *testing.T) {
	var testCase = []string{
		"wrong_key",
		BilibiliGroupConcernStateKey("wrong_group", Uid),
		YoutubeGroupConcernStateKey(GroupCode1, Sid),
	}

	for _, key := range testCase {
		_, _, err := ParseConcernStateKeyWithInt64(key)
		assert.NotNil(t, err)
	}

}

func TestParseConcernStateKeyWithString(t *testing.T) {
	var testCase = []string{
		YoutubeGroupConcernStateKey(GroupCode1, Sid),
		HuyaGroupConcernStateKey(GroupCode1, Sid),
	}
	var expected = [][]interface{}{
		{
			GroupCode1, Sid,
		},
		{
			GroupCode1, Sid,
		},
	}
	assert.Equal(t, len(expected), len(testCase))
	for index := range testCase {
		a, b, err := ParseConcernStateKeyWithString(testCase[index])
		assert.Nil(t, err)
		assert.EqualValues(t, []interface{}{a, b}, expected[index])
	}
}

func TestParseConcernStateKeyWithString2(t *testing.T) {
	var testCase = []string{
		"wrong_key",
		YoutubeGroupConcernStateKey("wrong_group", Sid),
	}
	for _, key := range testCase {
		_, _, err := ParseConcernStateKeyWithString(key)
		assert.NotNil(t, err)
	}
}

func TestModeKey(t *testing.T) {
	assert.Equal(t, "Mode", ModeKey())
}
