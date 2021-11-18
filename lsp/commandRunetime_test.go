package lsp

import (
	"github.com/Sora233/DDBOT/internal/test"
	tc "github.com/Sora233/DDBOT/internal/test_concern"
	"github.com/Sora233/DDBOT/lsp/concern"
	"github.com/Sora233/DDBOT/lsp/concern_type"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewRuntime(t *testing.T) {
	defer concern.ClearConcern()
	r := NewRuntime(Instance, true)
	assert.NotNil(t, r)

	tc1 := tc.NewTestConcern(concern.GetNotifyChan(), test.Site1, []concern_type.Type{test.T1})
	concern.RegisterConcern(tc1)

	tc2 := tc.NewTestConcern(concern.GetNotifyChan(), test.Site2, []concern_type.Type{test.T2})
	concern.RegisterConcern(tc2)

	assert.Len(t, concern.ListSite(), 2)
	site, err := r.ParseRawSite(test.Site1)
	assert.Nil(t, err)
	assert.EqualValues(t, test.Site1, site)

	site, ctype, err := r.ParseRawSiteAndType(test.Site1, test.T1.String())
	assert.Nil(t, err)
	assert.EqualValues(t, test.Site1, site)
	assert.EqualValues(t, test.T1, ctype)

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
