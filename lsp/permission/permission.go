package permission

import (
	localutils "github.com/Sora233/DDBOT/utils"
	"github.com/sirupsen/logrus"
)

type RoleType int64

const (
	Unknown RoleType = 0

	Admin RoleType = 1 << iota
	GroupAdmin
	User
)

const Enable = "enable"
const Disable = "disable"

func (t RoleType) String() string {
	switch t {
	case Admin:
		return "Admin"
	case GroupAdmin:
		return "GroupAdmin"
	case User:
		return "User"
	default:
		return ""
	}
}

func NewRoleFromString(s string) RoleType {
	switch s {
	case "Admin":
		return Admin
	case "GroupAdmin":
		return GroupAdmin
	case "User":
		return User
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
		logger.WithFields(logrus.Fields{
			"type": "AdminRole",
			"uin":  r.uin,
		}).Debug("adminRole permission pass")
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
		logger.WithFields(localutils.GroupLogFields(groupCode)).
			WithFields(logrus.Fields{
				"type": "GroupAdminRole",
				"uin":  uin,
			}).Debug("groupAdminRole permission pass")
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
		logger.WithFields(localutils.GroupLogFields(groupCode)).
			WithFields(logrus.Fields{
				"type": "QQGroupAdmin",
				"uin":  uin,
			}).Debug("qqAdmin permission pass")
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
		logger.WithFields(localutils.GroupLogFields(groupCode)).
			WithFields(logrus.Fields{
				"type":    "command",
				"uin":     uin,
				"command": cmd,
			}).Debug("groupCommand permission pass")
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
