package permission

import (
	"github.com/Sora233/DDBOT/lsp/mmsg/mt"
	localutils "github.com/Sora233/DDBOT/utils"
	"github.com/sirupsen/logrus"
)

type RoleType int64

const (
	Unknown RoleType = 0

	Admin RoleType = 1 << iota
	TargetAdmin
	User
)

const Enable = "enable"
const Disable = "disable"

func (t RoleType) String() string {
	switch t {
	case Admin:
		return "Admin"
	case TargetAdmin:
		return "TargetAdmin"
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
	case "TargetAdmin":
		return TargetAdmin
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

type targetAdminRoleRequireOption struct {
	target mt.Target
	uin    int64
}

func (g *targetAdminRoleRequireOption) Validate(s *StateManager) bool {
	uin := g.uin
	if s.CheckTargetRole(g.target, uin, TargetAdmin) {
		logger.WithFields(localutils.TargetFields(g.target)).
			WithFields(logrus.Fields{
				"type": "TargetAdminRole",
				"uin":  uin,
			}).Debug("targetAdminRole permission pass")
		return true
	}
	return false
}

func TargetAdminRoleRequireOption(target mt.Target, uin int64) RequireOption {
	return &targetAdminRoleRequireOption{target: target, uin: uin}
}

type qqAdminRequireOption struct {
	target mt.Target
	uin    int64
}

func (g *qqAdminRequireOption) Validate(s *StateManager) bool {
	uin := g.uin
	switch g.target.GetTargetType() {
	case mt.TargetPrivate:
		return true
	case mt.TargetGulid:
		// TODO support
		return false
	}
	if s.CheckGroupAdministrator(g.target, uin) {
		logger.WithFields(localutils.TargetFields(g.target)).
			WithFields(logrus.Fields{
				"type": "QQGroupAdmin",
				"uin":  uin,
			}).Debug("qqAdmin permission pass")
		return true
	}
	return false
}

func QQAdminRequireOption(target mt.Target, uin int64) RequireOption {
	return &qqAdminRequireOption{
		target: target,
		uin:    uin,
	}
}

type targetCommandRequireOption struct {
	target  mt.Target
	uin     int64
	command string
}

func (g *targetCommandRequireOption) Validate(s *StateManager) bool {
	uin := g.uin
	cmd := g.command
	if s.CheckTargetCommandPermission(g.target, uin, cmd) {
		logger.WithFields(localutils.TargetFields(g.target)).
			WithFields(logrus.Fields{
				"type":    "targetCommand",
				"uin":     uin,
				"command": cmd,
			}).Debug("groupCommand permission pass")
		return true
	}
	return false
}

func TargetCommandRequireOption(target mt.Target, uin int64, command string) RequireOption {
	return &targetCommandRequireOption{
		target:  target,
		uin:     uin,
		command: command,
	}
}
