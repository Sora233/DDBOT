package lsp

import "github.com/forestgiant/sliceutil"

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
	ImageContentCommand = "ic"
	HelpCommand         = "help"
	AboutCommand        = "about"
)

var all = [...]string{
	RollCommand, CheckinCommand, GrantCommand,
	LspCommand, WatchCommand, UnwatchCommand,
	ListCommand, SetuCommand, HuangtuCommand,
	EnableCommand, DisableCommand, ImageContentCommand,
	FaceCommand,
}

var nonOprateable = [...]string{
	EnableCommand, DisableCommand, GrantCommand, HelpCommand, AboutCommand,
}

func CheckValidCommand(command string) bool {
	return sliceutil.Contains(all, command)
}

func CheckOperateableCommand(command string) bool {
	return sliceutil.Contains(all, command) && !sliceutil.Contains(nonOprateable, command)
}
