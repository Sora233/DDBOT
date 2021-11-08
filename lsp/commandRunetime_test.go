package lsp

import (
	"github.com/Sora233/DDBOT/lsp/concern"
	"github.com/stretchr/testify/assert"
	"testing"

	_ "github.com/Sora233/DDBOT/lsp/acfun"
	_ "github.com/Sora233/DDBOT/lsp/bilibili"
	_ "github.com/Sora233/DDBOT/lsp/douyu"
	_ "github.com/Sora233/DDBOT/lsp/huya"
	_ "github.com/Sora233/DDBOT/lsp/weibo"
	_ "github.com/Sora233/DDBOT/lsp/youtube"
)

func TestNewRuntime(t *testing.T) {
	r := NewRuntime(Instance, true)
	assert.NotNil(t, r)

	assert.Len(t, concern.ListSite(), 6)
	site, err := r.ParseRawSite("bil")
	assert.Nil(t, err)
	assert.EqualValues(t, "bilibili", site)

	site, ctype, err := r.ParseRawSiteAndType("bil", "news")
	assert.Nil(t, err)
	assert.EqualValues(t, "bilibili", site)
	assert.EqualValues(t, "news", ctype)

	assert.False(t, r.exit)
	r.Exit(1)
	assert.True(t, r.exit)
	assert.False(t, r.debug)
	r.Debug()
	assert.True(t, r.debug)

	var testCmd struct {
		A string `arg:""`
		B string `optional:"" short:"b"`
	}
	r = NewRuntime(Instance)
	r.Command = "test"
	r.Args = []string{"qwe", "-b", "bbb"}
	_, output := r.parseCommandSyntax(&testCmd, "test")
	assert.Empty(t, output)
	assert.EqualValues(t, testCmd.A, "qwe")
	assert.EqualValues(t, testCmd.B, "bbb")

	r = NewRuntime(Instance)
	r.Command = "test"
	r.Args = []string{"-b", "bbb"}
	_, output = r.parseCommandSyntax(&testCmd, "test")
	assert.NotEmpty(t, output)
	assert.True(t, r.exit)

	r = NewRuntime(Instance, true)
	r.Command = "test"
	r.Args = []string{"-b", "bbb"}
	_, output = r.parseCommandSyntax(&testCmd, "test")
	assert.Empty(t, output)
	assert.True(t, r.exit)

	r = NewRuntime(Instance)
	r.Command = "test"
	r.Args = []string{"-h"}
	_, output = r.parseCommandSyntax(&testCmd, "test")
	assert.NotEmpty(t, output)
	assert.True(t, r.exit)

	r = NewRuntime(Instance, true)
	_, output = r.parseCommandSyntax(nil, "test")
	assert.Empty(t, output)
	assert.True(t, r.exit)
}
