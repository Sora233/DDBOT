package cfg

import (
	"github.com/Sora233/MiraiGo-Template/config"
	"strings"
	"time"
)

func GetCommandPrefix() string {
	if config.GlobalConfig == nil {
		return "/"
	}
	prefix := strings.TrimSpace(config.GlobalConfig.GetString("bot.commandPrefix"))
	if len(prefix) == 0 {
		prefix = "/"
	}
	return prefix
}

func GetEmitInterval() time.Duration {
	if config.GlobalConfig == nil {
		return 0
	}
	return config.GlobalConfig.GetDuration("concern.emitInterval")
}

func GetLargeNotifyLimit() int {
	if config.GlobalConfig == nil {
		return 50
	}
	var limit = config.GlobalConfig.GetInt("dispatch.largeNotifyLimit")
	if limit <= 0 {
		limit = 50
	}
	return limit
}

func GetCustomGroupCommand() []string {
	if config.GlobalConfig == nil {
		return nil
	}
	return config.GlobalConfig.GetStringSlice("autoreply.group.command")
}

func GetCustomPrivateCommand() []string {
	if config.GlobalConfig == nil {
		return nil
	}
	return config.GlobalConfig.GetStringSlice("autoreply.private.command")
}
