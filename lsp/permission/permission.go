package permission

type RoleType int64

const (
	Unknown RoleType = 0

	Admin RoleType = 1 << iota
)

type Level int64

const (
	Command Level = 1 << iota
	Group
	Role

	Empty Level = 0
)

func (t RoleType) String() string {
	switch t {
	case Admin:
		return "Admin"
	default:
		return ""
	}
}

func FromString(s string) RoleType {
	switch s {
	case "Admin":
		return Admin
	default:
		return Unknown
	}
}

type RequireOption interface {
	Type() Level
}

type roleRequireOption struct {
	uin int64
}

func (r *roleRequireOption) Type() Level {
	return Role
}

func RoleRequireOption(uin int64) RequireOption {
	return &roleRequireOption{uin}
}

type groupRequireOption struct {
	groupCode int64
	uin       int64
}

func (g *groupRequireOption) Type() Level {
	return Group
}

func GroupRequireOption(groupCode int64, uin int64) RequireOption {
	return &groupRequireOption{
		groupCode: groupCode,
		uin:       uin,
	}
}

type groupCommandRequireOption struct {
	groupCode int64
	uin       int64
	command   string
}

func (g *groupCommandRequireOption) Type() Level {
	return Command
}

func GroupCommandRequireOption(groupCode int64, uin int64, command string) RequireOption {
	return &groupCommandRequireOption{
		groupCode: groupCode,
		uin:       uin,
		command:   command,
	}
}
