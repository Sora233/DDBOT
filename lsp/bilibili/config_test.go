package bilibili

import (
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Sora233/DDBOT/internal/test"
	localdb "github.com/Sora233/DDBOT/lsp/buntdb"
	"github.com/Sora233/DDBOT/lsp/concern"
	"github.com/Sora233/DDBOT/lsp/mmsg/mt"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
)

func newGroupLiveInfo(uid int64, living bool, liveStatusChanged bool, liveTitleChanged bool) *ConcernLiveNotify {
	notify := &ConcernLiveNotify{
		LiveInfo: &LiveInfo{
			UserInfo: UserInfo{
				Mid: uid,
			},
			liveStatusChanged: liveStatusChanged,
			liveTitleChanged:  liveTitleChanged,
		},
	}
	if living {
		notify.Status = LiveStatus_Living
	} else {
		notify.Status = LiveStatus_NoLiving
	}
	notify.Target = mt.NewGroupTarget(test.G1)
	return notify
}

func newGroupNewsInfo(uid int64, cardTypes ...DynamicDescType) []*ConcernNewsNotify {
	var result []*ConcernNewsNotify
	for _, t := range cardTypes {
		notify := &ConcernNewsNotify{
			UserInfo: &UserInfo{
				Mid: uid,
			},
			Target:  mt.NewGroupTarget(test.G1),
			concern: NewConcern(concern.GetNotifyChan()),
		}
		notify.Card = NewCacheCard(&Card{
			Desc: &Card_Desc{
				Type: t,
			},
		})
		result = append(result, notify)
	}
	return result
}

func TestNewGroupConcernConfig(t *testing.T) {
	test.InitBuntdb(t)
	defer test.CloseBuntdb(t)

	c := initConcern(t)

	g := c.GetConcernConfig(mt.NewGroupTarget(test.G1), test.UID1)

	assert.NotNil(t, g)
	assert.Nil(t, g.Validate())

	g.GetConcernFilter(mt.TargetGroup).Type = concern.FilterTypeNotType
	g.GetConcernFilter(mt.TargetGroup).Config = (&concern.GroupConcernFilterConfigByType{Type: []string{"q", "a"}}).ToString()

	assert.NotNil(t, g.Validate())
	g.GetConcernFilter(mt.TargetGroup).Config = "wrong"
	assert.NotNil(t, g.Validate())
	g.GetConcernFilter(mt.TargetGroup).Config = ""
	g.GetConcernFilter(mt.TargetGroup).Type = ""
	assert.Nil(t, g.Validate())

	g = c.GetConcernConfig(mt.NewGroupTarget(test.G1), test.UID1)
	err := c.OperateConcernConfig(mt.NewGroupTarget(test.G1), test.UID1, g, func(concernConfig concern.IConfig) bool {
		concernConfig.GetConcernFilter(mt.TargetGroup).Type = concern.FilterTypeNotType
		concernConfig.GetConcernFilter(mt.TargetGroup).Config = (&concern.GroupConcernFilterConfigByType{Type: []string{"wrong"}}).ToString()
		return true
	})
	assert.NotNil(t, err)

	g = c.GetConcernConfig(mt.NewGroupTarget(test.G1), test.UID1)
	err = c.OperateConcernConfig(mt.NewGroupTarget(test.G1), test.UID1, g, func(concernConfig concern.IConfig) bool {
		concernConfig.GetConcernFilter(mt.TargetGroup).Type = concern.FilterTypeNotType
		concernConfig.GetConcernFilter(mt.TargetGroup).Config = (&concern.GroupConcernFilterConfigByType{Type: []string{Tougao}}).ToString()
		return true
	})
	assert.Nil(t, err)
}

func TestGroupConcernConfig_ShouldSendHook(t *testing.T) {
	test.InitBuntdb(t)
	defer test.CloseBuntdb(t)

	var notify = []concern.Notify{
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
		// 无法处理news，应该pass
		newGroupNewsInfo(test.UID1, DynamicDescType_TextOnly)[0],
	}
	var testCase = []*ConcernConfig{
		{
			IConfig: &concern.ConcernConfig{},
		},
		{
			IConfig: &concern.ConcernConfig{
				ConcernNotifyMap: map[mt.TargetType]*concern.ConcernNotifyConfig{
					mt.TargetGroup: {
						TitleChangeNotify: Live,
					},
				},
			},
		},
		{
			IConfig: &concern.ConcernConfig{
				ConcernNotifyMap: map[mt.TargetType]*concern.ConcernNotifyConfig{
					mt.TargetGroup: {
						OfflineNotify: Live,
					},
				},
			},
		},
		{
			IConfig: &concern.ConcernConfig{
				ConcernNotifyMap: map[mt.TargetType]*concern.ConcernNotifyConfig{
					mt.TargetGroup: {
						OfflineNotify:     Live,
						TitleChangeNotify: Live,
					},
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
			assert.Equalf(t, expected[index1][index2], result.Pass, "%v and %v check fail", index1, index2)
		}
	}
}

func TestGroupConcernConfig_AtBeforeHook(t *testing.T) {
	var liveInfos = []concern.Notify{
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
		newGroupNewsInfo(test.UID1, DynamicDescType_TextOnly)[0],
	}
	var g = NewConcernConfig(new(concern.ConcernConfig), NewConcern(concern.GetNotifyChan()))
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

	g.concern.unsafeStart.Store(true)
	for index, liveInfo := range liveInfos {
		result := g.AtBeforeHook(liveInfo)
		assert.Equalf(t, false, result.Pass, "%v check fail", index)
	}
}

func TestGroupConcernConfig_NewsFilterHook(t *testing.T) {
	var notifies = newGroupNewsInfo(test.UID1, DynamicDescType_WithOrigin, DynamicDescType_WithImage, DynamicDescType_TextOnly)
	var g = NewConcernConfig(new(concern.ConcernConfig), nil)

	// 默认应该不过滤
	for _, notify := range notifies {
		assert.True(t, g.FilterHook(notify).Pass)
	}

	var typeFilter = []*concern.GroupConcernFilterConfigByType{
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
		notifies := newGroupNewsInfo(test.UID1, DynamicDescType_WithOrigin, DynamicDescType_WithImage, DynamicDescType_TextOnly)
		var g = NewConcernConfig(&concern.ConcernConfig{
			ConcernFilterMap: map[mt.TargetType]*concern.ConcernFilterConfig{
				mt.TargetGroup: {
					Type:   tp,
					Config: typeFilter[index].ToString(),
				},
			},
		}, nil)
		assert.Nil(t, g.Validate())

		var resultType []DynamicDescType
		for _, notify := range notifies {
			hookResult := g.FilterHook(notify)
			if hookResult.Pass {
				resultType = append(resultType, notify.Card.GetDesc().GetType())
			}
		}
		assert.EqualValues(t, expected, resultType)
	}

	for index := range typeFilter {
		testFn(index, concern.FilterTypeType, expectedType[index])
		testFn(index, concern.FilterTypeNotType, expectedNotType[index])
	}

	live := newGroupLiveInfo(test.UID1, true, false, false)
	g.FilterHook(live)
}

func TestCheckTypeDefine(t *testing.T) {
	result := CheckTypeDefine([]string{"invalid", Zhuanlan, "1024", "0", "9"})
	assert.Len(t, result, 3)
	assert.EqualValues(t, []string{"invalid", "0", "9"}, result)
}

func TestGroupConcernConfig_NotifyBeforeCallback(t *testing.T) {
	test.InitBuntdb(t)
	defer test.CloseBuntdb(t)

	c := initConcern(t)

	_, err := c.GetNotifyMsg(mt.NewGroupTarget(test.G1), test.BVID1)
	assert.True(t, localdb.IsNotFound(err))

	var notify = newGroupNewsInfo(test.UID1, DynamicDescType_WithOrigin)[0]
	notify.Card.Desc.OrigDyIdStr = test.BVID1

	var g = new(ConcernConfig)
	g.concern = c
	g.NotifyBeforeCallback(notify)
	assert.False(t, notify.shouldCompact)
	g.NotifyBeforeCallback(notify)
	assert.True(t, notify.shouldCompact)

	notify = newGroupNewsInfo(test.UID1, DynamicDescType_WithVideo)[0]
	notify.Card.Desc.Bvid = test.BVID2

	g.NotifyBeforeCallback(notify)
	assert.False(t, notify.shouldCompact)
	g.NotifyBeforeCallback(notify)
	assert.True(t, notify.shouldCompact)

	notify = newGroupNewsInfo(test.UID1, DynamicDescType_TextOnly)[0]
	notify.Card.Desc.DynamicIdStr = strconv.FormatInt(test.DynamicID1, 10)

	g.NotifyBeforeCallback(notify)
	assert.False(t, notify.shouldCompact)
	g.NotifyBeforeCallback(notify)
	assert.False(t, notify.shouldCompact)

	live := newGroupLiveInfo(test.UID1, true, false, false)
	g.NotifyBeforeCallback(live)
}

func TestGroupConcernConfig_NotifyAfterCallback(t *testing.T) {
	test.InitBuntdb(t)
	defer test.CloseBuntdb(t)

	c := initConcern(t)

	_, err := c.GetNotifyMsg(mt.NewGroupTarget(test.G1), test.BVID1)
	assert.True(t, localdb.IsNotFound(err))

	var notify = newGroupNewsInfo(test.UID1, DynamicDescType_WithOrigin)[0]
	notify.compactKey = test.BVID1
	var msg = &message.GroupMessage{
		Id:        1,
		GroupCode: test.G1,
		Elements: []message.IMessageElement{
			message.NewText("asd"),
		},
	}
	var g = new(ConcernConfig)
	g.concern = c

	g.NotifyAfterCallback(notify, msg)

	msg2, err := c.GetNotifyMsg(mt.NewGroupTarget(test.G1), test.BVID1)
	assert.Nil(t, err)
	assert.EqualValues(t, msg, msg2)

	notify.shouldCompact = true
	g.NotifyAfterCallback(notify, msg)

	live := newGroupLiveInfo(test.UID1, true, false, false)
	g.NotifyAfterCallback(live, nil)
}
