package utils

import (
	"github.com/Mrs4s/MiraiGo/client"
	"github.com/Sora233/DDBOT/lsp/mmsg/mt"
	miraiBot "github.com/Sora233/MiraiGo-Template/bot"
	"strconv"
)

// HackedBot 拦截一些方法方便测试
type HackedBot struct {
	Bot        **miraiBot.Bot
	testGroups []*client.GroupInfo
	testUin    int64
	testGuilds []*client.GuildInfo
}

func (h *HackedBot) valid() bool {
	if h == nil || h.Bot == nil || *h.Bot == nil || !(*h.Bot).Online.Load() {
		return false
	}
	return true
}

func (h *HackedBot) FindFriend(uin int64) *client.FriendInfo {
	if !h.valid() {
		return nil
	}
	return (*h.Bot).FindFriend(uin)
}

func (h *HackedBot) FindGroup(code int64) *client.GroupInfo {
	if !h.valid() {
		for _, gi := range h.testGroups {
			if gi.Code == code {
				return gi
			}
		}
		return nil
	}
	return (*h.Bot).FindGroup(code)
}

func (h *HackedBot) FindGuild(guildId uint64) *client.GuildInfo {
	if !h.valid() {
		for _, gi := range h.testGuilds {
			if gi.GuildId == guildId {
				return gi
			}
		}
		return nil
	}
	return (*h.Bot).GuildService.FindGuild(guildId)
}

func (h *HackedBot) FindGuildChannel(guildId, channelId uint64) (*client.GuildInfo, *client.ChannelInfo) {
	guild := h.FindGuild(guildId)
	if guild == nil {
		return nil, nil
	}
	return guild, guild.FindChannel(channelId)
}

func (h *HackedBot) FindGuildName(guildId uint64) string {
	guild := h.FindGuild(guildId)
	if guild == nil {
		return strconv.FormatUint(guildId, 10)
	}
	return guild.GuildName
}

func (h *HackedBot) FindChannelName(guildId, channelId uint64) string {
	_, channel := h.FindGuildChannel(guildId, channelId)
	if channel == nil {
		return strconv.FormatUint(channelId, 10)
	}
	return channel.ChannelName
}

func (h *HackedBot) CheckTarget(target mt.Target) bool {
	switch t := target.(type) {
	case *mt.GroupTarget:
		return h.FindGroup(t.GroupCode) != nil
	case *mt.PrivateTarget:
		return h.FindFriend(t.Uin) != nil
	case *mt.GuildTarget:
		_, c := h.FindGuildChannel(t.GuildId, t.ChannelId)
		return c != nil
	}
	return false
}

func (h *HackedBot) SolveFriendRequest(req *client.NewFriendRequest, accept bool) {
	if !h.valid() {
		return
	}
	(*h.Bot).SolveFriendRequest(req, accept)
}

func (h *HackedBot) SolveGroupJoinRequest(i interface{}, accept, block bool, reason string) {
	if !h.valid() {
		return
	}
	(*h.Bot).SolveGroupJoinRequest(i, accept, block, reason)
}

func (h *HackedBot) GetGroupList() []*client.GroupInfo {
	if !h.valid() {
		return h.testGroups
	}
	return (*h.Bot).GroupList
}

func (h *HackedBot) GetFriendList() []*client.FriendInfo {
	if !h.valid() {
		return nil
	}
	return (*h.Bot).FriendList
}

func (h *HackedBot) IsOnline() bool {
	return h.valid()
}

func (h *HackedBot) GetUin() int64 {
	if !h.valid() {
		return h.testUin
	}
	return (*h.Bot).Uin
}

var hackedBot = &HackedBot{Bot: &miraiBot.Instance}

func GetBot() *HackedBot {
	return hackedBot
}

func (h *HackedBot) CheckMember(target mt.Target, id int64) bool {
	var pass bool
	switch t := target.(type) {
	case *mt.GroupTarget:
		if gi := h.FindGroup(t.GroupCode); gi != nil && gi.FindMember(id) != nil {
			pass = true
		}
	case *mt.GuildTarget:
		if _, ci := h.FindGuildChannel(t.GuildId, t.ChannelId); ci != nil {
			if pi, _ := h.GetGuildService().FetchGuildMemberProfileInfo(t.GuildId, uint64(id)); pi != nil {
				pass = true
			}
		}
	case *mt.PrivateTarget:
		if h.FindFriend(id) != nil {
			pass = true
		}
	}
	return pass
}

func (h *HackedBot) GetGuildService() *client.GuildService {
	return (*h.Bot).GuildService
}

// TESTSetUin 仅可用于测试
func (h *HackedBot) TESTSetUin(uin int64) {
	h.testUin = uin
}

// TESTAddGroup 仅可用于测试
func (h *HackedBot) TESTAddGroup(groupCode int64) {
	for _, g := range h.testGroups {
		if g.Code == groupCode {
			return
		}
	}
	h.testGroups = append(h.testGroups, &client.GroupInfo{
		Uin:  groupCode,
		Code: groupCode,
	})
}

// TESTAddMember 仅可用于测试
func (h *HackedBot) TESTAddMember(groupCode int64, uin int64, permission client.MemberPermission) {
	h.TESTAddGroup(groupCode)
	for _, g := range h.testGroups {
		if g.Code != groupCode {
			continue
		}
		for _, m := range g.Members {
			if m.Uin == uin {
				return
			}
		}
		g.Members = append(g.Members, &client.GroupMemberInfo{
			Group:      g,
			Uin:        uin,
			Permission: permission,
		})
	}
}

// TESTReset 仅可用于测试
func (h *HackedBot) TESTReset() {
	h.testGroups = nil
	h.testUin = 0
}
