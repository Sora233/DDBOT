package lsp

import (
	"testing"

	"github.com/LagrangeDev/LagrangeGo/message"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/Sora233/DDBOT/v2/internal/test"
	"github.com/Sora233/DDBOT/v2/lsp/concern"
	"github.com/Sora233/DDBOT/v2/lsp/concern_type"
	"github.com/Sora233/DDBOT/v2/lsp/mmsg"
	"github.com/Sora233/DDBOT/v2/lsp/permission"
	"github.com/Sora233/DDBOT/v2/utils/msgstringer"
)

func TestLsp_ConcernNotify(t *testing.T) {
	logrus.SetLevel(logrus.TraceLevel)
	initLsp(t)
	defer closeLsp(t)

	defer func() {
		Instance.concernNotify = concern.ReadNotifyChan()
	}()

	testEventChan := make(chan concern.Event, 16)
	testNotifyChan := make(chan concern.Notify)
	defer close(testNotifyChan)

	Instance.concernNotify = testNotifyChan

	var result *mmsg.MSG
	msgChan := make(chan *mmsg.MSG, 10)
	target := mmsg.NewGroupTarget(test.G1)
	ctx := NewCtx(t, msgChan, test.Sender1, target)

	err := Instance.PermissionStateManager.GrantRole(test.Sender1.Uin, permission.Admin)
	assert.Nil(t, err)

	tc1 := newTestConcern(t, testEventChan, testNotifyChan, test.Site1, []concern_type.Type{test.T1})
	concern.RegisterConcern(tc1)
	defer tc1.Stop()

	IWatch(ctx, test.G1, test.NAME1, test.Site1, test.T1, false)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), success)

	IWatch(ctx, test.G2, test.NAME1, test.Site1, test.T1, false)
	result = <-msgChan
	assert.Contains(t, msgstringer.MsgToString(result.ToCombineMessage(target).Elements), success)

	testEventChan <- tc1.NewTestEvent(test.T1, 0, test.NAME1)

	go Instance.ConcernNotify()
	Instance.notifyWg.Wait()
	Instance.wg.Wait()

	close(testEventChan)
}

func TestNewAtAllMsg(t *testing.T) {
	var msg = mmsg.NewMSG()
	newAtAllMsg(msg)
	assert.NotNil(t, msg)
	e := msg.ToCombineMessage(mmsg.NewGroupTarget(test.G1)).FirstOrNil(func(e message.IMessageElement) bool {
		return e.Type() == message.At
	})
	assert.NotNil(t, e)
	assert.EqualValues(t, 0, e.(*message.AtElement).Target)
}

func TestNewAtIdsMsg(t *testing.T) {
	var msg = mmsg.NewMSG()
	newAtIdsMsg(msg, []int64{test.UID1, test.UID2})
	assert.NotNil(t, msg)
	assert.Len(t, msg.Elements(), 2)
}
