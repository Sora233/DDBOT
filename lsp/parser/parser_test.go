package parser

import (
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewParser(t *testing.T) {
	p := NewParser()
	assert.NotNil(t, p)
	p.Command = "cmd"
	p.Args = []string{"1", "2", "3"}

	assert.Equal(t, "cmd", p.GetCmd())
	assert.EqualValues(t, []string{"1", "2", "3"}, p.GetArgs())
	assert.EqualValues(t, []string{"cmd", "1", "2", "3"}, p.GetCmdArgs())
}

func TestParser_Parse(t *testing.T) {
	p := NewParser()
	assert.NotNil(t, p)

	p.Parse([]message.IMessageElement{message.NewAt(0), message.NewText(" "), message.NewText("/a -b 1 -c 2")})

	assert.EqualValues(t, "/a", p.GetCmd())
	assert.EqualValues(t, []string{"-b", "1", "-c", "2"}, p.GetArgs())
	assert.EqualValues(t, []string{"/a", "-b", "1", "-c", "2"}, p.GetCmdArgs())
	assert.True(t, p.AtCheck())
}
