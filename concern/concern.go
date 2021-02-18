package concern

import "strconv"

type Type int64

const Empty Type = 0

const (
	BibiliLive Type = 1 << iota
	BilibiliNews
	DouyuLive
	YoutubeLive
	YoutubeVideo
)

var all = [...]Type{BibiliLive, BilibiliNews, DouyuLive, YoutubeLive, YoutubeVideo}

type Notify interface {
	Type() Type
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
