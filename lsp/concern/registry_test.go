package concern

import (
	"errors"
	"github.com/Sora233/DDBOT/lsp/concern_type"
	"github.com/Sora233/DDBOT/lsp/mmsg"
	"github.com/Sora233/DDBOT/lsp/mmsg/mt"
	"github.com/stretchr/testify/assert"
	"testing"
)

type testConcern struct {
	site     string
	types    []concern_type.Type
	startErr error
}

func (t *testConcern) Site() string {
	return t.site
}

func (t *testConcern) Types() []concern_type.Type {
	return t.types
}

func (t *testConcern) Start() error {
	return t.startErr
}

func (t *testConcern) Stop() {
}

func (t *testConcern) ParseId(s string) (interface{}, error) {
	return s, nil
}

func (t *testConcern) Add(ctx mmsg.IMsgCtx, target mt.Target, id interface{}, ctype concern_type.Type) (IdentityInfo, error) {
	return nil, nil
}

func (t *testConcern) Remove(ctx mmsg.IMsgCtx, target mt.Target, id interface{}, ctype concern_type.Type) (IdentityInfo, error) {
	return nil, nil
}

func (t *testConcern) Get(id interface{}) (IdentityInfo, error) {
	return nil, nil
}

func (t *testConcern) GetStateManager() IStateManager {
	return nil
}

func (t *testConcern) FreshIndex(targets ...mt.Target) {
}

func TestGetNotifyChan(t *testing.T) {
	assert.NotNil(t, GetNotifyChan())
}

func TestReadNotifyChan(t *testing.T) {
	assert.NotNil(t, ReadNotifyChan())
}

func TestConcernManager(t *testing.T) {
	assert.Empty(t, ListConcern())

	RegisterConcern(&testConcern{site: "test1", types: []concern_type.Type{
		"1",
		"2",
		"3",
	}})

	RegisterConcern(&testConcern{site: "test2", types: []concern_type.Type{
		"4",
		"5",
		"6",
	}})

	assert.Panics(t,
		func() {
			RegisterConcern(&testConcern{site: "test2", types: []concern_type.Type{
				"4",
				"5",
				"6",
			}})
		},
	)

	assert.Panics(t,
		func() {
			RegisterConcern(&testConcern{site: "test3", types: []concern_type.Type{
				"4/5/6",
			}})
		},
	)
	assert.Panics(t,
		func() {
			RegisterConcern(&testConcern{site: "test10"})
		},
	)
	assert.Panics(t,
		func() {
			RegisterConcern(nil)
		},
	)

	RegisterConcern(&testConcern{
		site: "errSite",
		types: []concern_type.Type{
			"9",
		},
		startErr: errors.New("error"),
	})

	assert.Contains(t, ListSite(), "test1")
	assert.Contains(t, ListSite(), "test2")
	assert.Contains(t, ListSite(), "errSite")

	ctypes, err := GetConcernTypes("test1")
	assert.Nil(t, err)
	assert.EqualValues(t, []concern_type.Type{"1", "2", "3"}, ctypes.Split())
	ctypes, err = GetConcernTypes("test2")
	assert.Nil(t, err)
	assert.EqualValues(t, []concern_type.Type{"4", "5", "6"}, ctypes.Split())
	ctypes, err = GetConcernTypes("test3")
	assert.EqualValues(t, ErrSiteNotSupported, err)

	cm, err := GetConcernBySite("wrong")
	assert.EqualValues(t, ErrSiteNotSupported, err)
	assert.Nil(t, cm)
	cm, err = GetConcernBySite("test1")
	assert.Nil(t, err)
	assert.NotNil(t, cm)
	cm, err = GetConcernBySiteAndType("test1", "1")
	assert.Nil(t, err)
	assert.NotNil(t, cm)
	cm, err = GetConcernBySiteAndType("test1", "3")
	assert.Nil(t, err)
	assert.NotNil(t, cm)
	cm, err = GetConcernBySiteAndType("test1", "4")
	assert.EqualValues(t, ErrTypeNotSupported, err)
	assert.Nil(t, cm)
	cm, err = GetConcernBySiteAndType("test4", "10")
	assert.EqualValues(t, ErrSiteNotSupported, err)
	assert.Nil(t, cm)

	cm, err = GetConcernByParseSite("wrong")
	assert.Nil(t, cm)
	assert.EqualValues(t, ErrSiteNotSupported, err)

	cm, err = GetConcernByParseSite("test1")
	assert.Nil(t, err)
	assert.NotNil(t, cm)

	cm, _, ctypes, err = GetConcernByParseSiteAndType("test1", "1")
	assert.Nil(t, err)
	assert.NotNil(t, cm)
	assert.EqualValues(t, "1", ctypes)

	cm, _, ctypes, err = GetConcernByParseSiteAndType("test2", "")
	assert.Nil(t, err)
	assert.NotNil(t, cm)
	assert.EqualValues(t, "4", ctypes)

	cm, _, ctypes, err = GetConcernByParseSiteAndType("test", "")
	assert.Nil(t, cm)
	assert.EqualValues(t, ErrSiteNotSupported, err)

	StartAll()

	assert.NotContains(t, ListSite(), "errSite")

	_, err = GetConcernBySite("errSite")
	assert.EqualValues(t, ErrSiteNotSupported, err)

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
