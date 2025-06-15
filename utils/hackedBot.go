package utils

import (
	"github.com/LagrangeDev/LagrangeGo/client/entity"
	"github.com/LagrangeDev/LagrangeGo/client/event"
	"github.com/samber/lo"
	"golang.org/x/exp/constraints"

	miraiBot "github.com/Sora233/MiraiGo-Template/bot"
)

// HackedBot 拦截一些方法方便测试
type HackedBot[UT, GT constraints.Integer] struct {
	Bot              **miraiBot.Bot
	testGroups       []*entity.Group
	testGroupMembers map[GT][]*entity.GroupMember
	testUin          UT
}

func (h *HackedBot[UT, GT]) valid() bool {
	if h == nil || h.Bot == nil || *h.Bot == nil || !(*h.Bot).Online.Load() {
		return false
	}
	return true
}

func (h *HackedBot[UT, GT]) FindFriend(uin UT) *entity.User {
	if !h.valid() {
		return nil
	}
	return (*h.Bot).GetCachedFriendInfo(uint32(uin))
}

func (h *HackedBot[UT, GT]) FindGroup(code GT) *entity.Group {
	if !h.valid() {
		for _, gi := range h.testGroups {
			if gi.GroupUin == uint32(code) {
				return gi
			}
		}
		return nil
	}
	return (*h.Bot).GetCachedGroupInfo(uint32(code))
}

func (h *HackedBot[UT, GT]) FindGroupMember(groupCode GT, uin UT) *entity.GroupMember {
	if !h.valid() {
		for _, gm := range h.testGroupMembers[groupCode] {
			if gm.User.Uin == uint32(uin) {
				return gm
			}
		}
		return nil
	}
	return (*h.Bot).GetCachedMemberInfo(uint32(groupCode), uint32(uin))
}

func (h *HackedBot[UT, GT]) SolveFriendRequest(req *event.NewFriendRequest, accept bool) {
	if !h.valid() {
		return
	}
	(*h.Bot).SetFriendRequest(accept, req.SourceUID)
}

func (h *HackedBot[UT, GT]) SolveGroupJoinRequest(i *event.GroupInvite, accept, _ bool, reason string) {
	if !h.valid() {
		return
	}
	b := (*h.Bot)
	msgs, err := b.GetGroupSystemMessages(false, 20, i.GroupUin)
	if err != nil {
		logger.Errorf("获取群系统消息失败: %v", err)
		return
	}
	filteredmsgs, err := b.GetGroupSystemMessages(true, 20)
	if err != nil {
		logger.Errorf("获取群系统消息失败: %v", err)
		return
	}
	for _, req := range append(msgs.InvitedRequests[:], filteredmsgs.InvitedRequests[:]...) {
		if req.Sequence != i.RequestSeq {
			continue
		}
		if req.Checked {
			logger.Warnf("处理群系统消息失败: 无法操作已处理的消息.")
			return
		}
		if accept {
			_ = b.SetGroupRequest(req.IsFiltered, entity.GroupRequestOperateAllow, req.Sequence, uint32(req.EventType), req.GroupUin, "")
		} else {
			_ = b.SetGroupRequest(req.IsFiltered, entity.GroupRequestOperateDeny, req.Sequence, uint32(req.EventType), req.GroupUin, reason)
		}
	}
}

func (h *HackedBot[UT, GT]) GetGroupList() []*entity.Group {
	if !h.valid() {
		return h.testGroups
	}
	g, _ := (*h.Bot).GetAllGroupsInfo()
	return lo.Values(g)
}

func (h *HackedBot[UT, GT]) GetFriendList() []*entity.User {
	if !h.valid() {
		return nil
	}
	return lo.Values((*h.Bot).GetCachedAllFriendsInfo())
}

func (h *HackedBot[UT, GT]) IsOnline() bool {
	return h.valid()
}

func (h *HackedBot[UT, GT]) GetUin() UT {
	if !h.valid() {
		return h.testUin
	}
	return UT((*h.Bot).Uin)
}

var hackedBot = &HackedBot[uint32, uint32]{Bot: &miraiBot.QQClient, testGroupMembers: map[uint32][]*entity.GroupMember{}}

func GetBot() *HackedBot[uint32, uint32] {
	return hackedBot
}

// TESTSetUin 仅可用于测试
func (h *HackedBot[UT, GT]) TESTSetUin(uin UT) {
	h.testUin = uin
}

// TESTAddGroup 仅可用于测试
func (h *HackedBot[UT, GT]) TESTAddGroup(groupCode GT) {
	for _, g := range h.testGroups {
		if g.GroupUin == uint32(groupCode) {
			return
		}
	}
	h.testGroups = append(h.testGroups, &entity.Group{
		GroupUin: uint32(groupCode),
	})
}

// TESTAddMember 仅可用于测试
func (h *HackedBot[UT, GT]) TESTAddMember(groupCode GT, uin UT, permission entity.GroupMemberPermission) {
	h.TESTAddGroup(groupCode)
	h.testGroupMembers[groupCode] = append(h.testGroupMembers[groupCode], &entity.GroupMember{
		User: entity.User{
			Uin: uint32(uin),
		},
		Permission: permission,
	})
}

// TESTReset 仅可用于测试
func (h *HackedBot[UT, GT]) TESTReset() {
	h.testGroups = nil
	h.testUin = 0
}
