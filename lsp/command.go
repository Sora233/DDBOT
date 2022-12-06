package lsp

import (
	"github.com/Sora233/DDBOT/lsp/cfg"
	"github.com/Sora233/sliceutil"
)

// TODO command需要重构成注册模式，然后把这个文件废弃

var CommandMaps = map[string]string{
	// group commands
	"RollCommand":    RollCommand,
	"CheckinCommand": CheckinCommand,
	"ScoreCommand":   ScoreCommand,
	"GrantCommand":   GrantCommand,
	"LspCommand":     LspCommand,
	"WatchCommand":   WatchCommand,
	"UnwatchCommand": UnwatchCommand,
	"ListCommand":    ListCommand,
	"SetuCommand":    SetuCommand,
	"HuangtuCommand": HuangtuCommand,
	"EnableCommand":  EnableCommand,
	"DisableCommand": DisableCommand,
	"ReverseCommand": ReverseCommand,
	"HelpCommand":    HelpCommand,
	"ConfigCommand":  ConfigCommand,

	"WaifuCommand":      WaifuCommand,
	"DivinationCommand": DivinationCommand,
	"FightCommand":      FightCommand,

	// private commands
	"PingCommand":          PingCommand,
	"LogCommand":           LogCommand,
	"BlockCommand":         BlockCommand,
	"SysinfoCommand":       SysinfoCommand,
	"WhosyourdaddyCommand": WhosyourdaddyCommand,
	"QuitCommand":          QuitCommand,
	"ModeCommand":          ModeCommand,
	"GroupRequestCommand":  GroupRequestCommand,
	"FriendRequestCommand": FriendRequestCommand,
	"AdminCommand":         AdminCommand,
	"SilenceCommand":       SilenceCommand,
	"NoUpdateCommand":      NoUpdateCommand,
	"AbnormalConcernCheck": AbnormalConcernCheck,
	"CleanConcern":         CleanConcern,
}

const (
	RollCommand    = "roll"
	CheckinCommand = "签到"
	ScoreCommand   = "查询积分"
	GrantCommand   = "grant"
	LspCommand     = "lsp"
	WatchCommand   = "watch"
	UnwatchCommand = "unwatch"
	ListCommand    = "list"
	SetuCommand    = "色图"
	HuangtuCommand = "黄图"
	EnableCommand  = "enable"
	DisableCommand = "disable"
	ReverseCommand = "倒放"
	HelpCommand    = "help"
	ConfigCommand  = "config"

	WaifuCommand      = "老婆"
	DivinationCommand = "占卜"
	FightCommand      = "打人"
)

// private command
const (
	PingCommand          = "ping"
	LogCommand           = "log"
	BlockCommand         = "block"
	SysinfoCommand       = "sysinfo"
	WhosyourdaddyCommand = "whosyourdaddy"
	QuitCommand          = "quit"
	ModeCommand          = "mode"
	GroupRequestCommand  = "群邀请"
	FriendRequestCommand = "好友申请"
	AdminCommand         = "admin"
	SilenceCommand       = "silence"
	NoUpdateCommand      = "退订更新"
	AbnormalConcernCheck = "检测异常订阅"
	CleanConcern         = "清除订阅"
)

var allGroupCommand = [...]string{
	RollCommand, CheckinCommand, GrantCommand,
	LspCommand, WatchCommand, UnwatchCommand,
	ListCommand, SetuCommand, HuangtuCommand,
	EnableCommand, DisableCommand,
	ReverseCommand, ConfigCommand,
	WaifuCommand, DivinationCommand, FightCommand,
	HelpCommand, ScoreCommand, AdminCommand,
	SilenceCommand, NoUpdateCommand, CleanConcern,
}

var allPrivateOperate = [...]string{
	PingCommand, HelpCommand, LogCommand,
	BlockCommand, SysinfoCommand, ListCommand,
	WatchCommand, UnwatchCommand, DisableCommand,
	EnableCommand, GrantCommand, ConfigCommand,
	WhosyourdaddyCommand, QuitCommand, ModeCommand,
	GroupRequestCommand, FriendRequestCommand, AdminCommand,
	SilenceCommand, NoUpdateCommand, AbnormalConcernCheck,
	CleanConcern,
}

var nonOprateable = [...]string{
	EnableCommand, DisableCommand, GrantCommand,
	BlockCommand, LogCommand, PingCommand,
	WhosyourdaddyCommand, QuitCommand, ModeCommand,
	GroupRequestCommand, FriendRequestCommand, AdminCommand,
	SilenceCommand, NoUpdateCommand, AbnormalConcernCheck,
	CleanConcern,
}

func CheckValidCommand(command string) bool {
	return sliceutil.Contains(allGroupCommand, command)
}

func CheckCustomGroupCommand(command string) bool {
	return sliceutil.Contains(cfg.GetCustomGroupCommand(), command)
}

func CheckCustomPrivateCommand(command string) bool {
	return sliceutil.Contains(cfg.GetCustomPrivateCommand(), command)
}

func CheckOperateableCommand(command string) bool {
	return (sliceutil.Contains(allGroupCommand, command) || CheckCustomGroupCommand(command)) && !sliceutil.Contains(nonOprateable, command)
}

func CombineCommand(command string) string {
	if command == WatchCommand || command == UnwatchCommand {
		return WatchCommand
	}
	return command
}
