package permission

import (
	localdb "github.com/Sora233/DDBOT/lsp/buntdb"
	"github.com/Sora233/DDBOT/lsp/test"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/buntdb"
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

	assert.Nil(t, c.EnableGroupCommand(test.G1, test.CMD1, ExpireOption(time.Millisecond*100)))
	assert.True(t, c.CheckGroupCommandEnabled(test.G1, test.CMD1))
	assert.False(t, c.CheckGroupCommandDisabled(test.G1, test.CMD1))

	time.Sleep(time.Millisecond * 150)
	assert.False(t, c.CheckGroupCommandEnabled(test.G1, test.CMD1))
	assert.False(t, c.CheckGroupCommandDisabled(test.G1, test.CMD1))

	assert.Nil(t, c.DisableGroupCommand(test.G1, test.CMD1, ExpireOption(time.Millisecond*100)))
	assert.False(t, c.CheckGroupCommandEnabled(test.G1, test.CMD1))
	assert.True(t, c.CheckGroupCommandDisabled(test.G1, test.CMD1))

	time.Sleep(time.Millisecond * 150)
	assert.False(t, c.CheckGroupCommandEnabled(test.G1, test.CMD1))
	assert.False(t, c.CheckGroupCommandDisabled(test.G1, test.CMD1))
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

	assert.Nil(t, c.GlobalEnableGroupCommand(test.CMD1, ExpireOption(time.Millisecond*100)))

	time.Sleep(time.Millisecond * 150)
	assert.False(t, c.CheckGlobalCommandDisabled(test.CMD1))

	assert.Nil(t, c.GlobalDisableGroupCommand(test.CMD1))
	assert.Equal(t, ErrGlobalDisabled, c.EnableGroupCommand(test.G1, test.CMD1))
	assert.Equal(t, ErrGlobalDisabled, c.DisableGroupCommand(test.G1, test.CMD1))
	assert.Nil(t, c.GlobalEnableGroupCommand(test.CMD1))
	assert.Nil(t, c.EnableGroupCommand(test.G1, test.CMD1))
	assert.True(t, c.CheckGroupCommandEnabled(test.G1, test.CMD1))
	assert.Nil(t, c.GlobalDisableGroupCommand(test.CMD1))
	assert.False(t, c.CheckGroupCommandEnabled(test.G1, test.CMD1))
	assert.False(t, c.CheckGroupCommandDisabled(test.G1, test.CMD1))
	assert.Equal(t, ErrGlobalDisabled, c.GrantPermission(test.G1, test.UID1, test.CMD1))
	assert.Equal(t, ErrGlobalDisabled, c.UngrantPermission(test.G1, test.UID1, test.CMD1))
}

func TestStateManager_CheckRole(t *testing.T) {
	test.InitBuntdb(t)
	defer test.CloseBuntdb(t)
	c := initStateManager(t)

	adminOpt1 := AdminRoleRequireOption(test.UID1)
	adminOpt2 := AdminRoleRequireOption(test.UID2)
	assert.False(t, adminOpt1.Validate(c))
	assert.False(t, adminOpt2.Validate(c))
	assert.False(t, c.RequireAny(adminOpt1, adminOpt2))

	assert.Nil(t, c.GrantRole(test.UID2, Admin))
	assert.False(t, adminOpt1.Validate(c))
	assert.True(t, adminOpt2.Validate(c))
	assert.True(t, c.RequireAny(adminOpt1, adminOpt2))

	assert.NotNil(t, c.GrantRole(test.UID2, Admin))

	assert.Nil(t, c.UngrantRole(test.UID2, Admin))

	assert.False(t, adminOpt1.Validate(c))
	assert.False(t, adminOpt2.Validate(c))
	assert.False(t, c.RequireAny(adminOpt1, adminOpt2))

	assert.NotNil(t, c.UngrantRole(test.UID2, Admin))
}

func TestStateManager_CheckGroupRole(t *testing.T) {
	test.InitBuntdb(t)
	defer test.CloseBuntdb(t)
	c := initStateManager(t)

	gadminOpt1 := GroupAdminRoleRequireOption(test.G1, test.UID1)
	gadminOpt2 := GroupAdminRoleRequireOption(test.G2, test.UID1)
	assert.False(t, gadminOpt1.Validate(c))
	assert.False(t, gadminOpt2.Validate(c))
	assert.False(t, c.RequireAny(gadminOpt1, gadminOpt2))

	assert.Nil(t, c.GrantGroupRole(test.G2, test.UID1, GroupAdmin))
	assert.False(t, gadminOpt1.Validate(c))
	assert.True(t, gadminOpt2.Validate(c))
	assert.True(t, c.RequireAny(gadminOpt1, gadminOpt2))

	assert.NotNil(t, c.GrantGroupRole(test.G2, test.UID1, GroupAdmin))

	assert.Nil(t, c.UngrantGroupRole(test.G2, test.UID1, GroupAdmin))
	assert.False(t, gadminOpt1.Validate(c))
	assert.False(t, gadminOpt2.Validate(c))
	assert.False(t, c.RequireAny(gadminOpt1, gadminOpt2))

	assert.NotNil(t, c.UngrantGroupRole(test.G2, test.UID1, GroupAdmin))
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

	db := localdb.MustGetClient()
	db.CreateIndex(c.GroupPermissionKey(test.G1), c.GroupPermissionKey(test.G1, "*"), buntdb.IndexString)
	db.CreateIndex(c.GroupPermissionKey(test.G2), c.GroupPermissionKey(test.G2, "*"), buntdb.IndexString)

	assert.Nil(t, c.GrantGroupRole(test.G1, test.UID1, GroupAdmin))
	assert.Nil(t, c.GrantGroupRole(test.G2, test.UID1, GroupAdmin))

	gadminOpt1 := GroupAdminRoleRequireOption(test.G1, test.UID1)
	// 当发现group command 应该用GroupPermissionKey的时候已经太晚了
	//gcadminOpt1 := GroupCommandRequireOption(test.G1, test.UID1, test.CMD1)
	gcadminOpt2 := GroupAdminRoleRequireOption(test.G2, test.UID1)

	assert.True(t, gadminOpt1.Validate(c))
	//assert.True(t, gcadminOpt1.Validate(c))
	assert.True(t, gcadminOpt2.Validate(c))

	assert.Nil(t, c.RemoveAllByGroup(test.G1))

	assert.False(t, gadminOpt1.Validate(c))
	//assert.False(t, gcadminOpt1.Validate(c))
	assert.True(t, gcadminOpt2.Validate(c))
}
