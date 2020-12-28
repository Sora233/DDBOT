package concern

import "strconv"

type Type int64

const (
	BibiliLive Type = 1 << iota
	BilibiliNews
)

type Notify interface {
	Type() Type
}

func (t Type) String() string {
	return strconv.FormatInt(int64(t), 10)
}

func (t Type) Contain(o Type) bool {
	return t&o == o
}
func (t Type) Remove(o Type) Type {
	newT := t
	for _, c := range []Type{BibiliLive, BilibiliNews} {
		if t.Contain(c) && o.Contain(c) {
			newT ^= c
		}
	}
	return newT
}

func FromString(s string) Type {
	t, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return Type(0)
	}
	return Type(t)
}
