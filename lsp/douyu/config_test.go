package douyu

import (
	"github.com/Sora233/DDBOT/internal/test"
	"github.com/Sora233/DDBOT/lsp/concern"
	"github.com/stretchr/testify/assert"
	"testing"
)

func newLiveInfo(roomId int64, live bool, liveStatusChanged bool, liveTitleChanged bool) *ConcernLiveNotify {
	li := &ConcernLiveNotify{
		LiveInfo: LiveInfo{
			RoomId:            roomId,
			liveStatusChanged: liveStatusChanged,
			liveTitleChanged:  liveTitleChanged,
		},
	}
	if live {
		li.ShowStatus = ShowStatus_Living
		li.VideoLoop = VideoLoopStatus_Off
	} else {
		li.VideoLoop = VideoLoopStatus_On
	}
	return li
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
			assert.Equal(t, expected[index1][index2], result.Pass)
		}
	}
}

func TestGroupConcernConfig_AtBeforeHook(t *testing.T) {
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
	var expected = []bool{
		false, false, false, false, false, false, true, true,
	}
	var config = &GroupConcernConfig{IConfig: &concern.GroupConcernConfig{}}
	for idx, n := range notify {
		hook := config.AtBeforeHook(n)
		assert.EqualValues(t, expected[idx], hook.Pass)
	}
}
