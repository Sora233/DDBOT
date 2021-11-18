package acfun

import (
	"github.com/Sora233/DDBOT/internal/test"
	"github.com/Sora233/DDBOT/lsp/concern"
	"github.com/stretchr/testify/assert"
	"testing"
)

func newLiveInfo(uid int64, living bool, liveStatusChanged bool, liveTitleChanged bool) *ConcernLiveNotify {
	notify := &ConcernLiveNotify{
		LiveInfo: &LiveInfo{
			UserInfo: UserInfo{
				Uid: uid,
			},
			IsLiving:          living,
			liveStatusChanged: liveStatusChanged,
			liveTitleChanged:  liveTitleChanged,
		},
	}
	return notify
}

func TestNewGroupConcernConfig(t *testing.T) {
	g := NewGroupConcernConfig(new(concern.GroupConcernConfig))
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
	var testCase = []*GroupConcernConfig{
		{
			IConfig: &concern.GroupConcernConfig{},
		},
		{
			IConfig: &concern.GroupConcernConfig{
				GroupConcernNotify: concern.GroupConcernNotifyConfig{
					TitleChangeNotify: Live,
				},
			},
		},
		{
			IConfig: &concern.GroupConcernConfig{
				GroupConcernNotify: concern.GroupConcernNotifyConfig{
					OfflineNotify: Live,
				},
			},
		},
		{
			IConfig: &concern.GroupConcernConfig{
				GroupConcernNotify: concern.GroupConcernNotifyConfig{
					OfflineNotify:     Live,
					TitleChangeNotify: Live,
				},
			},
		},
	}
	var expected = [][]bool{
		{
			false, false, false, false,
			false, false, true, true,
		},
		{
			false, false, false, false,
			false, true, true, true,
		},
		{
			false, false, true, true,
			false, false, true, true,
		},
		{
			false, false, true, true,
			false, true, true, true,
		},
	}
	assert.Equal(t, len(expected), len(testCase))
	for index1, g := range testCase {
		assert.Equal(t, len(expected[index1]), len(notify))
		for index2, liveInfo := range notify {
			result := g.ShouldSendHook(liveInfo)
			assert.NotNil(t, result)
			assert.Equalf(t, expected[index1][index2], result.Pass, "%v and %v check fail", index1, index2)
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
	var g = NewGroupConcernConfig(new(concern.GroupConcernConfig))
	var expected = []bool{
		false, false, false, false,
		false, false, true, true,
	}
	assert.Equal(t, len(expected), len(liveInfos))
	for index, liveInfo := range liveInfos {
		result := g.AtBeforeHook(liveInfo)
		assert.Equalf(t, expected[index], result.Pass, "%v check fail", index)
	}
}
