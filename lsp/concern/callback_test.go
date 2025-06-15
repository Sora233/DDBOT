package concern

import (
	"testing"

	"github.com/sirupsen/logrus"

	"github.com/Sora233/DDBOT/v2/internal/test"
	"github.com/Sora233/DDBOT/v2/lsp/concern_type"
	"github.com/Sora233/DDBOT/v2/lsp/mmsg"
)

type testNotify struct {
}

func (t *testNotify) Site() string {
	return "test"
}

func (t *testNotify) Type() concern_type.Type {
	return "test"
}

func (t *testNotify) GetUid() interface{} {
	return ""
}

func (t *testNotify) Logger() *logrus.Entry {
	return logrus.WithField("Site", t.Site())
}

func (t *testNotify) GetGroupCode() uint32 {
	return test.G1
}

func (t *testNotify) ToMessage() *mmsg.MSG {
	return mmsg.NewMSG()
}

func TestDefaultCallback(t *testing.T) {
	var d DefaultCallback
	d.NotifyBeforeCallback(nil)
	d.NotifyAfterCallback(nil, nil)
	d.NotifyBeforeCallback(new(testNotify))
	d.NotifyAfterCallback(new(testNotify), nil)
}
