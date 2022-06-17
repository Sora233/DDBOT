package parser

import (
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Sora233/DDBOT/internal/test"
	"github.com/Sora233/DDBOT/utils"
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
	defer utils.GetBot().TESTReset()
	p := NewParser()
	assert.NotNil(t, p)

	p.Parse([]message.IMessageElement{message.NewAt(0), message.NewText(" "), message.NewText("/a -b 1 -c 2")})

	assert.EqualValues(t, "/a", p.GetCmd())
	assert.EqualValues(t, []string{"-b", "1", "-c", "2"}, p.GetArgs())
	assert.EqualValues(t, []string{"/a", "-b", "1", "-c", "2"}, p.GetCmdArgs())
	assert.True(t, p.AtCheck())

	utils.GetBot().TESTSetUin(test.UID1)
	p.Parse([]message.IMessageElement{message.NewAt(test.UID2), message.NewText(" "), message.NewText("/a -b 1 -c 2")})

	assert.False(t, p.AtCheck())
}

func TestParser_Parse2(t *testing.T) {
	defer utils.GetBot().TESTReset()
	p := NewParser()
	assert.NotNil(t, p)

	p.Parse(
		[]message.IMessageElement{
			message.NewText(" "),
			message.NewText("/a -b 1 -c 2"),
			&message.GroupImageElement{},
			message.NewText("-d 3"),
			message.NewAt(test.UID1),
			message.NewAt(test.UID2),
			message.NewText("-e 4"),
		},
	)
	assert.EqualValues(t, "/a", p.GetCmd())
	assert.EqualValues(t, []string{"-b", "1", "-c", "2", "-d", "3", "-e", "4"}, p.GetArgs())
	assert.EqualValues(t, []string{"/a", "-b", "1", "-c", "2", "-d", "3", "-e", "4"}, p.GetCmdArgs())
	assert.EqualValues(t, []int64{test.UID1, test.UID2}, p.GetAtArgs())
}
