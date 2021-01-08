package permission

type Type int64

const (
	Unknown Type = 0

	Admin Type = 1 << iota
)

type Level int64

const (
	Command Level = 1 << iota
	Group
	Role

	Empty Level = 0
)

func (t Type) String() string {
	switch t {
	case Admin:
		return "Admin"
	default:
		return ""
	}
}

func FromString(s string) Type {
	switch s {
	case "Admin":
		return Admin
	default:
		return Unknown
	}
}
