package concern

import (
	"github.com/Sora233/DDBOT/lsp/concern_type"
	"github.com/Sora233/DDBOT/lsp/mmsg"
	"github.com/stretchr/testify/assert"
	"testing"
)

type testConcern struct {
	site  string
	types []concern_type.Type
}

func (t *testConcern) Start() error {
	return nil
}

func (t *testConcern) Stop() {
}

func (t *testConcern) ParseId(s string) (interface{}, error) {
	return s, nil
}

func (t *testConcern) Add(ctx mmsg.IMsgCtx, groupCode int64, id interface{}, ctype concern_type.Type) (IdentityInfo, error) {
	return nil, nil
}

func (t *testConcern) Remove(ctx mmsg.IMsgCtx, groupCode int64, id interface{}, ctype concern_type.Type) (IdentityInfo, error) {
	return nil, nil
}

func (t *testConcern) List(groupCode int64, ctype concern_type.Type) ([]IdentityInfo, []concern_type.Type, error) {
	return nil, nil, nil
}

func (t *testConcern) Get(id interface{}) (IdentityInfo, error) {
	return nil, nil
}

func (t *testConcern) GetStateManager() IStateManager {
	return nil
}

func (t *testConcern) FreshIndex(groupCode ...int64) {
}

func (t *testConcern) Site() string {
	return t.site
}
func (t *testConcern) Types() []concern_type.Type {
	return t.types
}

func TestGetNotifyChan(t *testing.T) {
	assert.NotNil(t, GetNotifyChan())
}

func TestReadNotifyChan(t *testing.T) {
	assert.NotNil(t, ReadNotifyChan())
}

func TestConcernManager(t *testing.T) {
	assert.Empty(t, ListConcernManager())

	RegisterConcernManager(&testConcern{site: "test1", types: []concern_type.Type{
		"1",
		"2",
		"3",
	}})

	RegisterConcernManager(&testConcern{site: "test2", types: []concern_type.Type{
		"4",
		"5",
		"6",
	}})

	assert.Panics(t,
		func() {
			RegisterConcernManager(&testConcern{site: "test2", types: []concern_type.Type{
				"4",
				"5",
				"6",
			}})
		},
	)

	assert.Panics(t,
		func() {
			RegisterConcernManager(&testConcern{site: "test3", types: []concern_type.Type{
				"4/5/6",
			}})
		},
	)
	assert.Panics(t,
		func() {
			RegisterConcernManager(&testConcern{site: "test10"})
		},
	)
	assert.Panics(t,
		func() {
			RegisterConcernManager(nil)
		},
	)

	assert.Contains(t, ListSite(), "test1")
	assert.Contains(t, ListSite(), "test2")

	ctypes, err := GetConcernManagerTypes("test1")
	assert.Nil(t, err)
	assert.EqualValues(t, []concern_type.Type{"1", "2", "3"}, ctypes.Split())
	ctypes, err = GetConcernManagerTypes("test2")
	assert.Nil(t, err)
	assert.EqualValues(t, []concern_type.Type{"4", "5", "6"}, ctypes.Split())
	ctypes, err = GetConcernManagerTypes("test3")
	assert.EqualValues(t, ErrSiteNotSupported, err)

	cm, err := GetConcernManagerBySite("wrong")
	assert.EqualValues(t, ErrSiteNotSupported, err)
	assert.Nil(t, cm)
	cm, err = GetConcernManagerBySite("test1")
	assert.Nil(t, err)
	assert.NotNil(t, cm)
	cm, err = GetConcernManagerBySiteAndType("test1", "1")
	assert.Nil(t, err)
	assert.NotNil(t, cm)
	cm, err = GetConcernManagerBySiteAndType("test1", "3")
	assert.Nil(t, err)
	assert.NotNil(t, cm)
	cm, err = GetConcernManagerBySiteAndType("test1", "4")
	assert.EqualValues(t, ErrTypeNotSupported, err)
	assert.Nil(t, cm)
	cm, err = GetConcernManagerBySiteAndType("test4", "10")
	assert.EqualValues(t, ErrSiteNotSupported, err)
	assert.Nil(t, cm)

	cm, err = GetConcernManagerByParseSite("wrong")
	assert.Nil(t, cm)
	assert.EqualValues(t, ErrSiteNotSupported, err)

	cm, err = GetConcernManagerByParseSite("test1")
	assert.Nil(t, err)
	assert.NotNil(t, cm)

	cm, _, ctypes, err = GetConcernManagerByParseSiteAndType("test1", "1")
	assert.Nil(t, err)
	assert.NotNil(t, cm)
	assert.EqualValues(t, "1", ctypes)

	cm, _, ctypes, err = GetConcernManagerByParseSiteAndType("test2", "")
	assert.Nil(t, err)
	assert.NotNil(t, cm)
	assert.EqualValues(t, "4", ctypes)

	cm, _, ctypes, err = GetConcernManagerByParseSiteAndType("test", "")
	assert.Nil(t, cm)
	assert.EqualValues(t, ErrSiteNotSupported, err)

	assert.Nil(t, StartAll())

	StopAll()

	site, err := ParseRawSite("test1")
	assert.Nil(t, err)
	assert.Equal(t, "test1", site)

	site, err = ParseRawSite("test2")
	assert.Nil(t, err)
	assert.Equal(t, "test2", site)

	site, err = ParseRawSite("test3")
	assert.NotNil(t, err)

	site, ctype, err := ParseRawSiteAndType("test1", "1")
	assert.Nil(t, err)
	assert.Equal(t, "test1", site)
	assert.EqualValues(t, "1", ctype)

	site, ctype, err = ParseRawSiteAndType("test1", "3")
	assert.Nil(t, err)
	assert.Equal(t, "test1", site)
	assert.EqualValues(t, "3", ctype)

	site, ctype, err = ParseRawSiteAndType("test1", "4")
	assert.NotNil(t, err)

	site, ctype, err = ParseRawSiteAndType("test2", "4")
	assert.Nil(t, err)
	assert.Equal(t, "test2", site)
	assert.EqualValues(t, "4", ctype)

	site, ctype, err = ParseRawSiteAndType("test3", "4")
	assert.NotNil(t, err)

	ClearConcern()
}
