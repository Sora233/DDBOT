package permission

import (
	"errors"
	"github.com/Logiase/MiraiGo-Template/bot"
	"github.com/Logiase/MiraiGo-Template/utils"
	"github.com/Mrs4s/MiraiGo/client"
	localdb "github.com/Sora233/DDBOT/lsp/buntdb"
	localutils "github.com/Sora233/DDBOT/utils"
	"github.com/sirupsen/logrus"
	"github.com/tidwall/buntdb"
	"strconv"
	"strings"
	"time"
)

var logger = utils.GetModuleLogger("permission")

type StateManager struct {
	*localdb.ShortCut
	*KeySet
}

// CheckBlockList return true if blocked
func (c *StateManager) CheckBlockList(caller int64) bool {
	var result bool
	err := c.RCoverTx(func(tx *buntdb.Tx) error {
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
	return c.RWCoverTx(func(tx *buntdb.Tx) error {
		key := c.BlockListKey(caller)
		_, replaced, err := tx.Set(key, "", localdb.ExpireOption(d))
		if err != nil {
			return err
		}
		if replaced {
			return localdb.ErrKeyExist
		}
		return nil
	})
}

func (c *StateManager) DeleteBlockList(caller int64) error {
	return c.RWCoverTx(func(tx *buntdb.Tx) error {
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
	err := c.RCoverTx(func(tx *buntdb.Tx) error {
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
	err := c.RCoverTx(func(tx *buntdb.Tx) error {
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
		logger.WithFields(localutils.GroupLogFields(groupCode)).
			WithField("caller", caller).
			WithField("role", role.String()).
			Errorf("check group role err %v", err)
		result = false
	}
	return result
}

func (c *StateManager) EnableGroupCommand(groupCode int64, command string) error {
	if c.CheckGlobalCommandDisabled(command) {
		return ErrGlobalDisabled
	}
	return c.operatorEnableKey(c.GroupEnabledKey(groupCode, command), Enable)
}

func (c *StateManager) DisableGroupCommand(groupCode int64, command string) error {
	if c.CheckGlobalCommandDisabled(command) {
		return ErrGlobalDisabled
	}
	return c.operatorEnableKey(c.GroupEnabledKey(groupCode, command), Disable)
}

func (c *StateManager) GlobalEnableGroupCommand(command string) error {
	return c.operatorEnableKey(c.GlobalEnabledKey(command), Enable)
}

func (c *StateManager) GlobalDisableGroupCommand(command string) error {
	return c.operatorEnableKey(c.GlobalEnabledKey(command), Disable)
}

func (c *StateManager) operatorEnableKey(key string, status string) error {
	return c.RWCoverTx(func(tx *buntdb.Tx) error {
		prev, replaced, err := tx.Set(key, status, nil)
		if err != nil {
			return err
		}
		if replaced && prev == status {
			return ErrPermissionExist
		}
		return nil
	})
}

// CheckGroupCommandEnabled check global first, check explicit enabled, must exist
func (c *StateManager) CheckGroupCommandEnabled(groupCode int64, command string) bool {
	var result bool
	_ = c.RCover(func() error {
		result = !c.CheckGlobalCommandDisabled(command) && c.CheckGroupCommandFunc(groupCode, command, func(val string, exist bool) bool {
			return exist && val == Enable
		})
		return nil
	})
	return result
}

// CheckGroupCommandDisabled check global first, then check explicit disabled, must exist
func (c *StateManager) CheckGroupCommandDisabled(groupCode int64, command string) bool {
	var result bool
	_ = c.RCover(func() error {
		result = c.CheckGlobalCommandDisabled(command) || c.CheckGroupCommandFunc(groupCode, command, func(val string, exist bool) bool {
			return exist && val == Disable
		})
		return nil
	})
	return result
}

func (c *StateManager) CheckGlobalCommandDisabled(command string) bool {
	return c.CheckGlobalCommandFunc(command, func(val string, exist bool) bool {
		return exist && val == Disable
	})
}

//func (c *StateManager) CheckGlobalCommandEnabled(command string) bool {
//	return c.CheckGlobalCommandFunc(command, func(val string, exist bool) bool {
//		return exist && val == Enable
//	})
//}

func (c *StateManager) CheckGroupCommandFunc(groupCode int64, command string, f func(val string, exist bool) bool) bool {
	var result bool
	err := c.RCoverTx(func(tx *buntdb.Tx) error {
		key := c.GroupEnabledKey(groupCode, command)
		val, err := tx.Get(key)
		if err != nil && err != buntdb.ErrNotFound {
			return err
		}
		result = f(val, err == nil)
		return nil
	})
	if err != nil {
		logger.WithFields(localutils.GroupLogFields(groupCode)).
			WithField("command", command).
			Errorf("check group enable err %v", err)
		result = false
	}
	return result
}

func (c *StateManager) CheckGlobalCommandFunc(command string, f func(val string, exist bool) bool) bool {
	var result bool
	err := c.RCoverTx(func(tx *buntdb.Tx) error {
		key := c.GlobalEnabledKey(command)
		val, err := tx.Get(key)
		if err != nil && err != buntdb.ErrNotFound {
			return err
		}
		result = f(val, err == nil)
		return nil
	})
	if err != nil {
		logger.WithField("command", command).
			Errorf("check global enable err %v", err)
		result = false
	}
	return result
}

func (c *StateManager) CheckGlobalSilence() bool {
	return c.Exist(c.GlobalSilenceKey())
}

func (c *StateManager) GlobalSilence() error {
	return c.Set(c.GlobalSilenceKey(), "")
}

func (c *StateManager) UndoGlobalSilence() error {
	_, err := c.Delete(c.GlobalSilenceKey(), localdb.IgnoreNotFoundOpt())
	return err
}

func (c *StateManager) CheckGroupSilence(groupCode int64) bool {
	return c.CheckGlobalSilence() || c.Exist(c.GroupSilenceKey(groupCode))
}

func (c *StateManager) GroupSilence(groupCode int64) error {
	if c.CheckGlobalSilence() {
		return ErrGlobalSilenced
	}
	return c.Set(c.GroupSilenceKey(groupCode), "")
}

func (c *StateManager) UndoGroupSilence(groupCode int64) error {
	if c.CheckGlobalSilence() {
		return ErrGlobalSilenced
	}
	_, err := c.Delete(c.GroupSilenceKey(groupCode), localdb.IgnoreNotFoundOpt())
	return err
}

func (c *StateManager) CheckGroupAdministrator(groupCode int64, caller int64) bool {
	log := logger.WithFields(logrus.Fields{
		"GroupCode": groupCode,
		"Caller":    caller,
	})
	b := bot.Instance
	if b == nil {
		log.Errorf("bot not init")
		return false
	}
	groupInfo := b.FindGroup(groupCode)
	if groupInfo == nil {
		log.Errorf("nil group info")
		return false
	}
	log = log.WithField("GroupName", groupInfo.Name)
	groupMemberInfo := groupInfo.FindMember(caller)
	if groupMemberInfo == nil {
		log.Errorf("nil member info")
		return false
	}
	return groupMemberInfo.Permission == client.Administrator || groupMemberInfo.Permission == client.Owner
}

func (c *StateManager) CheckGroupCommandPermission(groupCode int64, caller int64, command string) bool {
	return c.CheckRole(caller, Admin) || c.Exist(c.PermissionKey(groupCode, caller, command))
}

func (c *StateManager) GrantRole(target int64, role RoleType) error {
	if role.String() == "" {
		return errors.New("error role")
	}
	err := c.Set(c.PermissionKey(target, role.String()), "", localdb.SetNoOverWriteOpt())
	if localdb.IsRollback(err) {
		return ErrPermissionExist
	}
	return err
}

func (c *StateManager) UngrantRole(target int64, role RoleType) error {
	if role.String() == "" {
		return errors.New("error role")
	}
	_, err := c.Delete(c.PermissionKey(target, role.String()))
	if localdb.IsNotFound(err) {
		return ErrPermissionNotExist
	}
	return err
}

func (c *StateManager) CheckNoAdmin() bool {
	return len(c.ListAdmin()) == 0
}

func (c *StateManager) ListAdmin() []int64 {
	var result []int64
	err := c.RCoverTx(func(tx *buntdb.Tx) error {
		return tx.Ascend(c.PermissionKey(), func(key, value string) bool {
			splits := strings.Split(key, ":")
			if len(splits) != 3 {
				return true
			}
			if NewRoleFromString(splits[2]) == Admin {
				i, err := strconv.ParseInt(splits[1], 0, 64)
				if err != nil {
					logger.WithField("Key", key).Errorf("Parse PermissionKey error %v", err)
				} else {
					result = append(result, i)
				}
			}
			return true
		})
	})
	if err != nil {
		result = nil
		logger.Errorf("ListAdmin error %v", err)
	}
	return result
}

func (c *StateManager) GrantGroupRole(groupCode int64, target int64, role RoleType) error {
	if role.String() == "" {
		return errors.New("error role")
	}
	err := c.Set(c.GroupPermissionKey(groupCode, target, role.String()), "", localdb.SetNoOverWriteOpt())
	if localdb.IsRollback(err) {
		return ErrPermissionExist
	}
	return err
}

func (c *StateManager) UngrantGroupRole(groupCode int64, target int64, role RoleType) error {
	if role.String() == "" {
		return errors.New("error role")
	}
	_, err := c.Delete(c.GroupPermissionKey(groupCode, target, role.String()))
	if localdb.IsNotFound(err) {
		return ErrPermissionNotExist
	}
	return err
}

func (c *StateManager) GrantPermission(groupCode int64, target int64, command string) error {
	if c.CheckGlobalCommandDisabled(command) {
		return ErrGlobalDisabled
	}
	err := c.Set(c.PermissionKey(groupCode, target, command), "", localdb.SetNoOverWriteOpt())
	if localdb.IsRollback(err) {
		return ErrPermissionExist
	}
	return err
}

func (c *StateManager) UngrantPermission(groupCode int64, target int64, command string) error {
	if c.CheckGlobalCommandDisabled(command) {
		return ErrGlobalDisabled
	}
	_, err := c.Delete(c.PermissionKey(groupCode, target, command))
	if localdb.IsNotFound(err) {
		return ErrPermissionNotExist
	}
	return err
}

func (c *StateManager) RequireAny(option ...RequireOption) bool {
	for _, opt := range option {
		if opt.Validate(c) {
			return true
		}
	}
	return false
}

func (c *StateManager) RemoveAllByGroupCode(groupCode int64) ([]string, error) {
	var indexKey = []string{
		c.GroupPermissionKey(),
		c.PermissionKey(),
		c.GroupEnabledKey(),
	}
	var prefixKey = []string{
		c.GroupPermissionKey(groupCode),
		c.PermissionKey(groupCode),
		c.GroupEnabledKey(groupCode),
	}
	return localdb.RemoveByPrefixAndIndex(prefixKey, indexKey)
}

func (c *StateManager) FreshIndex() {
	for _, pattern := range []localdb.KeyPatternFunc{c.PermissionKey, c.GroupPermissionKey, c.GroupEnabledKey} {
		c.CreatePatternIndex(pattern, nil)
	}
	if bot.Instance != nil {
		for _, group := range bot.Instance.GroupList {
			c.CreatePatternIndex(c.GroupPermissionKey, []interface{}{group.Code})
		}
	}
}

func NewStateManager() *StateManager {
	sm := &StateManager{
		KeySet: NewKeySet(),
	}
	return sm
}
