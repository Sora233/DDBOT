package mt

import (
	"fmt"
	"strconv"
	"strings"
)

type TargetType string

const (
	TargetGroup   TargetType = "Group"
	TargetPrivate TargetType = "Private"
	TargetGuild   TargetType = "Guild"
)

func (t TargetType) IsGroup() bool {
	return t == TargetGroup
}

func (t TargetType) IsPrivate() bool {
	return t == TargetPrivate
}

func (t TargetType) IsGuild() bool {
	return t == TargetGuild
}

func (t TargetType) GetTargetType() TargetType {
	return t
}

type Target interface {
	GetTargetType() TargetType
	Equal(t Target) bool
	Hash() string
	Parse(hash string) bool
}

type PrivateTarget struct {
	TargetType
	Uin int64 `json:"uin"`
}

func (t *PrivateTarget) Hash() string {
	return fmt.Sprintf("%v_%v", TargetPrivate, t.Uin)
}

func (t *PrivateTarget) Equal(target2 Target) bool {
	if target2 == nil {
		return false
	}
	return t.Hash() == target2.Hash()
}

func (t *PrivateTarget) Parse(hash string) bool {
	if !strings.HasPrefix(hash, string(TargetPrivate)) {
		return false
	}
	spt := strings.Split(hash, "_")
	if len(spt) < 1 {
		return false
	}
	if x, err := strconv.ParseInt(spt[1], 10, 64); err != nil {
		return false
	} else {
		t.Uin = x
		return true
	}
}

func (t *PrivateTarget) TargetCode() int64 {
	return t.Uin
}

type GroupTarget struct {
	TargetType
	GroupCode int64 `json:"group_code"`
}

func (t *GroupTarget) TargetCode() int64 {
	return t.GroupCode
}

func (t *GroupTarget) Hash() string {
	return fmt.Sprintf("%v_%v", TargetGroup, t.GroupCode)
}

func (t *GroupTarget) Equal(target2 Target) bool {
	if target2 == nil {
		return false
	}
	return t.Hash() == target2.Hash()
}

func (t *GroupTarget) Parse(hash string) bool {
	if !strings.HasPrefix(hash, string(TargetGroup)) {
		return false
	}
	spt := strings.Split(hash, "_")
	if len(spt) <= 1 {
		return false
	}
	if x, err := strconv.ParseInt(spt[1], 10, 64); err != nil {
		return false
	} else {
		t.GroupCode = x
		return true
	}
}

type GuildTarget struct {
	TargetType
	GuildId   uint64 `json:"guild_id"`
	ChannelId uint64 `json:"channel_id"`
}

func (t *GuildTarget) Hash() string {
	return fmt.Sprintf("%v_%v_%v", TargetGuild, t.GuildId, t.ChannelId)
}

func (t *GuildTarget) Equal(target2 Target) bool {
	if target2 == nil {
		return false
	}
	return t.Hash() == target2.Hash()
}

func (t *GuildTarget) Parse(hash string) bool {
	if !strings.HasPrefix(hash, string(TargetGuild)) {
		return false
	}
	spt := strings.Split(hash, "_")
	if len(spt) <= 2 {
		return false
	}
	if x, err := strconv.ParseUint(spt[1], 10, 64); err != nil {
		return false
	} else {
		t.GuildId = x
	}
	if x, err := strconv.ParseUint(spt[2], 10, 64); err != nil {
		return false
	} else {
		t.ChannelId = x
		return true
	}
}

func NewGroupTarget(groupCode int64) *GroupTarget {
	return &GroupTarget{TargetType: TargetGroup, GroupCode: groupCode}
}

func NewPrivateTarget(uin int64) *PrivateTarget {
	return &PrivateTarget{TargetType: TargetPrivate, Uin: uin}
}

func NewGuildTarget(guildId uint64, channelId uint64) *GuildTarget {
	return &GuildTarget{
		TargetType: TargetGuild,
		GuildId:    guildId,
		ChannelId:  channelId,
	}
}

func ParseTargetHash(hash string) Target {
	var t Target
	if strings.HasPrefix(hash, string(TargetPrivate)) {
		t = NewPrivateTarget(0)
	} else if strings.HasPrefix(hash, string(TargetGroup)) {
		t = NewGroupTarget(0)
	} else if strings.HasPrefix(hash, string(TargetGuild)) {
		t = NewGuildTarget(0, 0)
	} else {
		return nil
	}
	if !t.Parse(hash) {
		return nil
	}
	return t
}

func AllTargetType() []TargetType {
	return []TargetType{
		TargetGroup, TargetPrivate, TargetGuild,
	}
}
