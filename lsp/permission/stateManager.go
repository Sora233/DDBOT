package permission

import (
	"errors"
	"github.com/Logiase/MiraiGo-Template/bot"
	"github.com/Logiase/MiraiGo-Template/utils"
	"github.com/Mrs4s/MiraiGo/client"
	localdb "github.com/Sora233/DDBOT/lsp/buntdb"
	"github.com/tidwall/buntdb"
	"time"
)

var logger = utils.GetModuleLogger("permission")

type StateManager struct {
	*localdb.ShortCut
	*KeySet
}

type commandOption struct {
	duration time.Duration
}

type CommandOption func(c *commandOption)

func ExpireOption(d time.Duration) CommandOption {
	return func(c *commandOption) {
		c.duration = d
	}
}

// CheckBlockList return true if blocked
func (c *StateManager) CheckBlockList(caller int64) bool {
	var result bool
	err := c.RTxCover(func(tx *buntdb.Tx) error {
		key := c.BlockListKey(caller)
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
		logger.WithField("caller", caller).Errorf("check block list err %v", err)
		result = false
	}
	return result
}

func (c *StateManager) AddBlockList(caller int64, d time.Duration) error {
	return c.RWTxCover(func(tx *buntdb.Tx) error {
		key := c.BlockListKey(caller)
		_, err := tx.Get(key)
		if err == nil {
			return localdb.ErrKeyExist
		} else if err != buntdb.ErrNotFound {
			return err
		}
		_, _, err = tx.Set(key, "", localdb.ExpireOption(d))
		return err
	})
}

func (c *StateManager) DeleteBlockList(caller int64) error {
	return c.RWTxCover(func(tx *buntdb.Tx) error {
		key := c.BlockListKey(caller)
		_, err := tx.Get(key)
		if err != nil {
			return err
		}
		_, err = tx.Delete(key)
		return err
	})
}

func (c *StateManager) CheckRole(caller int64, role RoleType) bool {
	if role.String() == "" {
		return false
	}
	var result bool
	err := c.RTxCover(func(tx *buntdb.Tx) error {
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
		result = false
	}
	return result
}
func (c *StateManager) CheckGroupRole(groupCode int64, caller int64, role RoleType) bool {
	if role.String() == "" {
		return false
	}
	var result bool
	err := c.RTxCover(func(tx *buntdb.Tx) error {
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
		result = false
	}
	return result
}

func (c *StateManager) EnableGroupCommand(groupCode int64, command string, opts ...CommandOption) error {
	opt := new(commandOption)
	for _, o := range opts {
		o(opt)
	}
	return c.RWTxCover(func(tx *buntdb.Tx) error {
		key := c.GroupEnabledKey(groupCode, command)
		prev, replaced, err := tx.Set(key, Enable, localdb.ExpireOption(opt.duration))
		if err != nil {
			return err
		}
		if replaced && prev == Enable {
			return ErrPermissionExist
		}
		return nil
	})
}

func (c *StateManager) DisableGroupCommand(groupCode int64, command string, opts ...CommandOption) error {
	opt := new(commandOption)
	for _, o := range opts {
		o(opt)
	}
	return c.RWTxCover(func(tx *buntdb.Tx) error {
		key := c.GroupEnabledKey(groupCode, command)
		prev, replaced, err := tx.Set(key, Disable, localdb.ExpireOption(opt.duration))
		if err != nil {
			return err
		}
		if replaced && prev == Disable {
			return ErrPermissionExist
		}
		return nil
	})
}

// CheckGroupCommandEnabled check explicit enabled, must exist
func (c *StateManager) CheckGroupCommandEnabled(groupCode int64, command string) bool {
	return c.CheckGroupCommandFunc(groupCode, command, func(val string, exist bool) bool {
		return exist && val == Enable
	})
}

// CheckGroupCommandDisabled check explicit disabled, must exist
func (c *StateManager) CheckGroupCommandDisabled(groupCode int64, command string) bool {
	return c.CheckGroupCommandFunc(groupCode, command, func(val string, exist bool) bool {
		return exist && val == Disable
	})
}

func (c *StateManager) CheckGroupCommandFunc(groupCode int64, command string, f func(val string, exist bool) bool) bool {
	var result bool
	err := c.RTxCover(func(tx *buntdb.Tx) error {
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
		result = false
	}
	return result
}

func (c *StateManager) CheckGroupAdministrator(groupCode int64, caller int64) bool {
	b := bot.Instance
	if b == nil {
		logger.Errorf("bot not init")
		return false
	}
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
	var result bool
	err := c.RTxCover(func(tx *buntdb.Tx) error {
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
		result = false
	}
	return result
}

func (c *StateManager) GrantRole(target int64, role RoleType) error {
	if role.String() == "" {
		return errors.New("error role")
	}
	return c.RWTxCover(func(tx *buntdb.Tx) error {
		key := c.PermissionKey(target, role.String())
		_, err := tx.Get(key)
		if err == buntdb.ErrNotFound {
			tx.Set(key, "", nil)
			return nil
		} else if err == nil {
			return ErrPermissionExist
		} else {
			return err
		}
	})
}

func (c *StateManager) UngrantRole(target int64, role RoleType) error {
	if role.String() == "" {
		return errors.New("error role")
	}
	return c.RWTxCover(func(tx *buntdb.Tx) error {
		key := c.PermissionKey(target, role.String())
		_, err := tx.Get(key)
		if err == buntdb.ErrNotFound {
			return ErrPermissionNotExist
		} else if err == nil {
			tx.Delete(key)
			return nil
		} else {
			return err
		}
	})
}

func (c *StateManager) GrantGroupRole(groupCode int64, target int64, role RoleType) error {
	if role.String() == "" {
		return errors.New("error role")
	}
	return c.RWTxCover(func(tx *buntdb.Tx) error {
		key := c.GroupPermissionKey(groupCode, target, role.String())
		_, err := tx.Get(key)
		if err == buntdb.ErrNotFound {
			_, _, err := tx.Set(key, "", nil)
			return err
		} else if err == nil {
			return ErrPermissionExist
		} else {
			return err
		}
	})
}

func (c *StateManager) UngrantGroupRole(groupCode int64, target int64, role RoleType) error {
	if role.String() == "" {
		return errors.New("error role")
	}
	return c.RWTxCover(func(tx *buntdb.Tx) error {
		key := c.GroupPermissionKey(groupCode, target, role.String())
		_, err := tx.Get(key)
		if err == buntdb.ErrNotFound {
			return ErrPermissionNotExist
		} else if err == nil {
			_, err := tx.Delete(key)
			return err
		} else {
			return err
		}
	})
}

func (c *StateManager) GrantPermission(groupCode int64, target int64, command string) error {
	return c.RWTxCover(func(tx *buntdb.Tx) error {
		key := c.PermissionKey(groupCode, target, command)
		_, err := tx.Get(key)
		if err == buntdb.ErrNotFound {
			tx.Set(key, "", nil)
			return nil
		} else if err == nil {
			return ErrPermissionExist
		} else {
			return err
		}
	})
}

func (c *StateManager) UngrantPermission(groupCode int64, target int64, command string) error {
	return c.RWTxCover(func(tx *buntdb.Tx) error {
		key := c.PermissionKey(groupCode, target, command)
		_, err := tx.Get(key)
		if err == buntdb.ErrNotFound {
			return ErrPermissionNotExist
		} else if err == nil {
			tx.Delete(key)
			return nil
		} else {
			return err
		}
	})
}

func (c *StateManager) RequireAny(option ...RequireOption) bool {
	for _, opt := range option {
		if opt.Validate(c) {
			return true
		}
	}
	return false
}

func (c *StateManager) RemoveAllByGroup(groupCode int64) error {
	var deleteKey []string
	err := c.RTxCover(func(tx *buntdb.Tx) error {
		tx.Ascend(c.GroupPermissionKey(groupCode), func(key, value string) bool {
			deleteKey = append(deleteKey, key)
			return true
		})
		return nil
	})
	if err != nil {
		return err
	}
	return c.RWTxCover(func(tx *buntdb.Tx) error {
		for _, k := range deleteKey {
			tx.Delete(k)
		}
		tx.DropIndex(c.GroupPermissionKey(groupCode))
		return nil
	})
}

func (c *StateManager) FreshIndex() {
	db := localdb.MustGetClient()
	db.CreateIndex(c.PermissionKey(), c.PermissionKey("*"), buntdb.IndexString)
	if bot.Instance != nil {
		for _, group := range bot.Instance.GroupList {
			db.CreateIndex(c.GroupPermissionKey(group.Code), c.GroupPermissionKey(group.Code, "*"), buntdb.IndexString)
		}
	}
}

func NewStateManager() *StateManager {
	sm := &StateManager{
		KeySet: NewKeySet(),
	}
	return sm
}
