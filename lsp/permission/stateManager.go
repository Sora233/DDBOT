package permission

import (
	"errors"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/LagrangeDev/LagrangeGo/client/entity"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"github.com/tidwall/buntdb"

	"golang.org/x/exp/constraints"

	localdb "github.com/Sora233/DDBOT/v2/lsp/buntdb"
	localutils "github.com/Sora233/DDBOT/v2/utils"
	"github.com/Sora233/MiraiGo-Template/utils"
)

var logger = utils.GetModuleLogger("permission")

type StateManager[UT, GT constraints.Integer] struct {
	*localdb.ShortCut
	*KeySet
}

// CheckBlockList return true if blocked
func (c *StateManager[UT, GT]) CheckBlockList(caller UT) bool {
	return c.Exist(c.BlockListKey(caller))
}

func (c *StateManager[UT, GT]) AddBlockList(caller UT, d time.Duration) error {
	err := c.Set(c.BlockListKey(caller), "", localdb.SetExpireOpt(d), localdb.SetNoOverWriteOpt())
	if localdb.IsRollback(err) {
		err = localdb.ErrKeyExist
	}
	return err
}

func (c *StateManager[UT, GT]) DeleteBlockList(caller UT) error {
	_, err := c.Delete(c.BlockListKey(caller))
	return err
}

func (c *StateManager[UT, GT]) CheckRole(caller UT, role RoleType) bool {
	if role.String() == "" {
		return false
	}
	return c.Exist(c.PermissionKey(caller, role.String()))
}

func (c *StateManager[UT, GT]) CheckAdmin(caller UT) bool {
	return c.CheckRole(caller, Admin)
}

func (c *StateManager[UT, GT]) CheckGroupAdmin(groupCode GT, caller UT) bool {
	return c.CheckGroupRole(groupCode, caller, GroupAdmin)
}

func (c *StateManager[UT, GT]) CheckGroupRole(groupCode GT, caller UT, role RoleType) bool {
	if role.String() == "" {
		return false
	}
	return c.Exist(c.GroupPermissionKey(groupCode, caller, role.String()))
}

func (c *StateManager[UT, GT]) EnableGroupCommand(groupCode GT, command string) error {
	if c.CheckGlobalCommandDisabled(command) {
		return ErrGlobalDisabled
	}
	return c.operatorEnableKey(c.GroupEnabledKey(groupCode, command), Enable)
}

func (c *StateManager[UT, GT]) DisableGroupCommand(groupCode GT, command string) error {
	if c.CheckGlobalCommandDisabled(command) {
		return ErrGlobalDisabled
	}
	return c.operatorEnableKey(c.GroupEnabledKey(groupCode, command), Disable)
}

func (c *StateManager[UT, GT]) GlobalEnableGroupCommand(command string) error {
	return c.operatorEnableKey(c.GlobalEnabledKey(command), Enable)
}

func (c *StateManager[UT, GT]) GlobalDisableGroupCommand(command string) error {
	return c.operatorEnableKey(c.GlobalEnabledKey(command), Disable)
}

func (c *StateManager[UT, GT]) operatorEnableKey(key string, status string) error {
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

// CheckGroupCommandEnabled check global first, check explicit enabled, must exist
func (c *StateManager[UT, GT]) CheckGroupCommandEnabled(groupCode GT, command string) bool {
	var result bool
	_ = c.RCover(func() error {
		result = !c.CheckGlobalCommandDisabled(command) &&
			c.CheckGroupCommandFunc(groupCode, command, func(val string, exist bool) bool {
				return exist && val == Enable
			})
		return nil
	})
	return result
}

// CheckGroupCommandDisabled check global first, then check explicit disabled, must exist
func (c *StateManager[UT, GT]) CheckGroupCommandDisabled(groupCode GT, command string) bool {
	var result bool
	_ = c.RCover(func() error {
		result = c.CheckGlobalCommandDisabled(command) ||
			c.CheckGroupCommandFunc(groupCode, command, func(val string, exist bool) bool {
				return exist && val == Disable
			})
		return nil
	})
	return result
}

func (c *StateManager[UT, GT]) CheckGlobalCommandDisabled(command string) bool {
	return c.CheckGlobalCommandFunc(command, func(val string, exist bool) bool {
		return exist && val == Disable
	})
}

func (c *StateManager[UT, GT]) CheckGroupCommandFunc(groupCode GT, command string, f func(val string, exist bool) bool) bool {
	var result bool
	err := c.RCoverTx(func(tx *buntdb.Tx) error {
		val, err := c.Get(c.GroupEnabledKey(groupCode, command))
		if err != nil && !localdb.IsNotFound(err) {
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

func (c *StateManager[UT, GT]) CheckGlobalCommandFunc(command string, f func(val string, exist bool) bool) bool {
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

func (c *StateManager[UT, GT]) CheckGlobalSilence() bool {
	return c.Exist(c.GlobalSilenceKey())
}

func (c *StateManager[UT, GT]) GlobalSilence() error {
	return c.Set(c.GlobalSilenceKey(), "")
}

func (c *StateManager[UT, GT]) UndoGlobalSilence() error {
	_, err := c.Delete(c.GlobalSilenceKey(), localdb.IgnoreNotFoundOpt())
	return err
}

func (c *StateManager[UT, GT]) CheckGroupSilence(groupCode GT) bool {
	return c.CheckGlobalSilence() || c.Exist(c.GroupSilenceKey(groupCode))
}

func (c *StateManager[UT, GT]) GroupSilence(groupCode GT) error {
	if c.CheckGlobalSilence() {
		return ErrGlobalSilenced
	}
	return c.Set(c.GroupSilenceKey(groupCode), "")
}

func (c *StateManager[UT, GT]) UndoGroupSilence(groupCode GT) error {
	if c.CheckGlobalSilence() {
		return ErrGlobalSilenced
	}
	_, err := c.Delete(c.GroupSilenceKey(groupCode), localdb.IgnoreNotFoundOpt())
	return err
}

func (c *StateManager[UT, GT]) CheckGroupAdministrator(groupCode GT, caller UT) bool {
	log := logger.WithFields(logrus.Fields{
		"GroupCode": groupCode,
		"Caller":    caller,
	})
	groupInfo := localutils.GetBot().FindGroup(uint32(groupCode))
	if groupInfo == nil {
		log.Errorf("nil group info")
		return false
	}
	log = log.WithField("GroupName", groupInfo.GroupName)
	groupMemberInfo := localutils.GetBot().FindGroupMember(uint32(groupCode), uint32(caller))
	if groupMemberInfo == nil {
		log.Errorf("nil member info")
		return false
	}
	return lo.Contains([]entity.GroupMemberPermission{entity.Admin, entity.Owner}, groupMemberInfo.Permission)
}

func (c *StateManager[UT, GT]) CheckGroupCommandPermission(groupCode GT, caller UT, command string) bool {
	return c.CheckRole(caller, Admin) || c.Exist(c.PermissionKey(groupCode, caller, command))
}

func (c *StateManager[UT, GT]) GrantRole(target UT, role RoleType) error {
	if role.String() == "" {
		return errors.New("error role")
	}
	err := c.Set(c.PermissionKey(target, role.String()), "", localdb.SetNoOverWriteOpt())
	if localdb.IsRollback(err) {
		return ErrPermissionExist
	}
	return err
}

func (c *StateManager[UT, GT]) UngrantRole(target UT, role RoleType) error {
	if role.String() == "" {
		return errors.New("error role")
	}
	_, err := c.Delete(c.PermissionKey(target, role.String()))
	if localdb.IsNotFound(err) {
		return ErrPermissionNotExist
	}
	return err
}

func (c *StateManager[UT, GT]) CheckNoAdmin() bool {
	return len(c.ListAdmin()) == 0
}

func (c *StateManager[UT, GT]) ListAdmin() []UT {
	var result []UT
	err := c.RCoverTx(func(tx *buntdb.Tx) error {
		return tx.Ascend(c.PermissionKey(), func(key, value string) bool {
			splits := strings.Split(key, ":")
			if len(splits) != 3 {
				return true
			}
			if NewRoleFromString(splits[2]) == Admin {
				var i UT
				var err error
				switch reflect.TypeOf(i).Kind() {
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					var x int64
					x, err = strconv.ParseInt(splits[1], 0, 64)
					i = UT(x)
				case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
					var x uint64
					x, err = strconv.ParseUint(splits[1], 0, 64)
					i = UT(x)
				default:
					panic("unhandled default case")
				}
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

func (c *StateManager[UT, GT]) ListGroupAdmin(groupCode GT) []UT {
	var result []UT
	err := c.RCoverTx(func(tx *buntdb.Tx) error {
		return tx.Ascend(c.GroupPermissionKey(groupCode), func(key, value string) bool {
			splits := strings.Split(key, ":")
			if len(splits) != 4 {
				return true
			}
			if NewRoleFromString(splits[3]) == GroupAdmin {
				var i UT
				var err error
				switch reflect.TypeOf(i).Kind() {
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					var x int64
					x, err = strconv.ParseInt(splits[1], 0, 64)
					i = UT(x)
				case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
					var x uint64
					x, err = strconv.ParseUint(splits[1], 0, 64)
					i = UT(x)
				}
				if err != nil {
					logger.WithField("Key", key).Errorf("Parse GroupPermissionKey error %v", err)
				} else {
					result = append(result, i)
				}
			}
			return true
		})
	})
	if err != nil {
		result = nil
		logger.Errorf("ListGroupAdmin error %v", err)
	}
	return result
}

func (c *StateManager[UT, GT]) GrantGroupRole(groupCode GT, target UT, role RoleType) error {
	if role.String() == "" {
		return errors.New("error role")
	}
	err := c.Set(c.GroupPermissionKey(groupCode, target, role.String()), "", localdb.SetNoOverWriteOpt())
	if localdb.IsRollback(err) {
		return ErrPermissionExist
	}
	return err
}

func (c *StateManager[UT, GT]) UngrantGroupRole(groupCode GT, target UT, role RoleType) error {
	if role.String() == "" {
		return errors.New("error role")
	}
	_, err := c.Delete(c.GroupPermissionKey(groupCode, target, role.String()))
	if localdb.IsNotFound(err) {
		return ErrPermissionNotExist
	}
	return err
}

func (c *StateManager[UT, GT]) GrantPermission(groupCode GT, target UT, command string) error {
	if c.CheckGlobalCommandDisabled(command) {
		return ErrGlobalDisabled
	}
	err := c.Set(c.PermissionKey(groupCode, target, command), "", localdb.SetNoOverWriteOpt())
	if localdb.IsRollback(err) {
		return ErrPermissionExist
	}
	return err
}

func (c *StateManager[UT, GT]) UngrantPermission(groupCode GT, target UT, command string) error {
	if c.CheckGlobalCommandDisabled(command) {
		return ErrGlobalDisabled
	}
	_, err := c.Delete(c.PermissionKey(groupCode, target, command))
	if localdb.IsNotFound(err) {
		return ErrPermissionNotExist
	}
	return err
}

func (c *StateManager[UT, GT]) RequireAny(option ...RequireOption[UT, GT]) bool {
	for _, opt := range option {
		if opt.Validate(c) {
			return true
		}
	}
	return false
}

func (c *StateManager[UT, GT]) RemoveAllByGroupCode(groupCode GT) ([]string, error) {
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

func (c *StateManager[UT, GT]) FreshIndex() {
	for _, pattern := range []localdb.KeyPatternFunc{c.PermissionKey, c.GroupPermissionKey, c.GroupEnabledKey} {
		c.CreatePatternIndex(pattern, nil)
	}
	for _, group := range localutils.GetBot().GetGroupList() {
		c.CreatePatternIndex(c.GroupPermissionKey, []interface{}{group.GroupUin})
		c.CreatePatternIndex(c.GroupEnabledKey, []interface{}{group.GroupUin})
	}
}

func NewStateManager[UT, GT constraints.Integer]() *StateManager[UT, GT] {
	sm := &StateManager[UT, GT]{
		KeySet: NewKeySet(),
	}
	return sm
}
