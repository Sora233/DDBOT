package concern

import (
	"github.com/Mrs4s/MiraiGo/message"
	"strconv"
	"strings"
)

type Type int64

const Empty Type = 0

const (
	BibiliLive Type = 1 << iota
	BilibiliNews
	DouyuLive
	YoutubeLive
	YoutubeVideo
	HuyaLive
)

var all = [...]Type{BibiliLive, BilibiliNews, DouyuLive, YoutubeLive, YoutubeVideo, HuyaLive}

type Notify interface {
	Type() Type
	GetGroupCode() int64
	GetUid() interface{}
	ToMessage() []message.IMessageElement
}

func (t Type) String() string {
	return strconv.FormatInt(int64(t), 10)
}

func (t Type) ContainAll(o Type) bool {
	if t.Empty() && o.Empty() {
		return false
	}
	return t&o == o
}

func (t Type) ContainAny(o Type) bool {
	if t.Empty() && o.Empty() {
		return false
	}
	for _, c := range all {
		if t.ContainAll(c) && o.ContainAll(c) {
			return true
		}
	}
	return false
}

// Split return a Type unit slice from a given Type
func (t Type) Split() []Type {
	var result []Type
	for _, a := range all {
		if t.ContainAny(a) {
			result = append(result, a)
		}
	}
	return result
}

func (t Type) Remove(o Type) Type {
	newT := t
	for _, c := range all {
		if t.ContainAll(c) && o.ContainAll(c) {
			newT ^= c
		}
	}
	return newT
}

func (t Type) Add(o Type) Type {
	newT := t
	for _, c := range all {
		if !t.ContainAll(c) && o.ContainAll(c) {
			newT ^= c
		}
	}
	return newT
}

func (t Type) Empty() bool {
	return t == 0
}

func FromString(s string) Type {
	t, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return Type(0)
	}
	return Type(t)
}

func (t Type) Description() string {
	switch t {
	case BibiliLive, HuyaLive, DouyuLive, YoutubeLive:
		return "live"
	case BilibiliNews, YoutubeVideo:
		return "news"
	}

	sb := strings.Builder{}
	var first = true
	for _, o := range all {
		if t.ContainAny(o) {
			if first {
				first = false
			} else {
				sb.WriteString("/")
			}
			sb.WriteString(o.Description())
		}
	}
	return sb.String()
}
