package concern

import (
	"github.com/Sora233/DDBOT/internal/test"
	"github.com/Sora233/DDBOT/lsp/concern_type"
	"github.com/Sora233/DDBOT/lsp/mmsg"
	"github.com/Sora233/DDBOT/lsp/mmsg/mt"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGroupConcernAtConfig_CheckAtAll(t *testing.T) {
	var g *ConcernAtConfig
	assert.False(t, g.CheckAtAll(test.BibiliLive))

	g = &ConcernAtConfig{
		AtAll: test.BilibiliNews,
	}
	assert.True(t, g.CheckAtAll(test.BilibiliNews))
	assert.False(t, g.CheckAtAll(test.BibiliLive))
}

func TestNewGroupConcernConfigFromString(t *testing.T) {
	var testCase = []string{
		`{"concern_at_map":{"Group": {"at_all":"bilibiliLive","at_someone":[{"ctype":"bilibiliLive", "at_list":[1,2,3,4,5]}]}}}`,
	}
	var expected = []*ConcernConfig{
		{
			ConcernAtMap: map[mt.TargetType]*ConcernAtConfig{
				mt.TargetGroup: {
					AtAll: test.BibiliLive,
					AtSomeone: []*AtSomeone{
						{
							Ctype:  test.BibiliLive,
							AtList: []int64{1, 2, 3, 4, 5},
						},
					},
				},
			},
		},
	}
	assert.Equal(t, len(testCase), len(expected))
	for i := 0; i < len(testCase); i++ {
		g, err := NewConcernConfigFromString(testCase[i])
		assert.Nil(t, err)
		assert.EqualValues(t, expected[i], g)
	}
}

func TestGroupConcernConfig_ToString(t *testing.T) {
	var testCase = []*ConcernConfig{
		{
			ConcernAtMap: map[mt.TargetType]*ConcernAtConfig{
				mt.TargetGroup: {
					AtAll: test.BibiliLive,
					AtSomeone: []*AtSomeone{
						{
							Ctype:  test.BibiliLive,
							AtList: []int64{1, 2, 3, 4, 5},
						},
					},
				},
			},
			ConcernNotifyMap: map[mt.TargetType]*ConcernNotifyConfig{
				mt.TargetGroup: {
					TitleChangeNotify: test.BibiliLive,
					OfflineNotify:     test.DouyuLive,
				},
			},
			ConcernFilterMap: map[mt.TargetType]*ConcernFilterConfig{
				mt.TargetGroup: {},
			},
		},
	}
	var expected = []string{
		`{
			"concern_at_map":{"Group": {
				"at_all":"bilibiliLive",
				"at_someone":[{"ctype":"bilibiliLive", "at_list":[1,2,3,4,5]}]
			}},
			"concern_notify_map":{"Group": {
				"title_change_notify": "bilibiliLive", "offline_notify": "douyuLive"
			}},
			"concern_filter_map": {"Group": {
				"type": "", "config":""
			}}
		}`,
	}
	assert.Equal(t, len(testCase), len(expected))
	for i := 0; i < len(testCase); i++ {
		assert.JSONEq(t, expected[i], testCase[i].ToString())
	}
}

func TestGroupConcernAtConfig_GetAtSomeoneList(t *testing.T) {
	var testCase = []*ConcernConfig{
		{
			ConcernAtMap: map[mt.TargetType]*ConcernAtConfig{
				mt.TargetGroup: {
					AtAll: test.BibiliLive,
					AtSomeone: []*AtSomeone{
						{
							Ctype:  test.BibiliLive,
							AtList: []int64{1, 2, 3, 4, 5},
						},
					},
				},
			},
		},
		{
			ConcernAtMap: map[mt.TargetType]*ConcernAtConfig{
				mt.TargetGroup: {
					AtAll:     test.BibiliLive,
					AtSomeone: nil,
				},
			},
		},
	}
	var expected = [][]int64{
		{1, 2, 3, 4, 5},
		nil,
	}
	assert.Equal(t, len(testCase), len(expected))
	for i := 0; i < len(testCase); i++ {
		assert.EqualValues(t, expected[i], testCase[i].ConcernAtMap[mt.TargetGroup].GetAtSomeoneList(test.BibiliLive))
	}

	var g *ConcernAtConfig
	assert.Nil(t, g.GetAtSomeoneList(test.BibiliLive))
}

func TestGroupConcernNotifyConfig_CheckTitleChangeNotify(t *testing.T) {
	var g = &ConcernNotifyConfig{
		TitleChangeNotify: concern_type.Empty.Add(test.BibiliLive, test.DouyuLive),
	}
	assert.True(t, g.CheckTitleChangeNotify(test.BibiliLive))
	assert.True(t, g.CheckTitleChangeNotify(test.DouyuLive))
	assert.False(t, g.CheckTitleChangeNotify(test.HuyaLive))
}

func TestGroupConcernAtConfig_ClearAtSomeoneList(t *testing.T) {
	var g = &ConcernAtConfig{
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
	var g = &ConcernAtConfig{
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
	var g *ConcernAtConfig
	g.MergeAtSomeoneList("", nil)
	g.SetAtSomeoneList("", nil)
	g.RemoveAtSomeoneList("", nil)
	g.ClearAtSomeoneList("")
	g = &ConcernAtConfig{
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
	var g = &ConcernAtConfig{
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
	var g = &ConcernNotifyConfig{
		OfflineNotify: concern_type.Empty.Add(test.BibiliLive, test.DouyuLive),
	}
	assert.True(t, g.CheckOfflineNotify(test.BibiliLive))
	assert.True(t, g.CheckOfflineNotify(test.DouyuLive))
	assert.False(t, g.CheckOfflineNotify(test.HuyaLive))
}

func TestGroupConcernFilterConfig_GetFilter(t *testing.T) {
	var g ConcernConfig
	assert.NotNil(t, g.GetConcernNotify(mt.TargetGroup))
	assert.NotNil(t, g.GetConcernAt(mt.TargetGroup))
	assert.NotNil(t, g.GetConcernFilter(mt.TargetGroup))

	_, err := g.GetConcernFilter(mt.TargetGroup).GetFilterByType()
	assert.NotNil(t, err)

	assert.True(t, g.GetConcernFilter(mt.TargetGroup).Empty())

	g.GetConcernFilter(mt.TargetGroup).Type = FilterTypeType
	g.GetConcernFilter(mt.TargetGroup).Config = new(GroupConcernFilterConfigByType).ToString()

	_, err = g.GetConcernFilter(mt.TargetGroup).GetFilterByType()
	assert.Nil(t, err)

	_, err = g.GetConcernFilter(mt.TargetGroup).GetFilterByText()
	assert.NotNil(t, err)

	assert.False(t, g.GetConcernFilter(mt.TargetGroup).Empty())

	g.GetConcernFilter(mt.TargetGroup).Type = FilterTypeText
	g.GetConcernFilter(mt.TargetGroup).Config = new(GroupConcernFilterConfigByText).ToString()

	_, err = g.GetConcernFilter(mt.TargetGroup).GetFilterByText()
	assert.Nil(t, err)

	_, err = g.GetConcernFilter(mt.TargetGroup).GetFilterByType()
	assert.NotNil(t, err)

	assert.False(t, g.GetConcernFilter(mt.TargetGroup).Empty())
}

func TestGroupConcernConfig_Validate(t *testing.T) {
	var g ConcernConfig
	assert.Nil(t, g.Validate())
	g.GetConcernFilter(mt.TargetGroup).Type = FilterTypeType
	g.GetConcernFilter(mt.TargetGroup).Config = "wrong"
	assert.NotNil(t, g.Validate())
}

type testInfo struct {
	isLive        bool
	living        bool
	titleChanged  bool
	statusChanged bool
	uid           int64
	target        mt.Target
	t             concern_type.Type
}

func (t *testInfo) Site() string {
	return testSite
}

func (t *testInfo) Type() concern_type.Type {
	if t.t.Empty() {
		return test.BibiliLive
	}
	return t.t
}

func (t *testInfo) GetUid() interface{} {
	return t.uid
}

func (t *testInfo) Logger() *logrus.Entry {
	return logrus.WithField("Site", t.Site())
}

func (t *testInfo) GetTarget() mt.Target {
	return t.target
}

func (t *testInfo) ToMessage() *mmsg.MSG {
	return mmsg.NewMSG()
}

func (t *testInfo) TitleChanged() bool {
	return t.titleChanged
}

func (t *testInfo) IsLive() bool {
	return t.isLive
}

func (t *testInfo) Living() bool {
	return t.living
}

func (t *testInfo) LiveStatusChanged() bool {
	return t.statusChanged
}

func newGroupLiveInfo(uid int64, living, liveStatusChanged, liveTitleChanged bool) *testInfo {
	return &testInfo{
		isLive:        true,
		living:        living,
		titleChanged:  liveTitleChanged,
		statusChanged: liveStatusChanged,
		uid:           uid,
		target:        mt.NewGroupTarget(test.G1),
	}
}

func TestGroupConcernConfig_ShouldSendHook(t *testing.T) {
	test.InitBuntdb(t)
	defer test.CloseBuntdb(t)

	var notify = []Notify{
		// 下播状态 什么也没变 不推
		newGroupLiveInfo(test.UID1, false, false, false),
		// 下播状态 标题变了 不推
		newGroupLiveInfo(test.UID1, false, false, true),
		// 下播了 检查配置
		newGroupLiveInfo(test.UID1, false, true, false),
		// 下播了 检查配置
		newGroupLiveInfo(test.UID1, false, true, true),
		// 直播状态 什么也没变 不推
		newGroupLiveInfo(test.UID1, true, false, false),
		// 直播状态 改了标题 检查配置
		newGroupLiveInfo(test.UID1, true, false, true),
		// 开播了 推
		newGroupLiveInfo(test.UID1, true, true, false),
		// 开播了改了标题 推
		newGroupLiveInfo(test.UID1, true, true, true),
	}
	// 其他类型应该pass
	notify = append(notify, &testInfo{
		uid: test.UID2,
		t:   test.BilibiliNews,
	})
	var testCase = []*ConcernConfig{
		{},
		{
			ConcernNotifyMap: map[mt.TargetType]*ConcernNotifyConfig{
				mt.TargetGroup: {
					TitleChangeNotify: test.BibiliLive,
				},
			},
		},
		{

			ConcernNotifyMap: map[mt.TargetType]*ConcernNotifyConfig{
				mt.TargetGroup: {
					OfflineNotify: test.BibiliLive,
				},
			},
		},
		{
			ConcernNotifyMap: map[mt.TargetType]*ConcernNotifyConfig{
				mt.TargetGroup: {
					OfflineNotify:     test.BibiliLive,
					TitleChangeNotify: test.BibiliLive,
				},
			},
		},
	}
	var expected = [][]bool{
		{
			false, false, false, false,
			false, false, true, true,
			true,
		},
		{
			false, false, false, false,
			false, true, true, true,
			true,
		},
		{
			false, false, true, true,
			false, false, true, true,
			true,
		},
		{
			false, false, true, true,
			false, true, true, true,
			true,
		},
	}
	assert.Equal(t, len(expected), len(testCase))
	for index1, g := range testCase {
		assert.Equal(t, len(expected[index1]), len(notify))
		for index2, liveInfo := range notify {
			result := g.ShouldSendHook(liveInfo)
			if index1 == 1 && index2 == 5 {
				result = g.ShouldSendHook(liveInfo)
			}
			assert.NotNil(t, result)
			assert.Equalf(t, expected[index1][index2], result.Pass, "%v and %v check fail", index1, index2)
		}
	}
}

func TestGroupConcernConfig_AtBeforeHook(t *testing.T) {
	var liveInfos = []Notify{
		// 下播状态 什么也没变 不推
		newGroupLiveInfo(test.UID1, false, false, false),
		// 下播状态 标题变了 不推
		newGroupLiveInfo(test.UID1, false, false, true),
		// 下播了 检查配置
		newGroupLiveInfo(test.UID1, false, true, false),
		// 下播了 检查配置
		newGroupLiveInfo(test.UID1, false, true, true),
		// 直播状态 什么也没变 不推
		newGroupLiveInfo(test.UID1, true, false, false),
		// 直播状态 改了标题 检查配置
		newGroupLiveInfo(test.UID1, true, false, true),
		// 开播了 推
		newGroupLiveInfo(test.UID1, true, true, false),
		// 开播了改了标题 推
		newGroupLiveInfo(test.UID1, true, true, true),
		// news 默认pass
		&testInfo{
			t: test.BilibiliNews,
		},
	}
	var g = &ConcernConfig{}
	var expected = []bool{
		false, false, false, false,
		false, false, true, true,
		true,
	}
	assert.Equal(t, len(expected), len(liveInfos))
	for index, liveInfo := range liveInfos {
		result := g.AtBeforeHook(liveInfo)
		assert.Equalf(t, expected[index], result.Pass, "%v check fail", index)
	}
}

func TestGroupConcernConfig_FilterHook(t *testing.T) {
	var g = &ConcernConfig{}
	result := g.FilterHook(newGroupLiveInfo(test.UID1, true, true, true))
	assert.True(t, result.Pass)
}
