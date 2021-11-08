package lsp

import (
	"fmt"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Sora233/DDBOT/internal/test"
	"github.com/Sora233/DDBOT/lsp/bilibili"
	"github.com/Sora233/DDBOT/lsp/concern"
	"github.com/Sora233/DDBOT/lsp/mmsg"
	"github.com/Sora233/DDBOT/lsp/permission"
	"github.com/Sora233/DDBOT/utils/msgstringer"
	"github.com/stretchr/testify/assert"
	"testing"

	_ "github.com/Sora233/DDBOT/lsp/acfun"
	_ "github.com/Sora233/DDBOT/lsp/bilibili"
	_ "github.com/Sora233/DDBOT/lsp/douyu"
	_ "github.com/Sora233/DDBOT/lsp/huya"
	_ "github.com/Sora233/DDBOT/lsp/weibo"
	_ "github.com/Sora233/DDBOT/lsp/youtube"
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
	ctx.GlobalDisabledReply = func() interface{} {
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

	const bsite = "bilibili"

	// 没初始化，应该报错
	IList(ctx, test.G1, bsite)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToMessage(target).Elements), "订阅查询失败")

	cm := getCM(bsite)
	assert.NotNil(t, cm)
	sm := cm.GetStateManager().(*bilibili.StateManager)
	assert.NotNil(t, sm)
	sm.FreshIndex(test.G1, test.G2)

	IList(ctx, test.G1, bsite)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToMessage(target).Elements), "暂无订阅")

	err := sm.AddUserInfo(&bilibili.UserInfo{
		Mid:  test.UID1,
		Name: test.NAME1,
	})
	assert.Nil(t, err)
	_, err = sm.AddGroupConcern(test.G1, test.UID1, test.BibiliLive)
	assert.Nil(t, err)

	err = sm.AddUserInfo(&bilibili.UserInfo{
		Mid:  test.UID2,
		Name: test.NAME2,
	})
	assert.Nil(t, err)
	_, err = sm.AddGroupConcern(test.G1, test.UID2, test.BilibiliNews)
	assert.Nil(t, err)

	IList(ctx, test.G1, bsite)
	result = <-msgChan

	assert.Contains(t, msgstringer.MsgToString(result.ToMessage(target).Elements),
		fmt.Sprintf("%v %v %v", test.NAME1, test.UID1, test.BibiliLive))

	assert.Contains(t, msgstringer.MsgToString(result.ToMessage(target).Elements),
		fmt.Sprintf("%v %v %v", test.NAME2, test.UID2, test.BilibiliNews))
}
