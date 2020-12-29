package command

import (
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Sora233/Sora233-MiraiGo/command"
	"strings"
)

type LspGroupCommandManager struct {
	commandMap    map[string]command.GroupCommand
	commandPrefix string
}

func NewLspGroupCommandManager(commandPrefix string) *LspGroupCommandManager {
	return &LspGroupCommandManager{
		commandMap:    make(map[string]command.GroupCommand),
		commandPrefix: commandPrefix,
	}
}

func (m *LspGroupCommandManager) Execute(msg *message.GroupMessage) error {
	return nil
}

func (m *LspGroupCommandManager) Register(primaryArg string, command command.GroupCommand) error {
	primaryArg = strings.TrimSpace(primaryArg)
	if !m.checkPrimaryArg(primaryArg) {
		return ErrInvalidPrimaryArg
	}
	if _, found := m.commandMap[strings.TrimSpace(primaryArg)]; found {
		return ErrPrimaryArgExist
	}
	m.commandMap[primaryArg] = command
	return nil
}

func (m *LspGroupCommandManager) checkPrimaryArg(arg string) bool {
	if strings.HasPrefix(arg, m.commandPrefix) {
		return false
	}
	return true
}
