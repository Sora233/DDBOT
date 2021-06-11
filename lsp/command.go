package lsp

import "github.com/Sora233/sliceutil"

const (
	RollCommand         = "roll"
	CheckinCommand      = "签到"
	GrantCommand        = "grant"
	LspCommand          = "lsp"
	WatchCommand        = "watch"
	UnwatchCommand      = "unwatch"
	ListCommand         = "list"
	SetuCommand         = "色图"
	HuangtuCommand      = "黄图"
	EnableCommand       = "enable"
	DisableCommand      = "disable"
	FaceCommand         = "face"
	ReverseCommand      = "reverse"
	ImageContentCommand = "ic"
	HelpCommand         = "help"
	ConfigCommand       = "config"
)

// private command
const (
	PingCommand    = "ping"
	LogCommand     = "log"
	BlockCommand   = "block"
	SysinfoCommand = "sysinfo"
)

var allGroupCommand = [...]string{
	RollCommand, CheckinCommand, GrantCommand,
	LspCommand, WatchCommand, UnwatchCommand,
	ListCommand, SetuCommand, HuangtuCommand,
	EnableCommand, DisableCommand, ImageContentCommand,
	FaceCommand, ConfigCommand,
}

var allPrivateOperate = [...]string{
	PingCommand, HelpCommand, LogCommand,
	BlockCommand, SysinfoCommand, ListCommand,
	WatchCommand, UnwatchCommand, DisableCommand,
	EnableCommand, GrantCommand, ConfigCommand,
}

var nonOprateable = [...]string{
	EnableCommand, DisableCommand, GrantCommand, HelpCommand,
	BlockCommand, LogCommand, PingCommand,
}

func CheckValidCommand(command string) bool {
	return sliceutil.Contains(allGroupCommand, command)
}

func CheckOperateableCommand(command string) bool {
	return sliceutil.Contains(allGroupCommand, command) && !sliceutil.Contains(nonOprateable, command)
}
