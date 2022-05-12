package test_concern

import (
	"fmt"
	"github.com/Sora233/DDBOT/lsp/concern"
	"github.com/Sora233/DDBOT/lsp/concern_type"
	"github.com/Sora233/DDBOT/lsp/mmsg"
	"github.com/Sora233/DDBOT/lsp/mmsg/mt"
	"github.com/sirupsen/logrus"
)

type TestEvent struct {
	site   string
	ctype  concern_type.Type
	id     string
	target mt.Target
}

func (t *TestEvent) GetTarget() mt.Target {
	return t.target
}

func (t *TestEvent) ToMessage() *mmsg.MSG {
	return mmsg.NewTextf("%v %v %v %v", t.site, t.ctype.String(), t.target, t.id)
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
		WithField("id", t.id).WithField("group_code", t.target)
}

type TestConcern struct {
	*concern.StateManager
	site   string
	Ctypes []concern_type.Type
}

func (t *TestConcern) NewTestEvent(p concern_type.Type, target mt.Target, id string) *TestEvent {
	return &TestEvent{
		site:   t.site,
		ctype:  p,
		id:     id,
		target: target,
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

func (t *TestConcern) Add(ctx mmsg.IMsgCtx, target mt.Target, id interface{}, ctype concern_type.Type) (concern.IdentityInfo, error) {
	_, err := t.StateManager.AddTargetConcern(target, id, ctype)
	return concern.NewIdentity(id, id.(string)), err
}

func (t *TestConcern) Remove(ctx mmsg.IMsgCtx, target mt.Target, id interface{}, ctype concern_type.Type) (concern.IdentityInfo, error) {
	_, err := t.StateManager.RemoveTargetConcern(target, id, ctype)
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
	return func(target mt.Target, event concern.Event) []concern.Notify {
		e := event.(*TestEvent)
		e.target = target
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
