package concern

import (
	"github.com/Sora233/DDBOT/internal/test"
	"github.com/Sora233/DDBOT/lsp/concern_type"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGroupConcernAtConfig_CheckAtAll(t *testing.T) {
	var g *GroupConcernAtConfig
	assert.False(t, g.CheckAtAll(test.BibiliLive))

	g = &GroupConcernAtConfig{
		AtAll: test.BilibiliNews,
	}
	assert.True(t, g.CheckAtAll(test.BilibiliNews))
	assert.False(t, g.CheckAtAll(test.BibiliLive))
}

func TestNewGroupConcernConfigFromString(t *testing.T) {
	var testCase = []string{
		`{"group_concern_at":{"at_all":"bilibiliLive","at_someone":[{"ctype":"bilibiliLive", "at_list":[1,2,3,4,5]}]}}`,
	}
	var expected = []*GroupConcernConfig{
		{
			GroupConcernAt: GroupConcernAtConfig{
				AtAll: test.BibiliLive,
				AtSomeone: []*AtSomeone{
					{
						Ctype:  test.BibiliLive,
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
				AtAll: test.BibiliLive,
				AtSomeone: []*AtSomeone{
					{
						Ctype:  test.BibiliLive,
						AtList: []int64{1, 2, 3, 4, 5},
					},
				},
			},
			GroupConcernNotify: GroupConcernNotifyConfig{
				TitleChangeNotify: test.BibiliLive,
				OfflineNotify:     test.DouyuLive,
			},
		},
	}
	var expected = []string{
		`{
			"group_concern_at":{
				"at_all":"bilibiliLive",
				"at_someone":[{"ctype":"bilibiliLive", "at_list":[1,2,3,4,5]}]
			},
			"group_concern_notify":{
				"title_change_notify": "bilibiliLive", "offline_notify": "douyuLive"
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
				AtAll: test.BibiliLive,
				AtSomeone: []*AtSomeone{
					{
						Ctype:  test.BibiliLive,
						AtList: []int64{1, 2, 3, 4, 5},
					},
				},
			},
		},
		{
			GroupConcernAt: GroupConcernAtConfig{
				AtAll:     test.BibiliLive,
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
		assert.EqualValues(t, expected[i], testCase[i].GroupConcernAt.GetAtSomeoneList(test.BibiliLive))
	}

	var g *GroupConcernAtConfig
	assert.Nil(t, g.GetAtSomeoneList(test.BibiliLive))
}

func TestGroupConcernNotifyConfig_CheckTitleChangeNotify(t *testing.T) {
	var g = &GroupConcernNotifyConfig{
		TitleChangeNotify: concern_type.Empty.Add(test.BibiliLive, test.DouyuLive),
	}
	assert.True(t, g.CheckTitleChangeNotify(test.BibiliLive))
	assert.True(t, g.CheckTitleChangeNotify(test.DouyuLive))
	assert.False(t, g.CheckTitleChangeNotify(test.HuyaLive))
}

func TestGroupConcernAtConfig_ClearAtSomeoneList(t *testing.T) {
	var g = &GroupConcernAtConfig{
		AtAll: concern_type.Empty,
		AtSomeone: []*AtSomeone{
			{
				Ctype:  test.BibiliLive,
				AtList: []int64{1, 2, 3, 4},
			},
		},
	}
	g.ClearAtSomeoneList(test.DouyuLive)
	for i := 1; i <= 4; i++ {
		assert.Contains(t, g.GetAtSomeoneList(test.BibiliLive), int64(i))
	}
	g.ClearAtSomeoneList(test.BibiliLive)
	assert.Equal(t, 0, len(g.GetAtSomeoneList(test.BibiliLive)))
}

func TestGroupConcernAtConfig_RemoveAtSomeoneList(t *testing.T) {
	var g = &GroupConcernAtConfig{
		AtAll: concern_type.Empty,
		AtSomeone: []*AtSomeone{
			{
				Ctype:  test.BibiliLive,
				AtList: []int64{1, 2, 3, 4},
			},
		},
	}
	g.RemoveAtSomeoneList(test.DouyuLive, []int64{1, 2, 3, 4})
	assert.EqualValues(t, []int64{1, 2, 3, 4}, g.GetAtSomeoneList(test.BibiliLive))
	g.RemoveAtSomeoneList(test.BibiliLive, []int64{3})
	for i := 1; i <= 4; i++ {
		if i != 3 {
			assert.Contains(t, g.GetAtSomeoneList(test.BibiliLive), int64(i))
		}
	}
	assert.NotContains(t, g.GetAtSomeoneList(test.BibiliLive), int64(3))
}

func TestGroupConcernAtConfig_MergeAtSomeoneList(t *testing.T) {
	var g = &GroupConcernAtConfig{
		AtAll:     concern_type.Empty,
		AtSomeone: nil,
	}

	g.MergeAtSomeoneList(test.BibiliLive, []int64{1, 2, 3, 4})
	for i := 1; i <= 4; i++ {
		assert.Contains(t, g.GetAtSomeoneList(test.BibiliLive), int64(i))
	}
	g.MergeAtSomeoneList(test.BibiliLive, []int64{3, 4, 5})
	for i := 1; i <= 5; i++ {
		assert.Contains(t, g.GetAtSomeoneList(test.BibiliLive), int64(i))
	}
	assert.EqualValues(t, 0, len(g.GetAtSomeoneList(test.DouyuLive)))
}

func TestGroupConcernAtConfig_SetAtSomeoneList(t *testing.T) {
	var g = &GroupConcernAtConfig{
		AtAll:     concern_type.Empty,
		AtSomeone: nil,
	}

	g.SetAtSomeoneList(test.BibiliLive, []int64{1, 2, 3, 4})
	for i := 1; i <= 4; i++ {
		assert.Contains(t, g.GetAtSomeoneList(test.BibiliLive), int64(i))
	}

	g.SetAtSomeoneList(test.BibiliLive, []int64{5, 6})
	for i := 1; i <= 6; i++ {
		if i <= 4 {
			assert.NotContains(t, g.GetAtSomeoneList(test.BibiliLive), int64(i))
		} else {
			assert.Contains(t, g.GetAtSomeoneList(test.BibiliLive), int64(i))
		}
	}
}

func TestGroupConcernNotifyConfig_CheckOfflineNotify(t *testing.T) {
	var g = &GroupConcernNotifyConfig{
		OfflineNotify: concern_type.Empty.Add(test.BibiliLive, test.DouyuLive),
	}
	assert.True(t, g.CheckOfflineNotify(test.BibiliLive))
	assert.True(t, g.CheckOfflineNotify(test.DouyuLive))
	assert.False(t, g.CheckOfflineNotify(test.HuyaLive))
}

func TestGroupConcernFilterConfig_GetFilter(t *testing.T) {
	var g GroupConcernConfig
	_, err := g.GroupConcernFilter.GetFilterByType()
	assert.NotNil(t, err)

	assert.True(t, g.GroupConcernFilter.Empty())

	g.GroupConcernFilter.Type = FilterTypeType
	g.GroupConcernFilter.Config = new(GroupConcernFilterConfigByType).ToString()

	_, err = g.GroupConcernFilter.GetFilterByType()
	assert.Nil(t, err)

	_, err = g.GroupConcernFilter.GetFilterByText()
	assert.NotNil(t, err)

	assert.False(t, g.GroupConcernFilter.Empty())

	g.GroupConcernFilter.Type = FilterTypeText
	g.GroupConcernFilter.Config = new(GroupConcernFilterConfigByText).ToString()

	_, err = g.GroupConcernFilter.GetFilterByText()
	assert.Nil(t, err)

	_, err = g.GroupConcernFilter.GetFilterByType()
	assert.NotNil(t, err)

	assert.False(t, g.GroupConcernFilter.Empty())

}
