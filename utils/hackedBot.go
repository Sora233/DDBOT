package utils

import (
	miraiBot "github.com/Logiase/MiraiGo-Template/bot"
	"github.com/Mrs4s/MiraiGo/client"
)

// HackedBot 拦截一些方法方便测试
type HackedBot struct {
	Bot **miraiBot.Bot
}

func (h *HackedBot) valid() bool {
	if h == nil || h.Bot == nil || *h.Bot == nil || !(*h.Bot).Online {
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
		return nil
	}
	return (*h.Bot).FindGroup(code)
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
		return nil
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

var hackedBot = &HackedBot{Bot: &miraiBot.Instance}

func GetBot() *HackedBot {
	return hackedBot
}
