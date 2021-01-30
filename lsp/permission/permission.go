package permission

type RoleType int64

const (
	Unknown RoleType = 0

	Admin RoleType = 1 << iota
	GroupAdmin
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
	Validate(s *StateManager) bool
}

type adminRoleRequireOption struct {
	uin int64
}

func (r *adminRoleRequireOption) Validate(s *StateManager) bool {
	if s.CheckRole(r.uin, Admin) {
		logger.WithField("type", "AdminRole").WithField("uin", r.uin).
			WithField("result", true).
			Debug("debug permission")
		return true
	}
	return false
}

func AdminRoleRequireOption(uin int64) RequireOption {
	return &adminRoleRequireOption{uin}
}

type groupAdminRoleRequireOption struct {
	groupCode int64
	uin       int64
}

func (g *groupAdminRoleRequireOption) Validate(s *StateManager) bool {
	uin := g.uin
	groupCode := g.groupCode
	if s.CheckGroupRole(groupCode, uin, GroupAdmin) {
		logger.WithField("type", "GroupAdminRole").
			WithField("group_code", groupCode).
			WithField("uin", uin).
			WithField("result", true).
			Debug("debug permission")
		return true
	}
	return false
}

func GroupAdminRoleRequireOption(groupCode int64, uin int64) RequireOption {
	return &groupAdminRoleRequireOption{groupCode: groupCode, uin: uin}
}

type qqAdminRequireOption struct {
	groupCode int64
	uin       int64
}

func (g *qqAdminRequireOption) Validate(s *StateManager) bool {
	uin := g.uin
	groupCode := g.groupCode
	if s.CheckGroupAdministrator(groupCode, uin) {
		logger.WithField("type", "QQGroupAdmin").WithField("uin", uin).
			WithField("group_code", groupCode).
			WithField("result", true).
			Debug("debug permission")
		return true
	}
	return false
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

func (g *groupCommandRequireOption) Validate(s *StateManager) bool {
	uin := g.uin
	groupCode := g.groupCode
	cmd := g.command
	if s.CheckGroupCommandPermission(groupCode, uin, cmd) {
		logger.WithField("type", "command").WithField("uin", uin).
			WithField("command", cmd).
			WithField("group_code", groupCode).
			WithField("result", true).
			Debug("debug permission")
		return true
	}
	return false
}

func GroupCommandRequireOption(groupCode int64, uin int64, command string) RequireOption {
	return &groupCommandRequireOption{
		groupCode: groupCode,
		uin:       uin,
		command:   command,
	}
}
