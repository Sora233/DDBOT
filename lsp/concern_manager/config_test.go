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
		`{
			"group_concern_at":{
				"at_all":1,
				"at_someone":[{"ctype":1, "at_list":[1,2,3,4,5]}]
			},
			"group_concern_notify":{
				"title_change_notify": 1, "offline_notify": 4
			},
			"group_concern_filter": {
				"type": "", "config":""
			}
		}`,
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

func TestGroupConcernAtConfig_ClearAtSomeoneList(t *testing.T) {
	var g = &GroupConcernAtConfig{
		AtAll: 0,
		AtSomeone: []*AtSomeone{
			{
				Ctype:  concern.BibiliLive,
				AtList: []int64{1, 2, 3, 4},
			},
		},
	}
	g.ClearAtSomeoneList(concern.DouyuLive)
	for i := 1; i <= 4; i++ {
		assert.Contains(t, g.GetAtSomeoneList(concern.BibiliLive), int64(i))
	}
	g.ClearAtSomeoneList(concern.BibiliLive)
	assert.Equal(t, 0, len(g.GetAtSomeoneList(concern.BibiliLive)))
}

func TestGroupConcernAtConfig_RemoveAtSomeoneList(t *testing.T) {
	var g = &GroupConcernAtConfig{
		AtAll: 0,
		AtSomeone: []*AtSomeone{
			{
				Ctype:  concern.BibiliLive,
				AtList: []int64{1, 2, 3, 4},
			},
		},
	}
	g.RemoveAtSomeoneList(concern.DouyuLive, []int64{1, 2, 3, 4})
	assert.EqualValues(t, []int64{1, 2, 3, 4}, g.GetAtSomeoneList(concern.BibiliLive))
	g.RemoveAtSomeoneList(concern.BibiliLive, []int64{3})
	for i := 1; i <= 4; i++ {
		if i != 3 {
			assert.Contains(t, g.GetAtSomeoneList(concern.BibiliLive), int64(i))
		}
	}
	assert.NotContains(t, g.GetAtSomeoneList(concern.BibiliLive), int64(3))
}

func TestGroupConcernAtConfig_MergeAtSomeoneList(t *testing.T) {
	var g = &GroupConcernAtConfig{
		AtAll: 0,
		AtSomeone: []*AtSomeone{
			{
				Ctype:  concern.BibiliLive,
				AtList: []int64{1, 2, 3, 4},
			},
		},
	}
	g.MergeAtSomeoneList(concern.BibiliLive, []int64{3, 4, 5})
	assert.Contains(t, g.GetAtSomeoneList(concern.BibiliLive), int64(5))
	assert.EqualValues(t, 0, len(g.GetAtSomeoneList(concern.DouyuLive)))
}

func TestGroupConcernAtConfig_SetAtSomeoneList(t *testing.T) {
	var g = &GroupConcernAtConfig{
		AtAll: 0,
		AtSomeone: []*AtSomeone{
			{
				Ctype:  concern.BibiliLive,
				AtList: []int64{1, 2, 3, 4},
			},
		},
	}
	g.SetAtSomeoneList(concern.BibiliLive, []int64{5, 6})
	for i := 1; i <= 6; i++ {
		if i <= 4 {
			assert.NotContains(t, g.GetAtSomeoneList(concern.BibiliLive), int64(i))
		} else {
			assert.Contains(t, g.GetAtSomeoneList(concern.BibiliLive), int64(i))
		}
	}
}

func TestGroupConcernNotifyConfig_CheckOfflineNotify(t *testing.T) {
	var g = &GroupConcernNotifyConfig{
		OfflineNotify: concern.BibiliLive | concern.DouyuLive,
	}
	assert.True(t, g.CheckOfflineNotify(concern.BibiliLive))
	assert.True(t, g.CheckOfflineNotify(concern.DouyuLive))
	assert.False(t, g.CheckOfflineNotify(concern.HuyaLive))

}
