package concern

type Type int64

const (
	Live Type = 1 << iota
	News
)

type Notify interface {
	Type() Type
}
