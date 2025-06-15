package cfg

import (
	"errors"
	"os"
	"strings"
	"time"

	"github.com/ghodss/yaml"
	"github.com/spf13/cast"
	"go.uber.org/atomic"

	"github.com/Sora233/MiraiGo-Template/config"
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

var customCommandPrefixAtomic atomic.Value

// ReloadCustomCommandPrefix TODO wtf
func ReloadCustomCommandPrefix() {
	var result map[string]string
	defer func() {
		customCommandPrefixAtomic.Store(result)
	}()
	data, err := os.ReadFile("application.yaml")
	if err != nil {
		return
	}
	var all = make(map[string]interface{})

	err = yaml.Unmarshal(data, &all)
	if err != nil {
		return
	}
	var a = all["customCommandPrefix"]
	if a == nil {
		return
	}
	result = cast.ToStringMapString(a)
}

func GetCustomCommandPrefix() map[string]string {
	var m = customCommandPrefixAtomic.Load()
	if m == nil {
		m = make(map[string]string)
	}
	return m.(map[string]string)
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

type CronJob struct {
	Cron         string `yaml:"cron"`
	TemplateName string `yaml:"templateName"`
	Target       struct {
		Group   []uint32 `yaml:"group"`
		Private []uint32 `yaml:"private"`
	} `yaml:"target"`
}

func GetCronJob() []*CronJob {
	var result []*CronJob
	if err := config.GlobalConfig.UnmarshalKey("cronjob", &result); err != nil {
		logger.Errorf("GetCronJob UnmarshalKey <cronjob> error %v", err)
		return nil
	}
	return result
}

func GetTemplateEnabled() bool {
	return config.GlobalConfig.GetBool("template.enable")
}

func GetCustomGroupCommand() []string {
	return config.GlobalConfig.GetStringSlice("autoreply.group.command")
}

func GetCustomPrivateCommand() []string {
	return config.GlobalConfig.GetStringSlice("autoreply.private.command")
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
