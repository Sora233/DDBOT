package concern_manager

import (
	"github.com/Sora233/DDBOT/concern"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGroupConcernAtConfig_CheckAtAll(t *testing.T) {
	var g = &GroupConcernAtConfig{
		AtAll: concern.BilibiliNews,
	}
	assert.True(t, g.CheckAtAll(concern.BilibiliNews))
	assert.False(t, g.CheckAtAll(concern.BibiliLive))
}

func TestNewGroupConcernConfigFromString(t *testing.T) {
	var testCase = []string{
		`{"group_concern_at":{"at_all":1,"at_someone":[{"ctype":1, "at_list":[1,2,3,4,5]}]}}`,
	}
	var expected = []*GroupConcernConfig{
		{
			GroupConcernAt: GroupConcernAtConfig{
				AtAll: concern.BibiliLive,
				AtSomeone: []*AtSomeone{
					{
						Ctype:  concern.BibiliLive,
						AtList: []int64{1, 2, 3, 4, 5},
					},
				},
			},
		},
	}
	assert.Equal(t, len(testCase), len(expected))
	for i := 0; i < len(testCase); i++ {
		g, err := NewGroupConcernConfigFromString(testCase[i])
		assert.Nil(t, err)
		assert.EqualValues(t, expected[i], g)
	}
}

func TestGroupConcernConfig_ToString(t *testing.T) {
	var testCase = []*GroupConcernConfig{
		{
			GroupConcernAt: GroupConcernAtConfig{
				AtAll: concern.BibiliLive,
				AtSomeone: []*AtSomeone{
					{
						Ctype:  concern.BibiliLive,
						AtList: []int64{1, 2, 3, 4, 5},
					},
				},
			},
		},
	}
	var expected = []string{
		`{"group_concern_at":{"at_all":1,"at_someone":[{"ctype":1, "at_list":[1,2,3,4,5]}]}}`,
	}
	assert.Equal(t, len(testCase), len(expected))
	for i := 0; i < len(testCase); i++ {
		assert.JSONEq(t, expected[i], testCase[i].ToString())
	}
}

func TestGroupConcernAtConfig_GetAtSomeoneList(t *testing.T) {
	var testCase = []*GroupConcernConfig{
		{
			GroupConcernAt: GroupConcernAtConfig{
				AtAll: concern.BibiliLive,
				AtSomeone: []*AtSomeone{
					{
						Ctype:  concern.BibiliLive,
						AtList: []int64{1, 2, 3, 4, 5},
					},
				},
			},
		},
	}
	var expected = [][]int64{
		{1, 2, 3, 4, 5},
	}
	assert.Equal(t, len(testCase), len(expected))
	for i := 0; i < len(testCase); i++ {
		assert.EqualValues(t, expected[i], testCase[i].GroupConcernAt.GetAtSomeoneList(concern.BibiliLive))
	}
}
