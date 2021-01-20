package permission

import (
	"errors"
	"github.com/Logiase/MiraiGo-Template/bot"
	"github.com/Logiase/MiraiGo-Template/utils"
	"github.com/Mrs4s/MiraiGo/client"
	localdb "github.com/Sora233/Sora233-MiraiGo/lsp/buntdb"
	"github.com/tidwall/buntdb"
)

var logger = utils.GetModuleLogger("permission")

type StateManager struct {
	*KeySet
}

func (c *StateManager) CheckRole(caller int64, role RoleType) bool {
	if role.String() == "" {
		return false
	}
	var (
		result bool
	)
	db, err := localdb.GetClient()
	if err != nil {
		logger.Errorf("get db failed %v", err)
		return false
	}
	err = db.View(func(tx *buntdb.Tx) error {
		key := c.PermissionKey(caller, role.String())
		_, err := tx.Get(key)
		if err == nil {
			result = true
			return nil
		} else if err == buntdb.ErrNotFound {
			return nil
		} else {
			return err
		}
	})
	if err != nil {
		logger.WithField("caller", caller).
			WithField("role", role.String()).
			Errorf("check role err %v", err)
	}
	return result
}
func (c *StateManager) CheckGroupRole(groupCode int64, caller int64, role RoleType) bool {
	if role.String() == "" {
		return false
	}
	var (
		result bool
	)
	db, err := localdb.GetClient()
	if err != nil {
		logger.Errorf("get db failed %v", err)
		return false
	}
	err = db.View(func(tx *buntdb.Tx) error {
		key := c.GroupPermissionKey(groupCode, caller, role.String())
		_, err := tx.Get(key)
		if err == nil {
			result = true
			return nil
		} else if err == buntdb.ErrNotFound {
			return nil
		} else {
			return err
		}
	})
	if err != nil {
		logger.WithField("group_code", groupCode).
			WithField("caller", caller).
			WithField("role", role.String()).
			Errorf("check group role err %v", err)
	}
	return result
}

func (c *StateManager) EnableGroupCommand(groupCode int64, command string) error {
	db, err := localdb.GetClient()
	if err != nil {
		return err
	}
	return db.Update(func(tx *buntdb.Tx) error {
		key := c.GroupEnabledKey(groupCode, command)
		prev, replaced, err := tx.Set(key, Enable, nil)
		if err != nil {
			return err
		}
		if replaced && prev == Enable {
			return ErrPermisionExist
		}
		return nil
	})
}

func (c *StateManager) DisableGroupCommand(groupCode int64, command string) error {
	db, err := localdb.GetClient()
	if err != nil {
		return err
	}
	return db.Update(func(tx *buntdb.Tx) error {
		key := c.GroupEnabledKey(groupCode, command)
		prev, replaced, err := tx.Set(key, Disable, nil)
		if err != nil {
			return err
		}
		if replaced && prev == Disable {
			return ErrPermisionExist
		}
		return nil
	})
}

// explicit enabled, must exist
func (c *StateManager) CheckGroupCommandEnabled(groupCode int64, command string) bool {
	return c.CheckGroupCommandFunc(groupCode, command, func(val string, exist bool) bool {
		return exist && val == Enable
	})
}

// explicit disabled, must exist
func (c *StateManager) CheckGroupCommandDisabled(groupCode int64, command string) bool {
	return c.CheckGroupCommandFunc(groupCode, command, func(val string, exist bool) bool {
		return exist && val == Disable
	})
}

func (c *StateManager) CheckGroupCommandFunc(groupCode int64, command string, f func(val string, exist bool) bool) bool {
	var result bool
	db, err := localdb.GetClient()
	if err != nil {
		logger.Errorf("get db failed %v", err)
		return false
	}
	err = db.View(func(tx *buntdb.Tx) error {
		key := c.GroupEnabledKey(groupCode, command)
		val, err := tx.Get(key)
		if err != nil && err != buntdb.ErrNotFound {
			return err
		}
		result = f(val, err == nil)
		return nil
	})
	if err != nil {
		logger.WithField("group_code", groupCode).
			WithField("command", command).
			Errorf("check group enable err %v", err)
	}
	return result
}

func (c *StateManager) CheckGroupAdministrator(groupCode int64, caller int64) bool {
	b := bot.Instance
	groupInfo := b.FindGroup(groupCode)
	if groupInfo == nil {
		logger.Errorf("nil group info")
		return false
	}
	groupMemberInfo := groupInfo.FindMember(caller)
	if groupMemberInfo == nil {
		logger.Errorf("nil member info")
		return false
	}
	logger.WithField("uin", caller).
		WithField("group_code", groupCode).
		WithField("permission", groupMemberInfo.Permission).
		Debug("debug member permission")
	return groupMemberInfo.Permission == client.Administrator || groupMemberInfo.Permission == client.Owner
}

func (c *StateManager) CheckGroupCommandPermission(groupCode int64, caller int64, command string) bool {
	if c.CheckRole(caller, Admin) {
		return true
	}
	var (
		result bool
	)
	db, err := localdb.GetClient()
	if err != nil {
		logger.Errorf("get db failed %v", err)
		return false
	}
	err = db.View(func(tx *buntdb.Tx) error {
		key := c.PermissionKey(groupCode, caller, command)
		_, err := tx.Get(key)
		if err == buntdb.ErrNotFound {
			return nil
		} else if err != nil {
			return err
		}
		result = true
		return nil
	})
	if err != nil {
		logger.WithField("group_code", groupCode).
			WithField("caller", caller).
			WithField("command", command).
			Errorf("check group command err %v", err)
	}
	return result
}

func (c *StateManager) GrantRole(target int64, role RoleType) error {
	if role.String() == "" {
		return errors.New("error role")
	}
	db, err := localdb.GetClient()
	if err != nil {
		return err
	}
	err = db.Update(func(tx *buntdb.Tx) error {
		key := c.PermissionKey(target, role.String())
		_, err := tx.Get(key)
		if err == buntdb.ErrNotFound {
			tx.Set(key, "", nil)
			return nil
		} else if err == nil {
			return ErrPermisionExist
		} else {
			return err
		}
	})
	return err
}

func (c *StateManager) GrantGroupRole(groupCode int64, target int64, role RoleType) error {
	if role.String() == "" {
		return errors.New("error role")
	}
	db, err := localdb.GetClient()
	if err != nil {
		return err
	}
	err = db.Update(func(tx *buntdb.Tx) error {
		key := c.GroupPermissionKey(groupCode, target, role.String())
		_, err := tx.Get(key)
		if err == buntdb.ErrNotFound {
			tx.Set(key, "", nil)
			return nil
		} else if err == nil {
			return ErrPermisionExist
		} else {
			return err
		}
	})
	return err
}

func (c *StateManager) GrantPermission(groupCode int64, target int64, command string) error {
	db, err := localdb.GetClient()
	if err != nil {
		return err
	}
	err = db.Update(func(tx *buntdb.Tx) error {
		key := c.PermissionKey(groupCode, target, command)
		_, err := tx.Get(key)
		if err == buntdb.ErrNotFound {
			tx.Set(key, "", nil)
			return nil
		} else if err == nil {
			return ErrPermisionExist
		} else {
			return err
		}
	})
	return err
}

func (c *StateManager) RequireAny(option ...RequireOption) bool {
	var ok bool
	for _, iopt := range option {
		switch iopt.Type() {
		case Role:
			switch opt := iopt.(type) {
			case *adminRoleRequireOption:
				ok = ok || c.requireAdminRole(opt)
			case *groupAdminRoleRequireOption:
				ok = ok || c.requireGroupAdminRole(opt)
			}
		case Group:
			opt := iopt.(*qqAdminRequireOption)
			ok = ok || c.requireQQGroupAdmin(opt)
		case Command:
			opt := iopt.(*groupCommandRequireOption)
			ok = ok || c.requireGroupCommand(opt)
		}
		if ok {
			return true
		}
	}
	return false
}

func (c *StateManager) RemoveAll(groupCode int64) error {
	db, err := localdb.GetClient()
	if err != nil {
		return err
	}
	var deleteKey []string
	_ = db.View(func(tx *buntdb.Tx) error {
		tx.Ascend(c.GroupPermissionKey(groupCode), func(key, value string) bool {
			deleteKey = append(deleteKey, key)
			return true
		})
		return nil
	})
	db.Update(func(tx *buntdb.Tx) error {
		for _, k := range deleteKey {
			tx.Delete(k)
		}
		tx.DropIndex(c.GroupPermissionKey(groupCode))
		return nil
	})
	return nil
}

func (c *StateManager) requireAdminRole(opt *adminRoleRequireOption) bool {
	uin := opt.uin
	if c.CheckRole(uin, Admin) {
		logger.WithField("type", "Role").WithField("uin", uin).
			WithField("result", true).
			Debug("debug permission")
		return true
	}
	return false
}

func (c *StateManager) requireGroupAdminRole(opt *groupAdminRoleRequireOption) bool {
	uin := opt.uin
	groupCode := opt.groupCode
	if c.CheckGroupRole(groupCode, uin, GroupAdmin) {
		logger.WithField("type", "GroupRole").
			WithField("group_code", groupCode).
			WithField("uin", uin).
			WithField("result", true).
			Debug("debug permission")
		return true
	}
	return false
}

func (c *StateManager) requireQQGroupAdmin(opt *qqAdminRequireOption) bool {
	uin := opt.uin
	groupCode := opt.groupCode
	if c.CheckGroupAdministrator(groupCode, uin) {
		logger.WithField("type", "Group").WithField("uin", uin).
			WithField("group_code", groupCode).
			WithField("result", true).
			Debug("debug permission")
		return true
	}
	return false
}

func (c *StateManager) requireGroupCommand(opt *groupCommandRequireOption) bool {
	uin := opt.uin
	groupCode := opt.groupCode
	cmd := opt.command
	if c.CheckGroupCommandPermission(groupCode, uin, cmd) {
		logger.WithField("type", "command").WithField("uin", uin).
			WithField("command", cmd).
			WithField("group_code", groupCode).
			WithField("result", true).
			Debug("debug permission")
		return true
	}
	return false
}

func (c *StateManager) FreshIndex() {
	db, _ := localdb.GetClient()
	db.CreateIndex(c.PermissionKey(), c.PermissionKey("*"), buntdb.IndexString)
	for _, group := range bot.Instance.GroupList {
		db.CreateIndex(c.GroupPermissionKey(group.Code), c.GroupPermissionKey(group.Code, "*"), buntdb.IndexString)
	}
}

func NewStateManager() *StateManager {
	sm := &StateManager{
		KeySet: NewKeySet(),
	}
	return sm
}
