package lsp

import (
	"context"
	"fmt"
	"github.com/Mrs4s/MiraiGo/client"
	"github.com/Sora233/DDBOT/internal/test"
	tc "github.com/Sora233/DDBOT/internal/test_concern"
	"github.com/Sora233/DDBOT/lsp/concern"
	"github.com/Sora233/DDBOT/lsp/concern_type"
	"github.com/Sora233/DDBOT/lsp/mmsg"
	"github.com/Sora233/DDBOT/lsp/mmsg/mt"
	"github.com/Sora233/DDBOT/lsp/permission"
	localutils "github.com/Sora233/DDBOT/utils"
	"github.com/Sora233/DDBOT/utils/msgstringer"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
	"time"
)

const (
	noPermission   = "no permission"
	globalDisabled = "global disabled"
	disabled       = "disabled"
	success        = "成功"
	failed         = "失败"
)

var (
	Sender1 = &sender{test.UID1, test.NAME1}
	Sender2 = &sender{test.UID2, test.NAME2}
	g1      = mt.NewGroupTarget(test.G1)
	g2      = mt.NewGroupTarget(test.G2)
)

func initLsp(t *testing.T) {
	test.InitMirai()
	test.InitBuntdb(t)
	Instance.PermissionStateManager = permission.NewStateManager()
	Instance.LspStateManager = NewStateManager()
}

func closeLsp(t *testing.T) {
	localutils.GetBot().TESTReset()
	concern.ClearConcern()
	test.CloseBuntdb(t)
	test.CloseMirai()
}

func NewCtx(t *testing.T, receiver chan<- *mmsg.MSG, sender mmsg.MessageSender, source mt.TargetType) *MessageContext {
	ctx := NewMessageContext()
	ctx.Lsp = Instance
	ctx.Log = logger.WithField("test", "test")
	ctx.Source = source
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
	assert.EqualValues(t, sender, ctx.GetSender())
	assert.EqualValues(t, source, ctx.GetSource())
	ctx.IsFromPrivate()
	ctx.IsFromGroup()
	ctx.GetLog()
	return ctx
}

func getCM(site string) concern.Concern {
	for _, cm := range concern.ListConcern() {
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

func newTestConcern(t *testing.T, e chan concern.Event, n chan<- concern.Notify, site string, ctypes []concern_type.Type) *tc.TestConcern {
	c := tc.NewTestConcern(n, site, ctypes)
	c.UseFreshFunc(testFresh(e))
	assert.Nil(t, c.Start())
	return c
}

func TestIList(t *testing.T) {
	initLsp(t)
	defer closeLsp(t)

	msgChan := make(chan *mmsg.MSG, 10)
	target := mt.NewGroupTarget(test.G1)
	ctx := NewCtx(t, msgChan, Sender1, mt.TargetGroup)

	assert.Nil(t, Instance.PermissionStateManager.DisableTargetCommand(g2, ListCommand))

	IList(ctx, g1, "xxx")
	result := <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), failed)

	assert.Nil(t, Instance.PermissionStateManager.DisableTargetCommand(g1, ListCommand))

	IList(ctx, g2, "xxx")
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), disabled)

	assert.Nil(t, Instance.PermissionStateManager.EnableTargetCommand(g1, ListCommand))
	assert.Nil(t, Instance.PermissionStateManager.GlobalDisableCommand(ListCommand))

	IList(ctx, g2, "xxx")
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), disabled)
	assert.Nil(t, Instance.PermissionStateManager.GlobalEnableCommand(ListCommand))

	testEventChan := make(chan concern.Event, 16)
	testNotifyChan := make(chan concern.Notify, 1)

	tc1 := newTestConcern(t, testEventChan, testNotifyChan, test.Site1, []concern_type.Type{test.T1})
	concern.RegisterConcern(tc1)
	defer tc1.Stop()

	tc2 := newTestConcern(t, testEventChan, testNotifyChan, test.Site2, []concern_type.Type{test.T2})
	concern.RegisterConcern(tc2)
	defer tc2.Stop()

	defer close(testNotifyChan)
	assert.Len(t, concern.ListSite(), 2)
	assert.Contains(t, concern.ListSite(), test.Site1)
	assert.Contains(t, concern.ListSite(), test.Site2)

	IList(ctx, g1, "")
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), "暂无订阅")

	_, err := tc1.GetStateManager().AddTargetConcern(g1, test.NAME1, test.T1)
	assert.Nil(t, err)
	_, err = tc2.GetStateManager().AddTargetConcern(g1, test.NAME2, test.T2)
	assert.Nil(t, err)

	IList(ctx, g1, "")
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements),
		fmt.Sprintf("%v %v %v", test.NAME1, test.NAME1, test.T1))
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements),
		fmt.Sprintf("%v %v %v", test.NAME2, test.NAME2, test.T2))

	_, err = tc1.GetStateManager().AddTargetConcern(g2, test.NAME1, test.T1)
	assert.Nil(t, err)

	assert.Nil(t, Instance.PermissionStateManager.EnableTargetCommand(g2, ListCommand))
	IList(ctx, g2, "")
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements),
		fmt.Sprintf("%v %v %v", test.NAME1, test.NAME1, test.T1))
	assert.NotContains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), test.NAME2)

	IList(ctx, g1, tc1.Site())
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements),
		fmt.Sprintf("%v %v %v", test.NAME1, test.NAME1, test.T1))
	assert.NotContains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), test.NAME2)
}

func TestIEnable(t *testing.T) {
	initLsp(t)
	defer closeLsp(t)

	msgChan := make(chan *mmsg.MSG, 10)
	target := mt.NewGroupTarget(test.G1)
	ctx := NewCtx(t, msgChan, Sender1, target.GetTargetType())

	IEnable(ctx, g1, WatchCommand, false)
	result := <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), noPermission)

	assert.Nil(t, Instance.PermissionStateManager.GrantRole(Sender1.Uin(), permission.Admin))

	IEnable(ctx, g1, "", false)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), failed)

	IEnable(ctx, g1, "???", false)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), failed)

	IEnable(ctx, g1, EnableCommand, false)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), failed)

	IEnable(ctx, g1, WatchCommand, false)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), success)

	IEnable(ctx, g1, WatchCommand, false)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), failed)

	assert.Nil(t, Instance.PermissionStateManager.GlobalDisableCommand(WatchCommand))

	IEnable(ctx, g1, WatchCommand, false)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), globalDisabled)

	assert.Nil(t, Instance.PermissionStateManager.GlobalEnableCommand(WatchCommand))

	IEnable(ctx, g1, WatchCommand, true)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), success)

	IEnable(ctx, g1, WatchCommand, true)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), failed)
}

func TestIGrantRole(t *testing.T) {
	initLsp(t)
	defer closeLsp(t)

	msgChan := make(chan *mmsg.MSG, 10)
	target := mt.NewGroupTarget(test.G1)
	ctx := NewCtx(t, msgChan, Sender1, target.GetTargetType())

	IGrantRole(ctx, g1, permission.TargetAdmin, test.UID2, false)
	result := <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), noPermission)

	assert.Nil(t, Instance.PermissionStateManager.GrantTargetRole(g1, Sender1.Uin(), permission.TargetAdmin))

	IGrantRole(ctx, g1, permission.RoleType(-1), test.UID2, false)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), "invalid role")

	IGrantRole(ctx, g1, permission.TargetAdmin, test.UID2, false)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), "未找到用户")

	localutils.GetBot().TESTAddMember(test.G1, test.UID2, client.Member)

	IGrantRole(ctx, g1, permission.TargetAdmin, test.UID2, false)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), success)

	IGrantRole(ctx, g1, permission.TargetAdmin, test.UID2, false)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), "失败 - 目标已有该权限")

	IGrantRole(ctx, g1, permission.TargetAdmin, test.UID2, true)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), success)

	IGrantRole(ctx, g1, permission.TargetAdmin, test.UID2, true)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), "失败 - 目标未有该权限")

	IGrantRole(ctx, mt.NewGroupTarget(0), permission.Admin, test.UID2, false)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), noPermission)

	assert.Nil(t, Instance.PermissionStateManager.GrantRole(Sender1.Uin(), permission.Admin))

	IGrantRole(ctx, mt.NewGroupTarget(0), permission.Admin, test.UID2, false)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), success)

	IGrantRole(ctx, mt.NewGroupTarget(0), permission.Admin, test.UID2, false)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), "失败 - 目标已有该权限")

	IGrantRole(ctx, mt.NewGroupTarget(0), permission.Admin, test.UID2, true)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), success)

	IGrantRole(ctx, mt.NewGroupTarget(0), permission.Admin, test.UID2, true)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), "失败 - 目标未有该权限")
}

func TestIGrantCmd(t *testing.T) {
	initLsp(t)
	defer closeLsp(t)

	msgChan := make(chan *mmsg.MSG, 10)
	target := mt.NewGroupTarget(test.G1)
	ctx := NewCtx(t, msgChan, Sender1, target.GetTargetType())

	IGrantCmd(ctx, g1, "", Sender2.Uin(), false)
	result := <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), noPermission)

	assert.Nil(t, Instance.PermissionStateManager.GrantTargetRole(g1, Sender1.Uin(), permission.TargetAdmin))

	IGrantCmd(ctx, g1, "", Sender2.Uin(), false)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), failed)

	IGrantCmd(ctx, g1, WatchCommand, Sender2.Uin(), false)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), "未找到用户")

	localutils.GetBot().TESTAddMember(test.G1, Sender2.Uin(), client.Member)

	IGrantCmd(ctx, g1, WatchCommand, Sender2.Uin(), false)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), success)

	IGrantCmd(ctx, g1, WatchCommand, Sender2.Uin(), false)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), failed)

	IGrantCmd(ctx, g1, WatchCommand, Sender2.Uin(), true)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), success)

	IGrantCmd(ctx, g1, WatchCommand, Sender2.Uin(), true)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), failed)

	assert.Nil(t, Instance.PermissionStateManager.GlobalDisableCommand(WatchCommand))

	IGrantCmd(ctx, g1, WatchCommand, Sender2.Uin(), true)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), globalDisabled)

	IGrantCmd(ctx, g1, WatchCommand, Sender2.Uin(), false)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), globalDisabled)
}

func TestISilenceCmd(t *testing.T) {
	initLsp(t)
	defer closeLsp(t)

	msgChan := make(chan *mmsg.MSG, 10)
	target := mt.NewGroupTarget(test.G1)
	ctx := NewCtx(t, msgChan, Sender1, target.GetTargetType())

	ISilenceCmd(ctx, mt.NewGroupTarget(0), true, false)
	result := <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), noPermission)

	assert.Nil(t, Instance.PermissionStateManager.GrantTargetRole(g1, Sender1.Uin(), permission.TargetAdmin))

	ISilenceCmd(ctx, mt.NewGroupTarget(0), true, false)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), noPermission)
	ISilenceCmd(ctx, g2, false, false)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), noPermission)

	ISilenceCmd(ctx, g1, false, false)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), success)
	ISilenceCmd(ctx, g1, false, false)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), success)

	ISilenceCmd(ctx, g1, false, true)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), success)
	ISilenceCmd(ctx, g1, false, true)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), success)

	assert.Nil(t, Instance.PermissionStateManager.GrantRole(Sender1.Uin(), permission.Admin))

	ISilenceCmd(ctx, mt.NewGroupTarget(0), true, false)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), success)

	ISilenceCmd(ctx, mt.NewGroupTarget(0), true, true)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), success)

	ISilenceCmd(ctx, mt.NewGroupTarget(0), true, false)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), success)

	ISilenceCmd(ctx, g1, false, true)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), failed)

	ISilenceCmd(ctx, g1, false, false)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), failed)

}

func TestIWatch(t *testing.T) {
	initLsp(t)
	defer closeLsp(t)

	testEventChan1 := make(chan concern.Event, 16)
	testEventChan2 := make(chan concern.Event, 16)
	testNotifyChan := make(chan concern.Notify, 1)
	defer close(testNotifyChan)

	var err error
	var result *mmsg.MSG
	msgChan := make(chan *mmsg.MSG, 10)
	target := mt.NewGroupTarget(test.G1)
	ctx := NewCtx(t, msgChan, Sender1, target.GetTargetType())

	IWatch(ctx, g1, test.NAME1, test.Site1, test.T1, false)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), noPermission)

	err = Instance.PermissionStateManager.GrantRole(Sender1.Uin(), permission.Admin)
	assert.Nil(t, err)
	assert.Nil(t, Instance.PermissionStateManager.DisableTargetCommand(g1, WatchCommand))

	IWatch(ctx, g1, test.NAME1, test.Site1, test.T1, false)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), disabled)

	assert.Nil(t, Instance.PermissionStateManager.EnableTargetCommand(g1, WatchCommand))

	IWatch(ctx, g1, test.NAME1, test.Site1, test.T1, false)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), failed)

	tc1 := newTestConcern(t, testEventChan1, testNotifyChan, test.Site1, []concern_type.Type{test.T1})
	concern.RegisterConcern(tc1)
	defer tc1.Stop()

	tc2 := newTestConcern(t, testEventChan2, testNotifyChan, test.Site2, []concern_type.Type{test.T2})
	concern.RegisterConcern(tc2)
	defer tc2.Stop()

	IWatch(ctx, g1, test.NAME1, test.Site1, test.T1, false)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), success)

	IWatch(ctx, g1, test.NAME1, test.Site1, test.T1, false)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), failed)

	IWatch(ctx, g1, test.NAME1, test.Site1, test.T1, true)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), success)

	IWatch(ctx, g1, test.NAME1, test.Site1, test.T1, true)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), failed)

	testEventChan1 <- tc1.NewTestEvent(test.T1, mt.NewGroupTarget(0), test.NAME1)

	select {
	case <-testNotifyChan:
		assert.Fail(t, "should no item received")
	case <-time.After(time.Second):
	}

	IWatch(ctx, g1, test.NAME1, test.Site1, test.T1, false)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), success)

	IWatch(ctx, g2, test.NAME1, test.Site1, test.T1, false)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), success)

	testEventChan1 <- tc1.NewTestEvent(test.T1, mt.NewGroupTarget(0), test.NAME1)

	for i := 0; i < 2; i++ {
		select {
		case notify := <-testNotifyChan:
			assert.EqualValues(t, test.NAME1, notify.GetUid())
			assert.EqualValues(t, test.Site1, notify.Site())
			assert.True(t, notify.GetTarget().Equal(g1) || notify.GetTarget().Equal(g2))
		case <-time.After(time.Second):
			assert.Fail(t, "no item received")
		}
	}
}

func TestIConfigAtCmd(t *testing.T) {
	initLsp(t)
	defer closeLsp(t)

	testEventChan1 := make(chan concern.Event, 16)
	testEventChan2 := make(chan concern.Event, 16)
	testNotifyChan := make(chan concern.Notify, 1)
	defer close(testNotifyChan)

	var result *mmsg.MSG
	var err error
	msgChan := make(chan *mmsg.MSG, 10)
	target := mt.NewGroupTarget(test.G1)
	ctx := NewCtx(t, msgChan, Sender1, target.GetTargetType())

	tc1 := newTestConcern(t, testEventChan1, testNotifyChan, test.Site1, []concern_type.Type{test.T1})
	concern.RegisterConcern(tc1)
	defer tc1.Stop()

	tc2 := newTestConcern(t, testEventChan2, testNotifyChan, test.Site2, []concern_type.Type{test.T2})
	concern.RegisterConcern(tc2)
	defer tc2.Stop()

	IConfigAtCmd(ctx, g1, test.NAME1, test.Site1, test.T1, "show", nil)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), noPermission)

	err = Instance.PermissionStateManager.GrantRole(Sender1.Uin(), permission.Admin)
	assert.Nil(t, err)
	assert.Nil(t, Instance.PermissionStateManager.DisableTargetCommand(g1, ConfigCommand))

	IConfigAtCmd(ctx, g1, test.NAME1, test.Site1, test.T1, "show", nil)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), disabled)

	assert.Nil(t, Instance.PermissionStateManager.EnableTargetCommand(g1, ConfigCommand))

	IWatch(ctx, g1, test.NAME1, test.Site1, test.T1, false)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), success)

	IConfigAtCmd(ctx, g1, test.NAME1, test.Site1, test.T1, "show", nil)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), "当前配置为空")

	IConfigAtCmd(ctx, g1, test.NAME1, test.Site1, test.T1, "add", nil)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), failed)

	IConfigAtCmd(ctx, g1, test.NAME1, test.Site1, test.T1, "add", []int64{test.UID1, test.UID2})
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), failed)

	localutils.GetBot().TESTAddGroup(test.G1)
	localutils.GetBot().TESTAddMember(test.G1, test.UID1, client.Member)

	IConfigAtCmd(ctx, g1, test.NAME1, test.Site1, test.T1, "add", []int64{test.UID1, test.UID2})
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), failed)

	localutils.GetBot().TESTAddMember(test.G1, test.UID2, client.Member)

	IConfigAtCmd(ctx, g1, test.NAME1, test.Site1, test.T1, "add", []int64{test.UID1, test.UID2})
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), success)

	IConfigAtCmd(ctx, g1, test.NAME1, test.Site1, test.T1, "show", nil)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), strconv.FormatInt(test.UID1, 10))
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), strconv.FormatInt(test.UID2, 10))

	IConfigAtCmd(ctx, g1, test.NAME1, test.Site1, test.T1, "remove", []int64{test.UID1, test.UID3})
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), success)

	IConfigAtCmd(ctx, g1, test.NAME1, test.Site1, test.T1, "show", nil)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), strconv.FormatInt(test.UID2, 10))
	assert.NotContains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), strconv.FormatInt(test.UID1, 10))
	assert.NotContains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), strconv.FormatInt(test.UID3, 10))

	IConfigAtCmd(ctx, g1, test.NAME1, test.Site1, test.T1, "clear", nil)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), success)

	IConfigAtCmd(ctx, g1, test.NAME1, test.Site1, test.T1, "show", nil)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), "当前配置为空")

	IConfigAtCmd(ctx, g1, test.NAME1, test.Site1, test.T1, "unknown", []int64{test.UID1})
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), failed)
}

func TestIConfigAtAllCmd(t *testing.T) {
	initLsp(t)
	defer closeLsp(t)

	testEventChan1 := make(chan concern.Event, 16)
	testEventChan2 := make(chan concern.Event, 16)
	testNotifyChan := make(chan concern.Notify, 1)
	defer close(testNotifyChan)

	var result *mmsg.MSG
	var err error
	msgChan := make(chan *mmsg.MSG, 10)
	target := mt.NewGroupTarget(test.G1)
	ctx := NewCtx(t, msgChan, Sender1, target.GetTargetType())

	tc1 := newTestConcern(t, testEventChan1, testNotifyChan, test.Site1, []concern_type.Type{test.T1})
	concern.RegisterConcern(tc1)
	defer tc1.Stop()

	tc2 := newTestConcern(t, testEventChan2, testNotifyChan, test.Site2, []concern_type.Type{test.T2})
	concern.RegisterConcern(tc2)

	IConfigAtAllCmd(ctx, g1, test.NAME1, test.Site1, test.T1, true)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), noPermission)

	err = Instance.PermissionStateManager.GrantRole(Sender1.Uin(), permission.Admin)
	assert.Nil(t, err)
	assert.Nil(t, Instance.PermissionStateManager.DisableTargetCommand(g1, ConfigCommand))

	IConfigAtAllCmd(ctx, g1, test.NAME1, test.Site1, test.T1, true)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), disabled)

	assert.Nil(t, Instance.PermissionStateManager.EnableTargetCommand(g1, ConfigCommand))

	IConfigAtAllCmd(ctx, g1, test.NAME1, test.Site1, test.T1, true)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), failed)

	IWatch(ctx, g1, test.NAME1, test.Site1, test.T1, false)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), success)

	IConfigAtAllCmd(ctx, g1, test.NAME1, test.Site1, test.T1, true)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), success)

	IConfigAtAllCmd(ctx, g1, test.NAME1, test.Site1, test.T1, true)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), failed)

	IConfigAtAllCmd(ctx, g1, test.NAME1, test.Site1, test.T1, false)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), success)

	IConfigAtAllCmd(ctx, g1, test.NAME1, test.Site1, test.T1, false)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), failed)
}

func TestIConfigTitleNotifyCmd(t *testing.T) {
	initLsp(t)
	defer closeLsp(t)

	testEventChan1 := make(chan concern.Event, 16)
	testEventChan2 := make(chan concern.Event, 16)
	testNotifyChan := make(chan concern.Notify, 1)
	defer close(testNotifyChan)

	var result *mmsg.MSG
	var err error
	msgChan := make(chan *mmsg.MSG, 10)
	target := mt.NewGroupTarget(test.G1)
	ctx := NewCtx(t, msgChan, Sender1, target.GetTargetType())

	tc1 := newTestConcern(t, testEventChan1, testNotifyChan, test.Site1, []concern_type.Type{test.T1})
	concern.RegisterConcern(tc1)
	defer tc1.Stop()

	tc2 := newTestConcern(t, testEventChan2, testNotifyChan, test.Site2, []concern_type.Type{test.T2})
	concern.RegisterConcern(tc2)

	IConfigTitleNotifyCmd(ctx, g1, test.NAME1, test.Site1, test.T1, true)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), noPermission)

	err = Instance.PermissionStateManager.GrantRole(Sender1.Uin(), permission.Admin)
	assert.Nil(t, err)
	assert.Nil(t, Instance.PermissionStateManager.DisableTargetCommand(g1, ConfigCommand))

	IConfigTitleNotifyCmd(ctx, g1, test.NAME1, test.Site1, test.T1, true)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), disabled)

	assert.Nil(t, Instance.PermissionStateManager.EnableTargetCommand(g1, ConfigCommand))

	IConfigTitleNotifyCmd(ctx, g1, test.NAME1, test.Site1, test.T1, true)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), failed)

	IWatch(ctx, g1, test.NAME1, test.Site1, test.T1, false)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), success)

	IConfigTitleNotifyCmd(ctx, g1, test.NAME1, test.Site1, test.T1, true)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), success)

	IConfigTitleNotifyCmd(ctx, g1, test.NAME1, test.Site1, test.T1, true)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), failed)

	IConfigTitleNotifyCmd(ctx, g1, test.NAME1, test.Site1, test.T1, false)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), success)

	IConfigTitleNotifyCmd(ctx, g1, test.NAME1, test.Site1, test.T1, false)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), failed)
}

func TestIConfigOfflineNotifyCmd(t *testing.T) {
	initLsp(t)
	defer closeLsp(t)

	testEventChan1 := make(chan concern.Event, 16)
	testEventChan2 := make(chan concern.Event, 16)
	testNotifyChan := make(chan concern.Notify, 1)
	defer close(testNotifyChan)

	var result *mmsg.MSG
	var err error
	msgChan := make(chan *mmsg.MSG, 10)
	target := mt.NewGroupTarget(test.G1)
	ctx := NewCtx(t, msgChan, Sender1, target.GetTargetType())

	tc1 := newTestConcern(t, testEventChan1, testNotifyChan, test.Site1, []concern_type.Type{test.T1})
	concern.RegisterConcern(tc1)
	defer tc1.Stop()

	tc2 := newTestConcern(t, testEventChan2, testNotifyChan, test.Site2, []concern_type.Type{test.T2})
	concern.RegisterConcern(tc2)

	IConfigOfflineNotifyCmd(ctx, g1, test.NAME1, test.Site1, test.T1, true)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), noPermission)

	err = Instance.PermissionStateManager.GrantRole(Sender1.Uin(), permission.Admin)
	assert.Nil(t, err)
	assert.Nil(t, Instance.PermissionStateManager.DisableTargetCommand(g1, ConfigCommand))

	IConfigOfflineNotifyCmd(ctx, g1, test.NAME1, test.Site1, test.T1, true)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), disabled)

	assert.Nil(t, Instance.PermissionStateManager.EnableTargetCommand(g1, ConfigCommand))

	IConfigOfflineNotifyCmd(ctx, g1, test.NAME1, test.Site1, test.T1, true)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), failed)

	IWatch(ctx, g1, test.NAME1, test.Site1, test.T1, false)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), success)

	IConfigOfflineNotifyCmd(ctx, g1, test.NAME1, test.Site1, test.T1, true)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), success)

	IConfigOfflineNotifyCmd(ctx, g1, test.NAME1, test.Site1, test.T1, true)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), failed)

	IConfigOfflineNotifyCmd(ctx, g1, test.NAME1, test.Site1, test.T1, false)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), success)

	IConfigOfflineNotifyCmd(ctx, g1, test.NAME1, test.Site1, test.T1, false)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), failed)
}

func TestIConfigFilterCmd(t *testing.T) {
	initLsp(t)
	defer closeLsp(t)

	testEventChan1 := make(chan concern.Event, 16)
	testEventChan2 := make(chan concern.Event, 16)
	testNotifyChan := make(chan concern.Notify, 1)
	defer close(testNotifyChan)

	var result *mmsg.MSG
	var err error
	msgChan := make(chan *mmsg.MSG, 10)
	target := mt.NewGroupTarget(test.G1)
	ctx := NewCtx(t, msgChan, Sender1, target.GetTargetType())

	tc1 := newTestConcern(t, testEventChan1, testNotifyChan, test.Site1, []concern_type.Type{test.T1})
	concern.RegisterConcern(tc1)
	defer tc1.Stop()

	tc2 := newTestConcern(t, testEventChan2, testNotifyChan, test.Site2, []concern_type.Type{test.T2})
	concern.RegisterConcern(tc2)

	IConfigFilterCmdType(ctx, g1, test.NAME1, test.Site1, test.T1, []string{test.Type1})
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), noPermission)

	err = Instance.PermissionStateManager.GrantRole(Sender1.Uin(), permission.Admin)
	assert.Nil(t, err)
	assert.Nil(t, Instance.PermissionStateManager.DisableTargetCommand(g1, ConfigCommand))

	IConfigFilterCmdType(ctx, g1, test.NAME1, test.Site1, test.T1, []string{test.Type1})
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), disabled)

	assert.Nil(t, Instance.PermissionStateManager.EnableTargetCommand(g1, ConfigCommand))

	IConfigFilterCmdType(ctx, g1, test.NAME1, test.Site1, test.T1, []string{test.Type1})
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), failed)

	IWatch(ctx, g1, test.NAME1, test.Site1, test.T1, false)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), success)

	IConfigFilterCmdType(ctx, g1, test.NAME1, test.Site1, test.T1, []string{})
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), failed)

	IConfigFilterCmdType(ctx, g1, test.NAME1, test.Site1, test.T1, []string{test.Type1})
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), failed)

	IConfigFilterCmdNotType(ctx, g1, test.NAME1, test.Site1, test.T1, []string{})
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), failed)

	IConfigFilterCmdNotType(ctx, g1, test.NAME1, test.Site1, test.T1, []string{test.Type1})
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), failed)

	IConfigFilterCmdText(ctx, g1, test.NAME1, test.Site1, test.T1, []string{})
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), failed)

	IConfigFilterCmdText(ctx, g1, test.NAME1, test.Site1, test.T1, []string{test.NAME1, test.NAME2})
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), success)

	IConfigFilterCmdShow(ctx, g1, test.NAME1, test.Site1, test.T1)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), "关键字过滤模式")
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), test.NAME1)
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), test.NAME2)

	IConfigFilterCmdClear(ctx, g1, test.NAME1, test.Site1, test.T1)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), success)

	IConfigFilterCmdShow(ctx, g1, test.NAME1, test.Site1, test.T1)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), "当前配置为空")
}
