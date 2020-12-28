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

func (t Type) ToString() string {
	return strconv.FormatInt(int64(t), 10)
}

func FromString(s string) Type {
	t, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return Type(0)
	}
	return Type(t)
}
