package template

import (
	"fmt"
	"github.com/Mrs4s/MiraiGo/message"
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

func TestInt64(t *testing.T) {
	// 因为64位没有问题
	// 在32位上这个case才有意义
	var m = mmsg.NewMSG()
	const target int64 = 1234567891011
	var template = `{{- if eq .target 1234567891011 -}}asd{{- end -}}`
	var tmpl = Must(New("test-int64").Parse(template))
	err := tmpl.Execute(m, map[string]interface{}{"target": target})
	assert.Nil(t, err)
	assert.Len(t, m.Elements(), 1)
	assert.EqualValues(t, m.Elements()[0].(*message.TextElement).Content, "asd")
}

func TestCompareStringAndInt(t *testing.T) {
	var m = mmsg.NewMSG()
	var err error
	var template = `{{- if eq .target 123456 -}}asd{{- end -}}`
	var tmpl = Must(New("test-compare-string-and-int").Parse(template))
	err = tmpl.Execute(m, map[string]interface{}{"target": 123456})
	assert.Nil(t, err)
	err = tmpl.Execute(m, map[string]interface{}{"target": "123456"})
	assert.Nil(t, err)

	assert.Len(t, m.Elements(), 1)
	assert.EqualValues(t, "asdasd", m.Elements()[0].(*message.TextElement).Content)

	m = mmsg.NewMSG()
	template = `{{- if eq .target "123456" -}}asd{{- end -}}`
	tmpl = Must(New("test-compare-string-and-int-1").Parse(template))
	err = tmpl.Execute(m, map[string]interface{}{"target": 123456})
	assert.Nil(t, err)
	err = tmpl.Execute(m, map[string]interface{}{"target": "123456"})
	assert.Nil(t, err)

	assert.Len(t, m.Elements(), 1)
	assert.EqualValues(t, "asdasd", m.Elements()[0].(*message.TextElement).Content)

	m = mmsg.NewMSG()
	template = `{{- if eq .target 123456 -}}asd{{- end -}}`
	tmpl = Must(New("test-compare-string-and-int-1").Parse(template))
	err = tmpl.Execute(m, map[string]interface{}{"target": "qweasdzxc"})
	assert.NotNil(t, err)

	m = mmsg.NewMSG()
	template = `{{- if lt .target 999 -}}1{{- end -}}
{{- if le .target 999 -}}2{{- end -}}
{{- if gt .target 100 -}}3{{- end -}}
{{- if ge .target 100 -}}4{{- end -}}
{{- if lt .target 0 -}}5{{- end -}}
{{- if le .target 0 -}}6{{- end -}}
{{- if gt .target 1000 -}}7{{- end -}}
{{- if ge .target 1000 -}}8{{- end -}}
`
	tmpl = Must(New("test-compare-string-and-int-1").Parse(template))
	err = tmpl.Execute(m, map[string]interface{}{"target": "200"})
	assert.Nil(t, err)
	assert.EqualValues(t, "1234", m.Elements()[0].(*message.TextElement).Content)
}
