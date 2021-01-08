package permission

type Type int64

const (
	Unknown = 0

	Admin Type = 1 << iota
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
