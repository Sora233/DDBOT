package permission

import (
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/constraints"

	localutils "github.com/Sora233/DDBOT/v2/utils"
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

type RequireOption[UT, GT constraints.Integer] interface {
	Validate(s *StateManager[UT, GT]) bool
}

type adminRoleRequireOption[UT, _ constraints.Integer] struct {
	uin UT
}

func (r *adminRoleRequireOption[UT, GT]) Validate(s *StateManager[UT, GT]) bool {
	if s.CheckRole(r.uin, Admin) {
		logger.WithFields(logrus.Fields{
			"type": "AdminRole",
			"uin":  r.uin,
		}).Debug("adminRole permission pass")
		return true
	}
	return false
}

func AdminRoleRequireOption[UT constraints.Integer](uin UT) RequireOption[UT, uint32] {
	return &adminRoleRequireOption[UT, uint32]{uin}
}

type groupAdminRoleRequireOption[UT, GT constraints.Integer] struct {
	groupCode GT
	uin       UT
}

func (g *groupAdminRoleRequireOption[UT, GT]) Validate(s *StateManager[UT, GT]) bool {
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

func GroupAdminRoleRequireOption[UT, GT constraints.Integer](groupCode GT, uin UT) RequireOption[UT, GT] {
	return &groupAdminRoleRequireOption[UT, GT]{groupCode: groupCode, uin: uin}
}

type qqAdminRequireOption[UT, GT constraints.Integer] struct {
	groupCode GT
	uin       UT
}

func (g *qqAdminRequireOption[UT, GT]) Validate(s *StateManager[UT, GT]) bool {
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

func QQAdminRequireOption[UT, GT constraints.Integer](groupCode GT, uin UT) RequireOption[UT, GT] {
	return &qqAdminRequireOption[UT, GT]{
		groupCode: groupCode,
		uin:       uin,
	}
}

type groupCommandRequireOption[UT, GT constraints.Integer] struct {
	groupCode GT
	uin       UT
	command   string
}

func (g *groupCommandRequireOption[UT, GT]) Validate(s *StateManager[UT, GT]) bool {
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

func GroupCommandRequireOption[UT, GT constraints.Integer](groupCode GT, uin UT, command string) RequireOption[UT, GT] {
	return &groupCommandRequireOption[UT, GT]{
		groupCode: groupCode,
		uin:       uin,
		command:   command,
	}
}
