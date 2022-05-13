package concern

import (
	"github.com/Sora233/DDBOT/internal/test"
	localdb "github.com/Sora233/DDBOT/lsp/buntdb"
	"github.com/Sora233/DDBOT/lsp/mmsg/mt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParseConcernStateKeyWithInt64(t *testing.T) {
	var testCase = []string{
		localdb.BilibiliConcernStateKey(mt.NewGroupTarget(test.G1), test.UID1),
		localdb.DouyuConcernStateKey(mt.NewGroupTarget(test.G1), test.UID1),
	}
	var expected1 = []mt.Target{
		mt.NewGroupTarget(test.G1),
		mt.NewGroupTarget(test.G1),
	}
	var expected2 = []int64{
		test.UID1,
		test.UID1,
	}
	assert.Equal(t, len(expected2), len(expected1))
	assert.Equal(t, len(expected2), len(testCase))
	for index := range testCase {
		a, b, err := ParseConcernStateKeyWithInt64(testCase[index])
		assert.Nil(t, err)
		assert.True(t, a.Equal(expected1[index]))
		assert.Equal(t, expected2[index], b)
	}
}

func TestParseConcernStateKeyWithInt642(t *testing.T) {
	var testCase = []string{
		"wrong_key",
		localdb.BilibiliConcernStateKey("wrong_group", test.UID1),
		localdb.YoutubeConcernStateKey(mt.NewGroupTarget(test.G1), test.NAME1),
	}

	for _, key := range testCase {
		_, _, err := ParseConcernStateKeyWithInt64(key)
		assert.NotNil(t, err)
	}

}

func TestParseConcernStateKeyWithString(t *testing.T) {
	var testCase = []string{
		localdb.YoutubeConcernStateKey(mt.NewGroupTarget(test.G1), test.NAME1),
		localdb.HuyaConcernStateKey(mt.NewGroupTarget(test.G1), test.NAME1),
	}
	var expected1 = []mt.Target{
		mt.NewGroupTarget(test.G1),
		mt.NewGroupTarget(test.G1),
	}
	var expected2 = []string{
		test.NAME1,
		test.NAME1,
	}
	assert.Equal(t, len(expected2), len(expected1))
	assert.Equal(t, len(expected1), len(testCase))
	for index := range testCase {
		a, b, err := ParseConcernStateKeyWithString(testCase[index])
		assert.Nil(t, err)
		assert.True(t, a.Equal(expected1[index]))
		assert.Equal(t, expected2[index], b)
	}
}

func TestParseConcernStateKeyWithString2(t *testing.T) {
	var testCase = []string{
		"wrong_key",
		localdb.YoutubeConcernStateKey("wrong_group", test.NAME1),
	}
	for _, key := range testCase {
		_, _, err := ParseConcernStateKeyWithString(key)
		assert.NotNil(t, err)
	}
}

func TestNewPrefixKeySetWithInt64ID(t *testing.T) {
	pks := NewPrefixKeySetWithInt64ID("test1")
	assert.NotNil(t, pks)
	pks.FreshKey()
	pks.AtAllMarkKey()
	pks.ConcernConfigKey()
	g, id, err := pks.ParseConcernStateKey(pks.ConcernStateKey(mt.NewGroupTarget(test.G1), test.UID1))
	assert.Nil(t, err)
	assert.True(t, g.Equal(mt.NewGroupTarget(test.G1)))
	assert.EqualValues(t, test.UID1, id)
}

func TestNewPrefixKeySetWithStringID(t *testing.T) {
	pks := NewPrefixKeySetWithStringID("test2")
	assert.NotNil(t, pks)
	pks.FreshKey()
	pks.AtAllMarkKey()
	pks.ConcernConfigKey()
	g, id, err := pks.ParseConcernStateKey(pks.ConcernStateKey(mt.NewGroupTarget(test.G1), test.NAME1))
	assert.Nil(t, err)
	assert.True(t, g.Equal(mt.NewGroupTarget(test.G1)))
	assert.EqualValues(t, test.NAME1, id)
}
