package lsp

import (
	"context"
	"fmt"
	"github.com/Mrs4s/MiraiGo/client"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Sora233/DDBOT/internal/test"
	tc "github.com/Sora233/DDBOT/internal/test_concern"
	"github.com/Sora233/DDBOT/lsp/concern"
	"github.com/Sora233/DDBOT/lsp/concern_type"
	"github.com/Sora233/DDBOT/lsp/mmsg"
	"github.com/Sora233/DDBOT/lsp/permission"
	localutils "github.com/Sora233/DDBOT/utils"
	"github.com/Sora233/DDBOT/utils/msgstringer"
	"github.com/stretchr/testify/assert"
	"testing"
)

const (
	noPermission   = "no permission"
	globalDisabled = "global disabled"
	disabled       = "disabled"
)

func initLsp(t *testing.T) {
	test.InitMirai()
	test.InitBuntdb(t)
	Instance.commandPrefix = "/"
	Instance.PermissionStateManager = permission.NewStateManager()
	Instance.LspStateManager = NewStateManager()
}

func closeLsp(t *testing.T) {
	localutils.GetBot().TESTClear()
	concern.ClearConcern()
	test.CloseBuntdb(t)
	test.CloseMirai()
}

func NewCtx(receiver chan<- *mmsg.MSG, sender *message.Sender, target mmsg.Target) *MessageContext {
	ctx := NewMessageContext()
	ctx.Lsp = Instance
	ctx.Log = logger.WithField("test", "test")
	ctx.Target = target
	ctx.Sender = sender
	ctx.SendFunc = func(m *mmsg.MSG) interface{} {
		if receiver != nil {
			receiver <- m
		}
		return m
	}
	ctx.ReplyFunc = ctx.SendFunc
	ctx.DisabledReply = func() interface{} {
		return ctx.Send(mmsg.NewTextf(disabled))
	}
	ctx.GlobalDisabledReply = func() interface{} {
		return ctx.Send(mmsg.NewTextf(globalDisabled))
	}
	ctx.NoPermissionReplyFunc = func() interface{} {
		return ctx.Send(mmsg.NewTextf(noPermission))
	}
	return ctx
}

func getCM(site string) concern.Concern {
	for _, cm := range concern.ListConcernManager() {
		if cm.Site() == site {
			return cm
		}
	}
	return nil
}

func testFresh(testEventChan <-chan concern.Event) concern.FreshFunc {
	return func(ctx context.Context, eventChan chan<- concern.Event) {
		for {
			select {
			case e := <-testEventChan:
				if e != nil {
					eventChan <- e
				}
			case <-ctx.Done():
				return
			}
		}
	}
}

func newTestConcern(t *testing.T, e chan concern.Event, n chan concern.Notify, site string, ctypes []concern_type.Type) *tc.TestConcern {
	c := tc.NewTestConcern(n, site, ctypes)
	c.UseFreshFunc(testFresh(e))
	assert.Nil(t, c.Start())
	return c
}

func TestIList(t *testing.T) {
	initLsp(t)
	defer closeLsp(t)

	msgChan := make(chan *mmsg.MSG, 10)
	target := mmsg.NewGroupTarget(test.G1)
	ctx := NewCtx(msgChan, test.Sender1, target)

	assert.Nil(t, Instance.PermissionStateManager.DisableGroupCommand(test.G2, ListCommand))

	IList(ctx, test.G1, "xxx")
	result := <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToMessage(target).Elements), "失败")

	assert.Nil(t, Instance.PermissionStateManager.DisableGroupCommand(test.G1, ListCommand))

	IList(ctx, test.G1, "xxx")
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToMessage(target).Elements), disabled)

	assert.Nil(t, Instance.PermissionStateManager.EnableGroupCommand(test.G1, ListCommand))
	assert.Nil(t, Instance.PermissionStateManager.GlobalDisableGroupCommand(ListCommand))

	IList(ctx, test.G1, "xxx")
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToMessage(target).Elements), disabled)
	assert.Nil(t, Instance.PermissionStateManager.GlobalEnableGroupCommand(ListCommand))

	testEventChan := make(chan concern.Event, 16)
	testNotifyChan := make(chan concern.Notify, 1)

	tc1 := newTestConcern(t, testEventChan, testNotifyChan, test.Site1, []concern_type.Type{test.T1})
	concern.RegisterConcernManager(tc1, tc1.Ctypes)
	defer tc1.Stop()

	tc2 := newTestConcern(t, testEventChan, testNotifyChan, test.Site2, []concern_type.Type{test.T2})
	concern.RegisterConcernManager(tc2, tc2.Ctypes)
	defer tc2.Stop()

	defer close(testNotifyChan)
	assert.Len(t, concern.ListSite(), 2)
	assert.Contains(t, concern.ListSite(), test.Site1)
	assert.Contains(t, concern.ListSite(), test.Site2)

	IList(ctx, test.G1, "")
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToMessage(target).Elements), "暂无订阅")

	_, err := tc1.GetStateManager().AddGroupConcern(test.G1, test.NAME1, test.T1)
	assert.Nil(t, err)
	_, err = tc2.GetStateManager().AddGroupConcern(test.G1, test.NAME2, test.T2)
	assert.Nil(t, err)

	IList(ctx, test.G1, "")
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToMessage(target).Elements),
		fmt.Sprintf("%v %v %v", test.NAME1, test.NAME1, test.T1))
	assert.Contains(t, msgstringer.MsgToString(result.ToMessage(target).Elements),
		fmt.Sprintf("%v %v %v", test.NAME2, test.NAME2, test.T2))

	_, err = tc1.GetStateManager().AddGroupConcern(test.G2, test.NAME1, test.T1)
	assert.Nil(t, err)

	assert.Nil(t, Instance.PermissionStateManager.EnableGroupCommand(test.G2, ListCommand))
	IList(ctx, test.G2, "")
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToMessage(target).Elements),
		fmt.Sprintf("%v %v %v", test.NAME1, test.NAME1, test.T1))
	assert.NotContains(t, msgstringer.MsgToString(result.ToMessage(target).Elements), test.NAME2)

	IList(ctx, test.G1, tc1.Site())
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToMessage(target).Elements),
		fmt.Sprintf("%v %v %v", test.NAME1, test.NAME1, test.T1))
	assert.NotContains(t, msgstringer.MsgToString(result.ToMessage(target).Elements), test.NAME2)
}

func TestIEnable(t *testing.T) {
	initLsp(t)
	defer closeLsp(t)

	msgChan := make(chan *mmsg.MSG, 10)
	target := mmsg.NewGroupTarget(test.G1)
	ctx := NewCtx(msgChan, test.Sender1, target)

	IEnable(ctx, test.G1, "watch", false)
	result := <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToMessage(target).Elements), noPermission)

	assert.Nil(t, Instance.PermissionStateManager.GrantRole(test.Sender1.Uin, permission.Admin))

	IEnable(ctx, test.G1, "", false)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToMessage(target).Elements), "失败")

	IEnable(ctx, test.G1, "???", false)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToMessage(target).Elements), "失败")

	IEnable(ctx, test.G1, "enable", false)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToMessage(target).Elements), "失败")

	IEnable(ctx, test.G1, "watch", false)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToMessage(target).Elements), "成功")

	IEnable(ctx, test.G1, "watch", false)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToMessage(target).Elements), "失败")

	assert.Nil(t, Instance.PermissionStateManager.GlobalDisableGroupCommand("watch"))

	IEnable(ctx, test.G1, "watch", false)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToMessage(target).Elements), globalDisabled)

	assert.Nil(t, Instance.PermissionStateManager.GlobalEnableGroupCommand("watch"))

	IEnable(ctx, test.G1, "watch", true)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToMessage(target).Elements), "成功")

	IEnable(ctx, test.G1, "watch", true)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToMessage(target).Elements), "失败")
}

func TestIGrantRole(t *testing.T) {
	initLsp(t)
	defer closeLsp(t)

	msgChan := make(chan *mmsg.MSG, 10)
	target := mmsg.NewGroupTarget(test.G1)
	ctx := NewCtx(msgChan, test.Sender1, target)

	IGrantRole(ctx, test.G1, permission.GroupAdmin, test.UID2, false)
	result := <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToMessage(target).Elements), noPermission)

	assert.Nil(t, Instance.PermissionStateManager.GrantGroupRole(test.G1, test.Sender1.Uin, permission.GroupAdmin))

	IGrantRole(ctx, test.G1, permission.RoleType(-1), test.UID2, false)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToMessage(target).Elements), "invalid role")

	IGrantRole(ctx, test.G1, permission.GroupAdmin, test.UID2, false)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToMessage(target).Elements), "未找到用户")

	localutils.GetBot().TESTAddMember(test.G1, test.UID2, client.Member)

	IGrantRole(ctx, test.G1, permission.GroupAdmin, test.UID2, false)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToMessage(target).Elements), "成功")

	IGrantRole(ctx, test.G1, permission.GroupAdmin, test.UID2, false)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToMessage(target).Elements), "失败 - 目标已有该权限")

	IGrantRole(ctx, test.G1, permission.GroupAdmin, test.UID2, true)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToMessage(target).Elements), "成功")

	IGrantRole(ctx, test.G1, permission.GroupAdmin, test.UID2, true)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToMessage(target).Elements), "失败 - 目标未有该权限")

	IGrantRole(ctx, 0, permission.Admin, test.UID2, false)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToMessage(target).Elements), noPermission)

	assert.Nil(t, Instance.PermissionStateManager.GrantRole(test.Sender1.Uin, permission.Admin))

	IGrantRole(ctx, 0, permission.Admin, test.UID2, false)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToMessage(target).Elements), "成功")

	IGrantRole(ctx, 0, permission.Admin, test.UID2, false)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToMessage(target).Elements), "失败 - 目标已有该权限")

	IGrantRole(ctx, 0, permission.Admin, test.UID2, true)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToMessage(target).Elements), "成功")

	IGrantRole(ctx, 0, permission.Admin, test.UID2, true)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToMessage(target).Elements), "失败 - 目标未有该权限")
}

func TestIGrantCmd(t *testing.T) {
	initLsp(t)
	defer closeLsp(t)

	msgChan := make(chan *mmsg.MSG, 10)
	target := mmsg.NewGroupTarget(test.G1)
	ctx := NewCtx(msgChan, test.Sender1, target)

	IGrantCmd(ctx, test.G1, "", test.Sender2.Uin, false)
	result := <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToMessage(target).Elements), noPermission)

	assert.Nil(t, Instance.PermissionStateManager.GrantGroupRole(test.G1, test.Sender1.Uin, permission.GroupAdmin))

	IGrantCmd(ctx, test.G1, "", test.Sender2.Uin, false)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToMessage(target).Elements), "失败")

	IGrantCmd(ctx, test.G1, "watch", test.Sender2.Uin, false)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToMessage(target).Elements), "未找到用户")

	localutils.GetBot().TESTAddMember(test.G1, test.Sender2.Uin, client.Member)

	IGrantCmd(ctx, test.G1, "watch", test.Sender2.Uin, false)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToMessage(target).Elements), "成功")

	IGrantCmd(ctx, test.G1, "watch", test.Sender2.Uin, false)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToMessage(target).Elements), "失败")

	IGrantCmd(ctx, test.G1, "watch", test.Sender2.Uin, true)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToMessage(target).Elements), "成功")

	IGrantCmd(ctx, test.G1, "watch", test.Sender2.Uin, true)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToMessage(target).Elements), "失败")

	assert.Nil(t, Instance.PermissionStateManager.GlobalDisableGroupCommand("watch"))

	IGrantCmd(ctx, test.G1, "watch", test.Sender2.Uin, true)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToMessage(target).Elements), globalDisabled)

	IGrantCmd(ctx, test.G1, "watch", test.Sender2.Uin, false)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToMessage(target).Elements), globalDisabled)
}

func TestISilenceCmd(t *testing.T) {
	initLsp(t)
	defer closeLsp(t)

	msgChan := make(chan *mmsg.MSG, 10)
	target := mmsg.NewGroupTarget(test.G1)
	ctx := NewCtx(msgChan, test.Sender1, target)

	ISilenceCmd(ctx, 0, false)
	result := <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToMessage(target).Elements), noPermission)

	assert.Nil(t, Instance.PermissionStateManager.GrantGroupRole(test.G1, test.Sender1.Uin, permission.GroupAdmin))

	ISilenceCmd(ctx, 0, false)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToMessage(target).Elements), noPermission)
	ISilenceCmd(ctx, test.G2, false)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToMessage(target).Elements), noPermission)

	ISilenceCmd(ctx, test.G1, false)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToMessage(target).Elements), "成功")
	ISilenceCmd(ctx, test.G1, false)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToMessage(target).Elements), "成功")

	ISilenceCmd(ctx, test.G1, true)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToMessage(target).Elements), "成功")
	ISilenceCmd(ctx, test.G1, true)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToMessage(target).Elements), "成功")

	assert.Nil(t, Instance.PermissionStateManager.GrantRole(test.Sender1.Uin, permission.Admin))

	ISilenceCmd(ctx, 0, false)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToMessage(target).Elements), "成功")

	ISilenceCmd(ctx, 0, true)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToMessage(target).Elements), "成功")

	ISilenceCmd(ctx, 0, false)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToMessage(target).Elements), "成功")

	ISilenceCmd(ctx, test.G1, true)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToMessage(target).Elements), "失败")

	ISilenceCmd(ctx, test.G1, false)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToMessage(target).Elements), "失败")

}
