package permission

import (
	"github.com/Sora233/DDBOT/internal/test"
	localutils "github.com/Sora233/DDBOT/utils"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func initStateManager(t *testing.T) *StateManager {
	sm := NewStateManager()
	assert.NotNil(t, sm)
	sm.FreshIndex()
	return sm
}

func TestStateManager_CheckBlockList(t *testing.T) {
	test.InitBuntdb(t)
	defer test.CloseBuntdb(t)
	c := initStateManager(t)

	assert.False(t, c.CheckBlockList(test.UID1))
	assert.Nil(t, c.AddBlockList(test.UID1, time.Hour*24))
	assert.True(t, c.CheckBlockList(test.UID1))
	assert.False(t, c.CheckBlockList(test.UID2))

	assert.Nil(t, c.DeleteBlockList(test.UID1))
	assert.False(t, c.CheckBlockList(test.UID1))

	assert.Nil(t, c.AddBlockList(test.UID1, time.Millisecond*100))
	assert.True(t, c.CheckBlockList(test.UID1))
	time.Sleep(time.Millisecond * 150)
	assert.False(t, c.CheckBlockList(test.UID1))
}

func TestStateManager_CheckGroupCommandFunc(t *testing.T) {
	test.InitBuntdb(t)
	defer test.CloseBuntdb(t)
	c := initStateManager(t)

	assert.False(t, c.CheckGroupCommandEnabled(test.G1, test.CMD1))
	assert.False(t, c.CheckGroupCommandDisabled(test.G1, test.CMD1))
	assert.Nil(t, c.EnableGroupCommand(test.G1, test.CMD1))
	assert.NotNil(t, c.EnableGroupCommand(test.G1, test.CMD1))

	assert.True(t, c.CheckGroupCommandEnabled(test.G1, test.CMD1))
	assert.False(t, c.CheckGroupCommandDisabled(test.G1, test.CMD1))

	assert.Nil(t, c.DisableGroupCommand(test.G1, test.CMD1))
	assert.NotNil(t, c.DisableGroupCommand(test.G1, test.CMD1))

	assert.False(t, c.CheckGroupCommandEnabled(test.G1, test.CMD1))
	assert.True(t, c.CheckGroupCommandDisabled(test.G1, test.CMD1))

	assert.Nil(t, c.EnableGroupCommand(test.G1, test.CMD1))
	assert.True(t, c.CheckGroupCommandEnabled(test.G1, test.CMD1))
	assert.False(t, c.CheckGroupCommandDisabled(test.G1, test.CMD1))

	assert.Nil(t, c.DisableGroupCommand(test.G1, test.CMD1))
	assert.False(t, c.CheckGroupCommandEnabled(test.G1, test.CMD1))
	assert.True(t, c.CheckGroupCommandDisabled(test.G1, test.CMD1))

}

func TestStateManager_CheckGlobalCommandFunc(t *testing.T) {
	test.InitBuntdb(t)
	defer test.CloseBuntdb(t)
	c := initStateManager(t)

	assert.False(t, c.CheckGlobalCommandDisabled(test.CMD1))
	assert.Nil(t, c.GlobalDisableGroupCommand(test.CMD1))
	assert.True(t, c.CheckGlobalCommandDisabled(test.CMD1))
	assert.False(t, c.CheckGlobalCommandDisabled(test.CMD2))

	assert.Nil(t, c.GlobalEnableGroupCommand(test.CMD1))
	assert.NotNil(t, c.GlobalEnableGroupCommand(test.CMD1))
	assert.False(t, c.CheckGlobalCommandDisabled(test.CMD1))
	assert.Nil(t, c.GlobalDisableGroupCommand(test.CMD1))

	assert.Nil(t, c.GlobalEnableGroupCommand(test.CMD1))

	assert.Nil(t, c.GlobalDisableGroupCommand(test.CMD1))
	assert.Equal(t, ErrGlobalDisabled, c.EnableGroupCommand(test.G1, test.CMD1))
	assert.Equal(t, ErrGlobalDisabled, c.DisableGroupCommand(test.G1, test.CMD1))
	assert.Nil(t, c.GlobalEnableGroupCommand(test.CMD1))
	assert.Nil(t, c.EnableGroupCommand(test.G1, test.CMD1))
	assert.True(t, c.CheckGroupCommandEnabled(test.G1, test.CMD1))

	assert.Nil(t, c.GlobalDisableGroupCommand(test.CMD1))
	assert.False(t, c.CheckGroupCommandEnabled(test.G1, test.CMD1))
	assert.True(t, c.CheckGroupCommandDisabled(test.G1, test.CMD1))
	assert.Equal(t, ErrGlobalDisabled, c.GrantPermission(test.G1, test.UID1, test.CMD1))
	assert.Equal(t, ErrGlobalDisabled, c.UngrantPermission(test.G1, test.UID1, test.CMD1))
}

func TestStateManager_CheckRole(t *testing.T) {
	test.InitBuntdb(t)
	defer test.CloseBuntdb(t)
	c := initStateManager(t)

	assert.False(t, c.CheckRole(test.UID1, RoleType(-1)))
	assert.False(t, c.CheckGroupRole(test.UID1, test.G1, RoleType(-1)))
	assert.False(t, c.CheckAdmin(test.UID1))

	adminOpt1 := AdminRoleRequireOption(test.UID1)
	adminOpt2 := AdminRoleRequireOption(test.UID2)
	assert.False(t, adminOpt1.Validate(c))
	assert.False(t, adminOpt2.Validate(c))
	assert.False(t, c.RequireAny(adminOpt1, adminOpt2))

	assert.NotNil(t, c.GrantRole(test.UID2, RoleType(-1)))
	assert.Nil(t, c.GrantRole(test.UID2, Admin))
	assert.False(t, adminOpt1.Validate(c))
	assert.True(t, adminOpt2.Validate(c))
	assert.True(t, c.RequireAny(adminOpt1, adminOpt2))
	assert.True(t, c.CheckAdmin(test.UID2))

	assert.NotNil(t, c.GrantRole(test.UID2, Admin))

	assert.NotNil(t, c.UngrantRole(test.UID2, RoleType(-1)))
	assert.Nil(t, c.UngrantRole(test.UID2, Admin))

	assert.False(t, adminOpt1.Validate(c))
	assert.False(t, adminOpt2.Validate(c))
	assert.False(t, c.RequireAny(adminOpt1, adminOpt2))
	assert.False(t, c.CheckAdmin(test.UID2))

	assert.NotNil(t, c.UngrantRole(test.UID2, Admin))
}

func TestStateManager_CheckGroupRole(t *testing.T) {
	test.InitBuntdb(t)
	defer test.CloseBuntdb(t)
	c := initStateManager(t)
	localutils.GetBot().TESTAddGroup(test.G1)
	localutils.GetBot().TESTAddGroup(test.G2)
	c.FreshIndex()

	gadminOpt1 := GroupAdminRoleRequireOption(test.G1, test.UID1)
	gadminOpt2 := GroupAdminRoleRequireOption(test.G2, test.UID1)
	assert.False(t, gadminOpt1.Validate(c))
	assert.False(t, gadminOpt2.Validate(c))
	assert.False(t, c.RequireAny(gadminOpt1, gadminOpt2))
	assert.False(t, c.CheckGroupAdmin(test.G1, test.UID1))

	assert.NotNil(t, c.GrantGroupRole(test.G2, test.UID1, RoleType(-1)))
	assert.Nil(t, c.GrantGroupRole(test.G2, test.UID1, GroupAdmin))
	assert.False(t, gadminOpt1.Validate(c))
	assert.True(t, gadminOpt2.Validate(c))
	assert.True(t, c.RequireAny(gadminOpt1, gadminOpt2))
	assert.True(t, c.CheckGroupAdmin(test.G2, test.UID1))

	ids := c.ListGroupAdmin(test.G1)
	assert.Empty(t, ids)
	ids = c.ListGroupAdmin(test.G2)
	assert.Len(t, ids, 1)
	assert.EqualValues(t, test.UID1, ids[0])

	assert.NotNil(t, c.GrantGroupRole(test.G2, test.UID1, RoleType(-1)))
	assert.NotNil(t, c.GrantGroupRole(test.G2, test.UID1, GroupAdmin))

	assert.Nil(t, c.UngrantGroupRole(test.G2, test.UID1, GroupAdmin))
	assert.False(t, gadminOpt1.Validate(c))
	assert.False(t, gadminOpt2.Validate(c))
	assert.False(t, c.RequireAny(gadminOpt1, gadminOpt2))
	assert.False(t, c.CheckGroupAdmin(test.G2, test.UID1))

	assert.NotNil(t, c.UngrantGroupRole(test.G2, test.UID1, RoleType(-1)))
	assert.NotNil(t, c.UngrantGroupRole(test.G2, test.UID1, GroupAdmin))

	ids = c.ListGroupAdmin(test.G2)
	assert.Empty(t, ids)
}

func TestStateManager_CheckGroupCommandPermission(t *testing.T) {
	test.InitBuntdb(t)
	defer test.CloseBuntdb(t)
	c := initStateManager(t)

	opt := QQAdminRequireOption(test.G1, test.UID1)
	assert.False(t, opt.Validate(c))

	gcadminOpt1 := GroupCommandRequireOption(test.G1, test.UID1, test.CMD1)
	gcadminOpt2 := GroupCommandRequireOption(test.G1, test.UID1, test.CMD2)
	assert.False(t, gcadminOpt1.Validate(c))
	assert.False(t, gcadminOpt2.Validate(c))
	assert.False(t, c.RequireAny(gcadminOpt1, gcadminOpt2))

	assert.Nil(t, c.GrantPermission(test.G1, test.UID1, test.CMD2))
	assert.False(t, gcadminOpt1.Validate(c))
	assert.True(t, gcadminOpt2.Validate(c))
	assert.True(t, c.RequireAny(gcadminOpt1, gcadminOpt2))

	assert.NotNil(t, c.GrantPermission(test.G1, test.UID1, test.CMD2))

	assert.Nil(t, c.UngrantPermission(test.G1, test.UID1, test.CMD2))
	assert.False(t, gcadminOpt1.Validate(c))
	assert.False(t, gcadminOpt2.Validate(c))
	assert.False(t, c.RequireAny(gcadminOpt1, gcadminOpt2))

	assert.NotNil(t, c.UngrantPermission(test.G1, test.UID1, test.CMD2))
}

func TestStateManager_RemoveAllByGroup(t *testing.T) {
	test.InitBuntdb(t)
	defer test.CloseBuntdb(t)
	c := initStateManager(t)

	assert.Nil(t, c.CreatePatternIndex(c.GroupPermissionKey, []interface{}{test.G1}))
	assert.Nil(t, c.CreatePatternIndex(c.GroupPermissionKey, []interface{}{test.G2}))

	assert.Nil(t, c.GrantGroupRole(test.G1, test.UID1, GroupAdmin))
	assert.Nil(t, c.GrantGroupRole(test.G2, test.UID1, GroupAdmin))

	gadminOpt1 := GroupAdminRoleRequireOption(test.G1, test.UID1)
	// 当发现group command 应该用GroupPermissionKey的时候已经太晚了
	//gcadminOpt1 := GroupCommandRequireOption(test.G1, test.UID1, test.CMD1)
	gcadminOpt2 := GroupAdminRoleRequireOption(test.G2, test.UID1)

	assert.True(t, gadminOpt1.Validate(c))
	//assert.True(t, gcadminOpt1.Validate(c))
	assert.True(t, gcadminOpt2.Validate(c))

	_, err := c.RemoveAllByGroupCode(test.G1)
	assert.Nil(t, err)

	assert.False(t, gadminOpt1.Validate(c))
	//assert.False(t, gcadminOpt1.Validate(c))
	assert.True(t, gcadminOpt2.Validate(c))
}

func TestStateManager_CheckNoAdmin(t *testing.T) {
	test.InitBuntdb(t)
	defer test.CloseBuntdb(t)
	c := initStateManager(t)

	assert.True(t, c.CheckNoAdmin())

	assert.Nil(t, c.GrantRole(test.UID1, Admin))
	assert.False(t, c.CheckNoAdmin())
	ids := c.ListAdmin()
	assert.Len(t, ids, 1)
	assert.EqualValues(t, test.UID1, ids[0])
	assert.Nil(t, c.UngrantRole(test.UID1, Admin))
	assert.True(t, c.CheckNoAdmin())
	ids = c.ListAdmin()
	assert.Empty(t, ids)

	assert.Nil(t, c.GrantGroupRole(test.G1, test.UID1, GroupAdmin))
	assert.True(t, c.CheckNoAdmin())
	ids = c.ListAdmin()
	assert.Empty(t, ids)
	assert.Nil(t, c.GrantRole(test.UID1, Admin))
	assert.False(t, c.CheckNoAdmin())
	ids = c.ListAdmin()
	assert.Len(t, ids, 1)
	assert.EqualValues(t, test.UID1, ids[0])

	assert.Nil(t, c.GrantRole(test.UID2, Admin))
	assert.False(t, c.CheckNoAdmin())
	ids = c.ListAdmin()
	assert.Len(t, ids, 2)
	assert.Contains(t, ids, test.UID1)
	assert.Contains(t, ids, test.UID2)
}

func TestStateManager_CheckGlobalSilence(t *testing.T) {
	test.InitBuntdb(t)
	defer test.CloseBuntdb(t)
	c := initStateManager(t)

	assert.False(t, c.CheckGlobalSilence())
	assert.Nil(t, c.GlobalSilence())
	assert.Nil(t, c.GlobalSilence())
	assert.True(t, c.CheckGlobalSilence())
	assert.True(t, c.CheckGlobalSilence())

	assert.Nil(t, c.UndoGlobalSilence())
	assert.Nil(t, c.UndoGlobalSilence())
	assert.False(t, c.CheckGlobalSilence())
	assert.False(t, c.CheckGlobalSilence())
}

func TestStateManager_CheckGroupSilence(t *testing.T) {
	test.InitBuntdb(t)
	defer test.CloseBuntdb(t)
	c := initStateManager(t)

	assert.False(t, c.CheckGroupSilence(test.G1))
	assert.False(t, c.CheckGroupSilence(test.G2))

	assert.Nil(t, c.GroupSilence(test.G1))

	assert.True(t, c.CheckGroupSilence(test.G1))
	assert.False(t, c.CheckGroupSilence(test.G2))

	assert.Nil(t, c.UndoGroupSilence(test.G1))

	assert.False(t, c.CheckGroupSilence(test.G1))
	assert.False(t, c.CheckGroupSilence(test.G2))

	assert.Nil(t, c.GlobalSilence())
	assert.True(t, c.CheckGroupSilence(test.G1))
	assert.True(t, c.CheckGroupSilence(test.G2))

	assert.NotNil(t, c.GroupSilence(test.G1))
	assert.NotNil(t, c.UndoGroupSilence(test.G1))
}
