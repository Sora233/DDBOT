package template

import (
	"fmt"
	"github.com/Sora233/DDBOT/internal/test"
	"github.com/Sora233/DDBOT/lsp/mmsg"
	"github.com/Sora233/DDBOT/utils/msgstringer"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadTemplate(t *testing.T) {
	InitTemplateLoader()
	defer Close()
	tmpl := LoadTemplate("command.group.help.tmpl")
	assert.NotNil(t, tmpl)
	var m = mmsg.NewMSG()
	err := tmpl.Execute(m, nil)
	assert.Nil(t, err)
	assert.Contains(t, msgstringer.MsgToString(m.Elements()), "DDBOT")

	m = mmsg.NewMSG()
	m, err = LoadAndExec("command.private.help.tmpl", nil)
	assert.Nil(t, err)
	assert.Contains(t, msgstringer.MsgToString(m.Elements()), "755612788")

	m = mmsg.NewMSG()
	tmpl = LoadTemplate("command.private.help.tmpl")
	err = tmpl.ExecuteTemplate(m, "command.private.help.tmpl", nil)
	assert.Nil(t, err)
	assert.Contains(t, msgstringer.MsgToString(m.Elements()), "755612788")

	m = mmsg.NewMSG()
	tmpl = LoadTemplate("trigger.group.member_in.tmpl")
	err = tmpl.ExecuteTemplate(m, "trigger.group.member_in.tmpl", nil)
	assert.Nil(t, err)
	assert.Empty(t, msgstringer.MsgToString(m.Elements()))
	assert.Empty(t, m.ToMessage(mmsg.NewGroupTarget(test.G1)))
}

func TestTemplateOption(t *testing.T) {
	var tmpl = New("test")
	tmpl.Option("missingkey=zero")
}

func TestTemplate(t *testing.T) {
	var templates = []string{
		`{{ range $x := .nums }}{{print $x}}{{ end }}`,
		`{{ range $x, $y := .data }}{{ $x }}{{ $y }}{{ end }}`,
	}
	var data = []map[string]interface{}{
		{
			"nums": []int{11, 22, 33, 44, 55},
		},
		{
			"data": map[string]string{
				"t1": "s1",
				"t2": "s2",
			},
		},
	}
	var expected = []string{
		"1122334455",
		"t1s1t2s2",
	}

	assert.EqualValues(t, len(templates), len(data))
	assert.EqualValues(t, len(templates), len(expected))

	for idx := range templates {
		var m = mmsg.NewMSG()
		var tmpl = Must(New(fmt.Sprintf("test-%v", idx)).Parse(templates[idx]))
		assert.Nil(t, tmpl.Execute(m, data[idx]))
		assert.EqualValues(t, expected[idx], msgstringer.MsgToString(m.Elements()), "%v mismatched", idx)
	}
}

func TestRoll(t *testing.T) {
	var a = roll(0, 10)
	assert.True(t, a >= 0)
	assert.True(t, a <= 10)
}

func TestPic(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "")
	assert.Nil(t, err)
	defer os.RemoveAll(tempDir)
	f, err := os.Create(filepath.Join(tempDir, "test.jpg"))
	assert.Nil(t, err)
	f.Write([]byte{0, 1, 2, 3})
	f.Close()
	var e = pic(tempDir)
	assert.NotNil(t, e)
	assert.EqualValues(t, []byte{0, 1, 2, 3}, e.Buf)
}
