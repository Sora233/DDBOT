package lsp

import "github.com/Sora233/sliceutil"

// TODO command需要重构成注册模式，然后把这个文件废弃

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
	ReverseCommand      = "倒放"
	ImageContentCommand = "ic"
	HelpCommand         = "help"
	ConfigCommand       = "config"
)

// private command
const (
	PingCommand          = "ping"
	LogCommand           = "log"
	BlockCommand         = "block"
	SysinfoCommand       = "sysinfo"
	WhosyourdaddyCommand = "whosyourdaddy"
)

var allGroupCommand = [...]string{
	RollCommand, CheckinCommand, GrantCommand,
	LspCommand, WatchCommand, UnwatchCommand,
	ListCommand, SetuCommand, HuangtuCommand,
	EnableCommand, DisableCommand, ImageContentCommand,
	FaceCommand, ReverseCommand, ConfigCommand,
	HelpCommand,
}

var allPrivateOperate = [...]string{
	PingCommand, HelpCommand, LogCommand,
	BlockCommand, SysinfoCommand, ListCommand,
	WatchCommand, UnwatchCommand, DisableCommand,
	EnableCommand, GrantCommand, ConfigCommand,
	WhosyourdaddyCommand,
}

var nonOprateable = [...]string{
	EnableCommand, DisableCommand, GrantCommand,
	BlockCommand, LogCommand, PingCommand, WhosyourdaddyCommand,
}

func CheckValidCommand(command string) bool {
	return sliceutil.Contains(allGroupCommand, command)
}

func CheckOperateableCommand(command string) bool {
	return sliceutil.Contains(allGroupCommand, command) && !sliceutil.Contains(nonOprateable, command)
}

func CombineCommand(command string) string {
	if command == WatchCommand || command == UnwatchCommand {
		return WatchCommand
	}
	return command
}
