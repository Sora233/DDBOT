package concern

type Type int64

const (
	BibiliLive Type = 1 << iota
	BilibiliNews
)

type Notify interface {
	Type() Type
}
