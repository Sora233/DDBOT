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
	"time"
)

func init() {
	RegisterExtFunc("atrue", func(b bool) string {
		if !b {
			panic("not true")
		}
		return ""
	})
	RegisterExtFunc("afalse", func(b bool) string {
		if b {
			panic("not false")
		}
		return ""
	})
}

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

func TestTemplateFuncs(t *testing.T) {
	s, err := runTemplate(`{{- float64 1 -}}{{- int 1 -}}{{- int64 1 -}}{{- toString 1 -}}{{- toString "1" -}}`, map[string]interface{}{"target": 123456})
	assert.Nil(t, err)
	assert.EqualValues(t, "11111", s)

	s, err = runTemplate(
		`{{- max 1 2 -}}{{- maxf 1 2 -}}{{- min 1 2 -}}{{- minf 1 2 -}}`, nil)

	assert.Nil(t, err)
	assert.EqualValues(t, "2211", s)
	s, err = runTemplate(
		`{{- hour -}}{{- minute -}}{{- second -}}{{- month -}}{{- year -}}{{- day -}}{{- yearday -}}{{- weekday -}}`, nil)
	assert.Nil(t, err)

	s, err = runTemplate(`
{{- $el := list -}}{{- $l := list 1 -}}{{- $ed := dict -}}{{- $d := dict 1 1 -}}
{{- if empty $el -}}0{{- end -}}
{{- if empty $l -}}1{{- end -}}
{{- if empty $ed -}}2{{- end -}}
{{- if empty $d -}}3{{- end -}}
{{- if empty nil -}}4{{- end -}}
{{- if empty 0 -}}5{{- end -}}
{{- if empty 1 -}}6{{- end -}}
{{- if empty 2.0 -}}7{{- end -}}
{{- if empty "" -}}8{{- end -}}
{{- if empty "0" -}}9{{- end -}}
{{- if nonEmpty "0"}}10{{- end -}}
`, nil)
	assert.Nil(t, err)
	assert.EqualValues(t, "0245810", s)

	s, err = runTemplate(`
{{- if all 0 1 2 3 -}}0{{- end -}}
{{- if all 1 2 3 -}}1{{- end -}}
{{- coalesce 0 1 2 3 -}}
{{- if any 0 1 2 3 -}}2{{- end -}}
`, nil)
	assert.Nil(t, err)
	assert.EqualValues(t, "112", s)

	s, err = runTemplate(`
{{- $d := dict -}}
{{- if empty (get $d "100") -}}0{{- end -}}
{{- $d = set $d "100" 100 -}}
{{- atrue (hasKey $d "100") -}}
{{- if empty (get $d "100") -}}1{{- end -}}
{{- $d = unset $d "100" -}}
{{- if empty (get $d "100") -}}2{{- end -}}
{{- afalse (hasKey $d "100") -}}
{{- $d = set $d "100" 100 -}}
{{- atrue (eq (index (keys $d) 0) "100" ) -}}
`, nil)
	assert.Nil(t, err)
	assert.EqualValues(t, "02", s)

	s, err = runTemplate(`
{{ $d := base64encode "xx" }}
{{ base64decode $d }}
{{ md5sum "123 "}}
{{ sha1sum "123" }}
{{ sha256sum "123" }}
{{ adler32sum "123"}}
{{ uuid }}
`, nil)
	assert.Nil(t, err)

	s, err = runTemplate(`
{{- $l := list 1 2 3 -}}
{{- atrue (eq (len $l) 3) -}}
{{- $l = append $l 5 -}}
{{- atrue (eq (index $l 3) 5) -}}
{{- $l = prepend $l 10 -}}
{{- atrue (eq (index $l 0) 10) -}}
{{- $l2 := list 100 -}}
{{- $l = concat $l2 $l -}}
{{- atrue (eq (index $l 0) 100) -}}
`, nil)

	s, err = runTemplate("{{- $g := toGJson `{\"a\": \"b\"}` -}}{{- $g = toGJson $g -}}", nil)
	assert.Nil(t, err)

	s, err = runTemplate(`
{{- $s := "abcdefghijkl" -}}
{{- $s = trunc 2 $s -}}
{{- atrue (eq "ab" $s) -}}
{{- $l := list "ab" "cd" "ef" "g" 1 -}}
{{- $s = join "_" $l -}}
{{- atrue (eq "ab_cd_ef_g_1" $s) -}}
`, nil)
}

func TestTemplateCoolDown(t *testing.T) {
	test.InitBuntdb(t)
	defer test.CloseBuntdb(t)
	s, err := runTemplate(`
{{- if (cooldown "2s" "test1") -}}
true
{{- else -}}
false
{{- end -}}`, nil)
	assert.Nil(t, err)
	assert.EqualValues(t, "true", s)

	s, err = runTemplate(`
{{- if (cooldown "2s" "test1") -}}
true
{{- else -}}
false
{{- end -}}`, nil)
	assert.Nil(t, err)
	assert.EqualValues(t, "false", s)

	s, err = runTemplate(`
{{- if (cooldown "2s" "test2") -}}
true
{{- else -}}
false
{{- end -}}`, nil)
	assert.Nil(t, err)
	assert.EqualValues(t, "true", s)

	time.Sleep(time.Millisecond * 2500)

	s, err = runTemplate(`
{{- if (cooldown "2s" "test1") -}}
true
{{- else -}}
false
{{- end -}}`, nil)
	assert.Nil(t, err)
	assert.EqualValues(t, "true", s)
}

func TestAbort(t *testing.T) {
	test.InitBuntdb(t)
	defer test.CloseBuntdb(t)
	s, err := runTemplate(`abcd
	cdef
	{{- abort -}}`, nil)
	assert.Nil(t, err)
	assert.EqualValues(t, "", s)

	s, err = runTemplate(`abcd
	{{- if false -}}
		{{- abort -}}
	{{- end -}}`, nil)
	assert.Nil(t, err)
	assert.EqualValues(t, "abcd", s)

	s, err = runTemplate(`abcd
	{{- abort "tttt" -}}`, nil)
	assert.Nil(t, err)
	assert.EqualValues(t, "tttt", s)

	s, err = runTemplate(`abcd
	{{- abort (printf "%v-%v" 5 2) -}}`, nil)
	assert.Nil(t, err)
	assert.EqualValues(t, "5-2", s)

	s, err = runTemplate(`{{- if eq 1 5 -}}
	 {{- abort (printf "出现错误: %v居然等于%v" 1 5) -}}
	{{- end -}}`, nil)
	assert.Nil(t, err)
	assert.EqualValues(t, "", s)

	s, err = runTemplate(`{{- abort (pic "invalid") -}}`, nil)
	assert.Nil(t, err)
	assert.EqualValues(t, "[Image]", s)
}

func TestFin(t *testing.T) {
	test.InitBuntdb(t)
	defer test.CloseBuntdb(t)
	s, err := runTemplate(`abcd
{{- fin -}}
defg`, nil)
	assert.Nil(t, err)
	assert.EqualValues(t, "abcd", s)

	s, err = runTemplate(`abcd
{{- if false -}}
	{{- fin -}}
{{- end -}}
defg`, nil)
	assert.Nil(t, err)
	assert.EqualValues(t, "abcddefg", s)
}

func runTemplate(template string, data map[string]interface{}) (string, error) {
	var m = mmsg.NewMSG()
	var tmpl = Must(New("").Parse(template))
	var err = tmpl.Execute(m, data)
	return msgstringer.MsgToString(m.Elements()), err
}
