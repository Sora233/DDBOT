package concern_manager

import (
	"github.com/Sora233/DDBOT/concern"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGroupConcernAtConfig_CheckAtAll(t *testing.T) {
	var g *GroupConcernAtConfig
	assert.False(t, g.CheckAtAll(concern.BibiliLive))

	g = &GroupConcernAtConfig{
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
			GroupConcernNotify: GroupConcernNotifyConfig{
				TitleChangeNotify: concern.BibiliLive,
				OfflineNotify:     concern.DouyuLive,
			},
		},
	}
	var expected = []string{
		`{"group_concern_at":{"at_all":1,"at_someone":[{"ctype":1, "at_list":[1,2,3,4,5]}]},"group_concern_notify":{"title_change_notify": 1, "offline_notify": 4}}`,
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
		{
			GroupConcernAt: GroupConcernAtConfig{
				AtAll:     concern.BibiliLive,
				AtSomeone: nil,
			},
		},
	}
	var expected = [][]int64{
		{1, 2, 3, 4, 5},
		nil,
	}
	assert.Equal(t, len(testCase), len(expected))
	for i := 0; i < len(testCase); i++ {
		assert.EqualValues(t, expected[i], testCase[i].GroupConcernAt.GetAtSomeoneList(concern.BibiliLive))
	}

	var g *GroupConcernAtConfig
	assert.Nil(t, g.GetAtSomeoneList(concern.BibiliLive))
}

func TestGroupConcernNotifyConfig_CheckTitleChangeNotify(t *testing.T) {
	var g = &GroupConcernNotifyConfig{
		TitleChangeNotify: concern.BibiliLive | concern.DouyuLive,
	}
	assert.True(t, g.CheckTitleChangeNotify(concern.BibiliLive))
	assert.True(t, g.CheckTitleChangeNotify(concern.DouyuLive))
	assert.False(t, g.CheckTitleChangeNotify(concern.HuyaLive))
}
