package bilibili

import (
	"github.com/Sora233/DDBOT/concern"
	"github.com/Sora233/DDBOT/internal/test"
	"github.com/Sora233/DDBOT/lsp/concern_manager"
	"github.com/stretchr/testify/assert"
	"testing"
)

func newLiveInfo(uid int64, living bool, liveStatusChanged bool, liveTitleChanged bool) *ConcernLiveNotify {
	notify := &ConcernLiveNotify{
		LiveInfo: LiveInfo{
			UserInfo: UserInfo{
				Mid: uid,
			},
			LiveStatusChanged: liveStatusChanged,
			LiveTitleChanged:  liveTitleChanged,
		},
	}
	if living {
		notify.Status = LiveStatus_Living
	} else {
		notify.Status = LiveStatus_NoLiving
	}
	return notify
}

func newNewsInfo(uid int64, cardTypes ...DynamicDescType) []*ConcernNewsNotify {

	var result []*ConcernNewsNotify
	for _, t := range cardTypes {
		notify := &ConcernNewsNotify{
			UserInfo: &UserInfo{
				Mid: uid,
			},
		}
		notify.Card = &Card{
			Desc: &Card_Desc{
				Type: t,
			},
		}
		result = append(result, notify)
	}
	return result
}

func TestNewGroupConcernConfig(t *testing.T) {
	g := NewGroupConcernConfig(new(concern_manager.GroupConcernConfig), nil)
	assert.NotNil(t, g)
}

func TestGroupConcernConfig_ShouldSendHook(t *testing.T) {
	var notify = []concern.Notify{
		// 下播状态 什么也没变 不推
		newLiveInfo(test.UID1, false, false, false),
		// 下播状态 标题变了 不推
		newLiveInfo(test.UID1, false, false, true),
		// 下播了 检查配置
		newLiveInfo(test.UID1, false, true, false),
		// 下播了 检查配置
		newLiveInfo(test.UID1, false, true, true),
		// 直播状态 什么也没变 不推
		newLiveInfo(test.UID1, true, false, false),
		// 直播状态 改了标题 检查配置
		newLiveInfo(test.UID1, true, false, true),
		// 开播了 推
		newLiveInfo(test.UID1, true, true, false),
		// 开播了改了标题 推
		newLiveInfo(test.UID1, true, true, true),
	}
	// 无法处理news，应该pass
	for _, card := range newNewsInfo(test.UID1, DynamicDescType_TextOnly) {
		notify = append(notify, card)
	}
	var testCase = []*GroupConcernConfig{
		{},
		{
			GroupConcernConfig: concern_manager.GroupConcernConfig{
				GroupConcernNotify: concern_manager.GroupConcernNotifyConfig{
					TitleChangeNotify: test.BibiliLive,
				},
			},
		},
		{
			GroupConcernConfig: concern_manager.GroupConcernConfig{
				GroupConcernNotify: concern_manager.GroupConcernNotifyConfig{
					OfflineNotify: test.BibiliLive,
				},
			},
		},
		{
			GroupConcernConfig: concern_manager.GroupConcernConfig{
				GroupConcernNotify: concern_manager.GroupConcernNotifyConfig{
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
			assert.NotNil(t, result)
			assert.Equal(t, expected[index1][index2], result.Pass)
		}
	}
}

func TestGroupConcernConfig_AtBeforeHook(t *testing.T) {
	var liveInfos = []*ConcernLiveNotify{
		// 下播状态 什么也没变 不推
		newLiveInfo(test.UID1, false, false, false),
		// 下播状态 标题变了 不推
		newLiveInfo(test.UID1, false, false, true),
		// 下播了 检查配置
		newLiveInfo(test.UID1, false, true, false),
		// 下播了 检查配置
		newLiveInfo(test.UID1, false, true, true),
		// 直播状态 什么也没变 不推
		newLiveInfo(test.UID1, true, false, false),
		// 直播状态 改了标题 检查配置
		newLiveInfo(test.UID1, true, false, true),
		// 开播了 推
		newLiveInfo(test.UID1, true, true, false),
		// 开播了改了标题 推
		newLiveInfo(test.UID1, true, true, true),
	}
	var g = new(GroupConcernConfig)
	var expected = []bool{
		false, false, false, false,
		false, false, true, true,
	}
	assert.Equal(t, len(expected), len(liveInfos))
	for index, liveInfo := range liveInfos {
		result := g.AtBeforeHook(liveInfo)
		assert.Equal(t, expected[index], result.Pass)
	}
}

func TestGroupConcernConfig_NewsFilterHook(t *testing.T) {
	var notifies = newNewsInfo(test.UID1, DynamicDescType_WithOrigin, DynamicDescType_WithImage, DynamicDescType_TextOnly)
	var g = new(GroupConcernConfig)

	// 默认应该不过滤
	for _, notify := range notifies {
		assert.True(t, g.NewsFilterHook(notify).Pass)
	}

	var typeFilter = []*concern_manager.GroupConcernFilterConfigByType{
		{
			Type: []string{
				Zhuanfa,
			},
		},
		{
			Type: []string{
				Tupian,
			},
		},
		{
			Type: []string{
				Tupian, Wenzi,
			},
		},
		{
			Type: []string{
				Zhibofenxiang,
			},
		},
	}

	var expectedType = [][]DynamicDescType{
		{
			DynamicDescType_WithOrigin,
		},
		{
			DynamicDescType_WithImage,
		},
		{
			DynamicDescType_WithImage, DynamicDescType_TextOnly,
		},
		nil,
	}

	var expectedNotType = [][]DynamicDescType{
		{
			DynamicDescType_WithImage, DynamicDescType_TextOnly,
		},
		{
			DynamicDescType_WithOrigin, DynamicDescType_TextOnly,
		},
		{
			DynamicDescType_WithOrigin,
		},
		{
			DynamicDescType_WithOrigin, DynamicDescType_WithImage, DynamicDescType_TextOnly,
		},
	}

	testFn := func(index int, tp string, expected []DynamicDescType) {
		notifies := newNewsInfo(test.UID1, DynamicDescType_WithOrigin, DynamicDescType_WithImage, DynamicDescType_TextOnly)
		var g = new(GroupConcernConfig)
		g.GroupConcernFilter = concern_manager.GroupConcernFilterConfig{
			Type:   tp,
			Config: typeFilter[index].ToString(),
		}
		var resultType []DynamicDescType
		for _, notify := range notifies {
			hookResult := g.NewsFilterHook(notify)
			if hookResult.Pass {
				resultType = append(resultType, notify.Card.GetDesc().GetType())
			}
		}
		assert.EqualValues(t, expected, resultType)
	}

	for index := range typeFilter {
		testFn(index, concern_manager.FilterTypeType, expectedType[index])
		testFn(index, concern_manager.FilterTypeNotType, expectedNotType[index])
	}
}

func TestCheckTypeDefine(t *testing.T) {
	result := CheckTypeDefine([]string{"invalid", Zhuanlan, "1024", "0", "9"})
	assert.Len(t, result, 3)
	assert.EqualValues(t, []string{"invalid", "0", "9"}, result)
}
