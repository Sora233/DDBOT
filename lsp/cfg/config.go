package cfg

import (
	"errors"
	"github.com/Sora233/MiraiGo-Template/config"
	"strings"
	"time"
)

func MatchCmdWithPrefix(cmd string) (prefix string, command string, err error) {
	var customPrefixCfg = GetCustomCommandPrefix()
	if customPrefixCfg != nil {
		for k, v := range customPrefixCfg {
			if v+k == cmd {
				return v, k, nil
			}
		}
	}
	commonPrefix := GetCommandPrefix()
	if strings.HasPrefix(cmd, commonPrefix) {
		return commonPrefix, strings.TrimPrefix(cmd, commonPrefix), nil
	}
	return "", "", errors.New("match failed")
}

func GetCommandPrefix(commands ...string) string {
	if len(commands) > 0 {
		var customPrefixCfg = GetCustomCommandPrefix()
		if customPrefixCfg != nil {
			if prefix, found := customPrefixCfg[commands[0]]; found {
				return prefix
			}
		}
	}
	prefix := strings.TrimSpace(config.GlobalConfig.GetString("bot.commandPrefix"))
	if len(prefix) == 0 {
		prefix = "/"
	}
	return prefix
}

func GetCustomCommandPrefix() map[string]string {
	return config.GlobalConfig.GetStringMapString("customCommandPrefix")
}

func GetEmitInterval() time.Duration {
	return config.GlobalConfig.GetDuration("concern.emitInterval")
}

func GetLargeNotifyLimit() int {
	var limit = config.GlobalConfig.GetInt("dispatch.largeNotifyLimit")
	if limit <= 0 {
		limit = 50
	}
	return limit
}

func GetCustomGroupCommand() []string {
	return config.GlobalConfig.GetStringSlice("autoreply.group.command")
}

func GetCustomPrivateCommand() []string {
	return config.GlobalConfig.GetStringSlice("autoreply.private.command")
}

func GetFramMessage() bool {
	return config.GlobalConfig.GetBool("bot.framMessage")
}

func GetBilibiliMinFollowerCap() int {
	return config.GlobalConfig.GetInt("bilibili.minFollowerCap")
}

func GetBilibiliDisableSub() bool {
	return config.GlobalConfig.GetBool("bilibili.disableSub")
}

func GetBilibiliHiddenSub() bool {
	return config.GlobalConfig.GetBool("bilibili.hiddenSub")
}

func GetBilibiliUnsub() bool {
	return config.GlobalConfig.GetBool("bilibili.unsub")
}

func GetNotifyParallel() int {
	var parallel = config.GlobalConfig.GetInt("notify.parallel")
	if parallel <= 0 {
		parallel = 1
	}
	return parallel
}

func GetBilibiliOnlyOnlineNotify() bool {
	return config.GlobalConfig.GetBool("bilibili.onlyOnlineNotify")
}
