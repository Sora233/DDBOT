package permission

type RoleType int64

const (
	Unknown RoleType = 0

	Admin RoleType = 1 << iota
	GroupAdmin
)

type Level int64

const (
	Command Level = 1 << iota
	Group
	Role

	Empty Level = 0
)

const Enable = "enable"
const Disable = "disable"

func (t RoleType) String() string {
	switch t {
	case Admin:
		return "Admin"
	case GroupAdmin:
		return "GroupAdmin"
	default:
		return ""
	}
}

func FromString(s string) RoleType {
	switch s {
	case "Admin":
		return Admin
	case "GroupAdmin":
		return GroupAdmin
	default:
		return Unknown
	}
}

type RequireOption interface {
	Type() Level
}

type adminRoleRequireOption struct {
	uin int64
}

func (r *adminRoleRequireOption) Type() Level {
	return Role
}

func AdminRoleRequireOption(uin int64) RequireOption {
	return &adminRoleRequireOption{uin}
}

type groupAdminRoleRequireOption struct {
	groupCode int64
	uin       int64
}

func (g *groupAdminRoleRequireOption) Type() Level {
	return Role
}

func GroupAdminRoleRequireOption(groupCode int64, uin int64) RequireOption {
	return &groupAdminRoleRequireOption{groupCode: groupCode, uin: uin}
}

type qqAdminRequireOption struct {
	groupCode int64
	uin       int64
}

func (g *qqAdminRequireOption) Type() Level {
	return Group
}

func QQAdminRequireOption(groupCode int64, uin int64) RequireOption {
	return &qqAdminRequireOption{
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
