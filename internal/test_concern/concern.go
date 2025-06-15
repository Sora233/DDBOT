package test_concern

import (
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/Sora233/DDBOT/v2/lsp/concern"
	"github.com/Sora233/DDBOT/v2/lsp/concern_type"
	"github.com/Sora233/DDBOT/v2/lsp/mmsg"
)

type TestEvent struct {
	site      string
	ctype     concern_type.Type
	id        string
	groupCode uint32
}

func (t *TestEvent) GetGroupCode() uint32 {
	return t.groupCode
}

func (t *TestEvent) ToMessage() *mmsg.MSG {
	return mmsg.NewTextf("%v %v %v %v", t.site, t.ctype.String(), t.groupCode, t.id)
}

func (t *TestEvent) Site() string {
	return t.site
}

func (t *TestEvent) Type() concern_type.Type {
	return t.ctype
}

func (t *TestEvent) GetUid() interface{} {
	return t.id
}

func (t *TestEvent) Logger() *logrus.Entry {
	return logrus.WithField("site", t.site).
		WithField("ctype", t.ctype.String()).
		WithField("id", t.id).WithField("group_code", t.groupCode)
}

type TestConcern struct {
	*concern.StateManager
	site   string
	Ctypes []concern_type.Type
}

func (t *TestConcern) NewTestEvent(p concern_type.Type, groupCode uint32, id string) *TestEvent {
	return &TestEvent{
		site:      t.site,
		ctype:     p,
		id:        id,
		groupCode: groupCode,
	}
}

func (t *TestConcern) Site() string {
	return t.site
}

func (t *TestConcern) Types() []concern_type.Type {
	return t.Ctypes
}

func (t *TestConcern) ParseId(s string) (interface{}, error) {
	return s, nil
}

func (t *TestConcern) Add(ctx mmsg.IMsgCtx, groupCode uint32, id interface{}, ctype concern_type.Type) (concern.IdentityInfo, error) {
	_, err := t.StateManager.AddGroupConcern(groupCode, id, ctype)
	return concern.NewIdentity(id, id.(string)), err
}

func (t *TestConcern) Remove(ctx mmsg.IMsgCtx, groupCode uint32, id interface{}, ctype concern_type.Type) (concern.IdentityInfo, error) {
	_, err := t.StateManager.RemoveGroupConcern(groupCode, id, ctype)
	return concern.NewIdentity(id, id.(string)), err
}

func (t *TestConcern) Get(id interface{}) (concern.IdentityInfo, error) {
	return concern.NewIdentity(id, id.(string)), nil
}

func (t *TestConcern) GetStateManager() concern.IStateManager {
	return t.StateManager
}

func (t *TestConcern) UseFresh(freshFunc concern.FreshFunc) {
	t.UseFreshFunc(freshFunc)
}

func (t *TestConcern) UseNotifyGenerator(generator concern.NotifyGeneratorFunc) {
	t.UseNotifyGeneratorFunc(generator)
}

func (t *TestConcern) TestNotifyGenerator() concern.NotifyGeneratorFunc {
	return func(groupCode uint32, event concern.Event) []concern.Notify {
		e := event.(*TestEvent)
		e.groupCode = groupCode
		return []concern.Notify{e}
	}
}

func NewTestConcern(notifyChan chan<- concern.Notify, site string, p []concern_type.Type) *TestConcern {
	tc := &TestConcern{
		StateManager: concern.NewStateManagerWithStringID(fmt.Sprintf("test-%v", site), notifyChan),
		site:         site,
		Ctypes:       p,
	}
	tc.UseNotifyGenerator(tc.TestNotifyGenerator())
	return tc
}
