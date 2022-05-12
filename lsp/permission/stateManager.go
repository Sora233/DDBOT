package permission

import (
	"errors"
	"github.com/Mrs4s/MiraiGo/client"
	localdb "github.com/Sora233/DDBOT/lsp/buntdb"
	"github.com/Sora233/DDBOT/lsp/mmsg/mt"
	localutils "github.com/Sora233/DDBOT/utils"
	"github.com/Sora233/MiraiGo-Template/utils"
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
	return c.Exist(c.BlockListKey(caller))
}

func (c *StateManager) AddBlockList(caller int64, d time.Duration) error {
	err := c.Set(c.BlockListKey(caller), "", localdb.SetExpireOpt(d), localdb.SetNoOverWriteOpt())
	if localdb.IsRollback(err) {
		err = localdb.ErrKeyExist
	}
	return err
}

func (c *StateManager) DeleteBlockList(caller int64) error {
	_, err := c.Delete(c.BlockListKey(caller))
	return err
}

func (c *StateManager) CheckRole(caller int64, role RoleType) bool {
	if role.String() == "" {
		return false
	}
	return c.Exist(c.PermissionKey(caller, role.String()))
}

func (c *StateManager) CheckAdmin(caller int64) bool {
	return c.CheckRole(caller, Admin)
}

func (c *StateManager) CheckTargetAdmin(target mt.Target, caller int64) bool {
	return c.CheckTargetRole(target, caller, TargetAdmin)
}

func (c *StateManager) CheckTargetRole(target mt.Target, caller int64, role RoleType) bool {
	if role.String() == "" {
		return false
	}
	return c.Exist(c.TargetPermissionKey(target, caller, role.String()))
}

func (c *StateManager) EnableTargetCommand(target mt.Target, command string) error {
	if c.CheckGlobalCommandDisabled(command) {
		return ErrGlobalDisabled
	}
	return c.operatorEnableKey(c.TargetEnabledKey(target, command), Enable)
}

func (c *StateManager) DisableTargetCommand(target mt.Target, command string) error {
	if c.CheckGlobalCommandDisabled(command) {
		return ErrGlobalDisabled
	}
	return c.operatorEnableKey(c.TargetEnabledKey(target, command), Disable)
}

func (c *StateManager) GlobalEnableCommand(command string) error {
	return c.operatorEnableKey(c.GlobalEnabledKey(command), Enable)
}

func (c *StateManager) GlobalDisableCommand(command string) error {
	return c.operatorEnableKey(c.GlobalEnabledKey(command), Disable)
}

func (c *StateManager) operatorEnableKey(key string, status string) error {
	var prev string
	var isOverwrite bool
	err := c.Set(key, status, localdb.SetGetIsOverwriteOpt(&isOverwrite), localdb.SetGetPreviousValueStringOpt(&prev))
	if err != nil {
		return err
	}
	if isOverwrite && prev == status {
		return ErrPermissionExist
	}
	return nil
}

// CheckTargetCommandEnabled check global first, check explicit enabled, must exist
func (c *StateManager) CheckTargetCommandEnabled(target mt.Target, command string) bool {
	var result bool
	_ = c.RCover(func() error {
		result = !c.CheckGlobalCommandDisabled(command) &&
			c.CheckTargetCommandFunc(target, command, func(val string, exist bool) bool {
				return exist && val == Enable
			})
		return nil
	})
	return result
}

// CheckTargetCommandDisabled check global first, then check explicit disabled, must exist
func (c *StateManager) CheckTargetCommandDisabled(target mt.Target, command string) bool {
	var result bool
	_ = c.RCover(func() error {
		result = c.CheckGlobalCommandDisabled(command) ||
			c.CheckTargetCommandFunc(target, command, func(val string, exist bool) bool {
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

func (c *StateManager) CheckTargetCommandFunc(target mt.Target, command string, f func(val string, exist bool) bool) bool {
	var result bool
	err := c.RCoverTx(func(tx *buntdb.Tx) error {
		val, err := c.Get(c.TargetEnabledKey(target, command))
		if err != nil && !localdb.IsNotFound(err) {
			return err
		}
		result = f(val, err == nil)
		return nil
	})
	if err != nil {
		logger.WithFields(localutils.TargetFields(target)).
			WithField("command", command).
			Errorf("check group enable err %v", err)
		result = false
	}
	return result
}

func (c *StateManager) CheckGlobalCommandFunc(command string, f func(val string, exist bool) bool) bool {
	var result bool
	err := c.RCoverTx(func(tx *buntdb.Tx) error {
		val, err := c.Get(c.GlobalEnabledKey(command))
		if err != nil && !localdb.IsNotFound(err) {
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

func (c *StateManager) CheckTargetSilence(target mt.Target) bool {
	return c.CheckGlobalSilence() || c.Exist(c.TargetSilenceKey(target))
}

func (c *StateManager) TargetSilence(target mt.Target) error {
	if c.CheckGlobalSilence() {
		return ErrGlobalSilenced
	}
	return c.Set(c.TargetSilenceKey(target), "")
}

func (c *StateManager) UndoTargetSilence(target mt.Target) error {
	if c.CheckGlobalSilence() {
		return ErrGlobalSilenced
	}
	_, err := c.Delete(c.TargetSilenceKey(target), localdb.IgnoreNotFoundOpt())
	return err
}

func (c *StateManager) CheckGroupAdministrator(target mt.Target, caller int64) bool {
	log := logger.WithField("Caller", caller).WithFields(localutils.TargetFields(target))
	if target.GetTargetType().IsPrivate() {
		return false
	}
	if target.GetTargetType().IsGroup() {
		groupInfo := localutils.GetBot().FindGroup(target.(*mt.GroupTarget).GroupCode)
		if groupInfo == nil {
			log.Errorf("nil group info")
			return false
		}
		groupMemberInfo := groupInfo.FindMember(caller)
		if groupMemberInfo == nil {
			log.Errorf("nil member info")
			return false
		}
		return groupMemberInfo.Permission == client.Administrator || groupMemberInfo.Permission == client.Owner
	} else if target.GetTargetType().IsGulid() {
		// TODO
		return false
	}
	return false
}

func (c *StateManager) CheckTargetCommandPermission(target mt.Target, caller int64, command string) bool {
	return c.CheckRole(caller, Admin) || c.Exist(c.TargetPermissionKey(target, caller, command))
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

func (c *StateManager) ListTargetAdmin(target mt.Target) []int64 {
	var result []int64
	err := c.RCoverTx(func(tx *buntdb.Tx) error {
		return tx.Ascend(c.TargetPermissionKey(), func(key, value string) bool {
			splits := strings.Split(key, ":")
			if len(splits) != 4 {
				return true
			}
			t := mt.ParseTargetHash(splits[1])
			if t == nil {
				return true
			}
			if !t.Equal(target) {
				return true
			}
			if NewRoleFromString(splits[3]) == TargetAdmin {
				i, err := strconv.ParseInt(splits[2], 0, 64)
				if err != nil {
					logger.WithField("Key", key).Errorf("Parse TargetPermissionKey error %v", err)
				} else {
					result = append(result, i)
				}
			}
			return true
		})
	})
	if err != nil {
		result = nil
		logger.Errorf("ListTargetAdmin error %v", err)
	}
	return result
}

func (c *StateManager) GrantTargetRole(target mt.Target, uin int64, role RoleType) error {
	if role.String() == "" {
		return errors.New("error role")
	}
	err := c.Set(c.TargetPermissionKey(target, uin, role.String()), "", localdb.SetNoOverWriteOpt())
	if localdb.IsRollback(err) {
		return ErrPermissionExist
	}
	return err
}

func (c *StateManager) UngrantTargetRole(target mt.Target, uin int64, role RoleType) error {
	if role.String() == "" {
		return errors.New("error role")
	}
	_, err := c.Delete(c.TargetPermissionKey(target, uin, role.String()))
	if localdb.IsNotFound(err) {
		return ErrPermissionNotExist
	}
	return err
}

func (c *StateManager) TargetGrantPermission(target mt.Target, uin int64, command string) error {
	if c.CheckGlobalCommandDisabled(command) {
		return ErrGlobalDisabled
	}
	err := c.Set(c.TargetPermissionKey(target, uin, command), "", localdb.SetNoOverWriteOpt())
	if localdb.IsRollback(err) {
		return ErrPermissionExist
	}
	return err
}

func (c *StateManager) UngrantPermission(target mt.Target, uin int64, command string) error {
	if c.CheckGlobalCommandDisabled(command) {
		return ErrGlobalDisabled
	}
	_, err := c.Delete(c.TargetPermissionKey(target, uin, command))
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

func (c *StateManager) RemoveAllByTarget(target mt.Target) ([]string, error) {
	var indexKey = []string{
		c.TargetPermissionKey(),
		c.PermissionKey(),
		c.TargetEnabledKey(),
	}
	var prefixKey = []string{
		c.TargetPermissionKey(target),
		c.PermissionKey(target),
		c.TargetEnabledKey(target),
	}
	return localdb.RemoveByPrefixAndIndex(prefixKey, indexKey)
}

func (c *StateManager) FreshIndex() {
	for _, pattern := range []localdb.KeyPatternFunc{c.PermissionKey, c.TargetPermissionKey, c.TargetEnabledKey} {
		c.CreatePatternIndex(pattern, nil)
	}
}

func NewStateManager() *StateManager {
	sm := &StateManager{
		KeySet: NewKeySet(),
	}
	return sm
}
