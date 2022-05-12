package permission

import (
	"github.com/Sora233/DDBOT/internal/test"
	"github.com/Sora233/DDBOT/lsp/mmsg/mt"
	localutils "github.com/Sora233/DDBOT/utils"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var (
	g1 = mt.NewGroupTarget(test.G1)
	g2 = mt.NewGroupTarget(test.G2)
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

	assert.False(t, c.CheckTargetCommandEnabled(g1, test.CMD1))
	assert.False(t, c.CheckTargetCommandDisabled(g1, test.CMD1))
	assert.Nil(t, c.EnableTargetCommand(g1, test.CMD1))
	assert.NotNil(t, c.EnableTargetCommand(g1, test.CMD1))

	assert.True(t, c.CheckTargetCommandEnabled(g1, test.CMD1))
	assert.False(t, c.CheckTargetCommandDisabled(g1, test.CMD1))

	assert.Nil(t, c.DisableTargetCommand(g1, test.CMD1))
	assert.NotNil(t, c.DisableTargetCommand(g1, test.CMD1))

	assert.False(t, c.CheckTargetCommandEnabled(g1, test.CMD1))
	assert.True(t, c.CheckTargetCommandDisabled(g1, test.CMD1))

	assert.Nil(t, c.EnableTargetCommand(g1, test.CMD1))
	assert.True(t, c.CheckTargetCommandEnabled(g1, test.CMD1))
	assert.False(t, c.CheckTargetCommandDisabled(g1, test.CMD1))

	assert.Nil(t, c.DisableTargetCommand(g1, test.CMD1))
	assert.False(t, c.CheckTargetCommandEnabled(g1, test.CMD1))
	assert.True(t, c.CheckTargetCommandDisabled(g1, test.CMD1))

}

func TestStateManager_CheckGlobalCommandFunc(t *testing.T) {
	test.InitBuntdb(t)
	defer test.CloseBuntdb(t)
	c := initStateManager(t)

	assert.False(t, c.CheckGlobalCommandDisabled(test.CMD1))
	assert.Nil(t, c.GlobalDisableCommand(test.CMD1))
	assert.True(t, c.CheckGlobalCommandDisabled(test.CMD1))
	assert.False(t, c.CheckGlobalCommandDisabled(test.CMD2))

	assert.Nil(t, c.GlobalEnableCommand(test.CMD1))
	assert.NotNil(t, c.GlobalEnableCommand(test.CMD1))
	assert.False(t, c.CheckGlobalCommandDisabled(test.CMD1))
	assert.Nil(t, c.GlobalDisableCommand(test.CMD1))

	assert.Nil(t, c.GlobalEnableCommand(test.CMD1))

	assert.Nil(t, c.GlobalDisableCommand(test.CMD1))
	assert.Equal(t, ErrGlobalDisabled, c.EnableTargetCommand(g1, test.CMD1))
	assert.Equal(t, ErrGlobalDisabled, c.DisableTargetCommand(g1, test.CMD1))
	assert.Nil(t, c.GlobalEnableCommand(test.CMD1))
	assert.Nil(t, c.EnableTargetCommand(g1, test.CMD1))
	assert.True(t, c.CheckTargetCommandEnabled(g1, test.CMD1))

	assert.Nil(t, c.GlobalDisableCommand(test.CMD1))
	assert.False(t, c.CheckTargetCommandEnabled(g1, test.CMD1))
	assert.True(t, c.CheckTargetCommandDisabled(g1, test.CMD1))
	assert.Equal(t, ErrGlobalDisabled, c.TargetGrantPermission(g1, test.UID1, test.CMD1))
	assert.Equal(t, ErrGlobalDisabled, c.UngrantPermission(g1, test.UID1, test.CMD1))
}

func TestStateManager_CheckRole(t *testing.T) {
	test.InitBuntdb(t)
	defer test.CloseBuntdb(t)
	c := initStateManager(t)

	assert.False(t, c.CheckRole(test.UID1, RoleType(-1)))
	assert.False(t, c.CheckTargetRole(g1, test.UID1, RoleType(-1)))
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
	localutils.GetBot().TESTAddGroup(g1.GroupCode)
	localutils.GetBot().TESTAddGroup(g2.GroupCode)
	c.FreshIndex()

	gadminOpt1 := TargetAdminRoleRequireOption(g1, test.UID1)
	gadminOpt2 := TargetAdminRoleRequireOption(g2, test.UID1)
	assert.False(t, gadminOpt1.Validate(c))
	assert.False(t, gadminOpt2.Validate(c))
	assert.False(t, c.RequireAny(gadminOpt1, gadminOpt2))
	assert.False(t, c.CheckTargetAdmin(g1, test.UID1))

	assert.NotNil(t, c.GrantTargetRole(g2, test.UID1, RoleType(-1)))
	assert.Nil(t, c.GrantTargetRole(g2, test.UID1, TargetAdmin))
	assert.False(t, gadminOpt1.Validate(c))
	assert.True(t, gadminOpt2.Validate(c))
	assert.True(t, c.RequireAny(gadminOpt1, gadminOpt2))
	assert.True(t, c.CheckTargetAdmin(g2, test.UID1))

	ids := c.ListTargetAdmin(g1)
	assert.Empty(t, ids)
	ids = c.ListTargetAdmin(g2)
	assert.Len(t, ids, 1)
	assert.EqualValues(t, test.UID1, ids[0])

	assert.NotNil(t, c.GrantTargetRole(g2, test.UID1, RoleType(-1)))
	assert.NotNil(t, c.GrantTargetRole(g2, test.UID1, TargetAdmin))

	assert.Nil(t, c.UngrantTargetRole(g2, test.UID1, TargetAdmin))
	assert.False(t, gadminOpt1.Validate(c))
	assert.False(t, gadminOpt2.Validate(c))
	assert.False(t, c.RequireAny(gadminOpt1, gadminOpt2))
	assert.False(t, c.CheckTargetAdmin(g2, test.UID1))

	assert.NotNil(t, c.UngrantTargetRole(g2, test.UID1, RoleType(-1)))
	assert.NotNil(t, c.UngrantTargetRole(g2, test.UID1, TargetAdmin))

	ids = c.ListTargetAdmin(g2)
	assert.Empty(t, ids)
}

func TestStateManager_CheckGroupCommandPermission(t *testing.T) {
	test.InitBuntdb(t)
	defer test.CloseBuntdb(t)
	c := initStateManager(t)

	opt := QQAdminRequireOption(g1, test.UID1)
	assert.False(t, opt.Validate(c))

	gcadminOpt1 := TargetCommandRequireOption(g1, test.UID1, test.CMD1)
	gcadminOpt2 := TargetCommandRequireOption(g1, test.UID1, test.CMD2)
	assert.False(t, gcadminOpt1.Validate(c))
	assert.False(t, gcadminOpt2.Validate(c))
	assert.False(t, c.RequireAny(gcadminOpt1, gcadminOpt2))

	assert.Nil(t, c.TargetGrantPermission(g1, test.UID1, test.CMD2))
	assert.False(t, gcadminOpt1.Validate(c))
	assert.True(t, gcadminOpt2.Validate(c))
	assert.True(t, c.RequireAny(gcadminOpt1, gcadminOpt2))

	assert.NotNil(t, c.TargetGrantPermission(g1, test.UID1, test.CMD2))

	assert.Nil(t, c.UngrantPermission(g1, test.UID1, test.CMD2))
	assert.False(t, gcadminOpt1.Validate(c))
	assert.False(t, gcadminOpt2.Validate(c))
	assert.False(t, c.RequireAny(gcadminOpt1, gcadminOpt2))

	assert.NotNil(t, c.UngrantPermission(g1, test.UID1, test.CMD2))
}

func TestStateManager_RemoveAllByGroup(t *testing.T) {
	test.InitBuntdb(t)
	defer test.CloseBuntdb(t)
	c := initStateManager(t)

	assert.Nil(t, c.CreatePatternIndex(c.TargetPermissionKey, []interface{}{g1}))
	assert.Nil(t, c.CreatePatternIndex(c.TargetPermissionKey, []interface{}{g2}))

	assert.Nil(t, c.GrantTargetRole(g1, test.UID1, TargetAdmin))
	assert.Nil(t, c.GrantTargetRole(g2, test.UID1, TargetAdmin))

	gadminOpt1 := TargetAdminRoleRequireOption(g1, test.UID1)
	// 当发现group command 应该用GroupPermissionKey的时候已经太晚了
	//gcadminOpt1 := TargetCommandRequireOption(g1, test.UID1, test.CMD1)
	gcadminOpt2 := TargetAdminRoleRequireOption(g2, test.UID1)

	assert.True(t, gadminOpt1.Validate(c))
	//assert.True(t, gcadminOpt1.Validate(c))
	assert.True(t, gcadminOpt2.Validate(c))

	_, err := c.RemoveAllByTarget(g1)
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

	assert.Nil(t, c.GrantTargetRole(g1, test.UID1, TargetAdmin))
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

	assert.False(t, c.CheckTargetSilence(g1))
	assert.False(t, c.CheckTargetSilence(g2))

	assert.Nil(t, c.TargetSilence(g1))

	assert.True(t, c.CheckTargetSilence(g1))
	assert.False(t, c.CheckTargetSilence(g2))

	assert.Nil(t, c.UndoTargetSilence(g1))

	assert.False(t, c.CheckTargetSilence(g1))
	assert.False(t, c.CheckTargetSilence(g2))

	assert.Nil(t, c.GlobalSilence())
	assert.True(t, c.CheckTargetSilence(g1))
	assert.True(t, c.CheckTargetSilence(g2))

	assert.NotNil(t, c.TargetSilence(g1))
	assert.NotNil(t, c.UndoTargetSilence(g1))
}
