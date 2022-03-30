package cfg

import (
	"github.com/Sora233/MiraiGo-Template/config"
	"strings"
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
