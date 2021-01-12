package lsp

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
	ImageContentCommand = "ic"
)

var all = [...]string{
	RollCommand, CheckinCommand, GrantCommand,
	LspCommand, WatchCommand, UnwatchCommand,
	ListCommand, SetuCommand, HuangtuCommand,
	EnableCommand, DisableCommand, ImageContentCommand,
}

func CheckCommand(command string) bool {
	for _, e := range all {
		if e == command {
			return true
		}
	}
	return false
}
