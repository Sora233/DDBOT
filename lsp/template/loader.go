package template

import (
	"errors"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"strings"
)

var (
	ErrTemplateNotFound         = errors.New("template not found")
	ErrTemplateNotSupport       = errors.New("template not support")
	ErrTemplateActionNotSupport = errors.New("template action not support")
)

func init() {
	yamlLoader.templateConfig = new(viper.Viper)
	if err := yamlLoader.templateConfig.MergeConfig(strings.NewReader(defaultTemplate)); err != nil {
		panic(err)
	}
}

const defaultTemplate = `
template.command.private.help:
  content:
    - |
      常见订阅用法：
      以作者UID:97505为例
      首先订阅直播信息：/watch 97505
      然后订阅动态信息：/watch -t news 97505
      由于通常动态内容较多，可以选择不推送转发的动态
      /config filter not_type 97505 转发
      还可以选择开启直播推送时@全体成员：
      /config at_all 97505 on
      以及开启下播推送：
      /config offline_notify 97505 on
      BOT还支持更多功能，详细命令介绍请查看命令文档：
      https://gitee.com/sora233/DDBOT/blob/master/EXAMPLE.md
      使用时请把作者UID换成你需要的UID
      当您完成所有配置后，可以使用/silence命令，让bot专注于推送，在群内发言更少
    - |
      B站专栏介绍：https://www.bilibili.com/read/cv10602230
      如果您有任何疑问或者建议，请反馈到唯一指定交流群：755612788

template.command.public.help:
  content:
    - |-
      DDBOT是一个多功能单推专用推送机器人，支持b站、斗鱼、油管、虎牙推送
`

type YamlTemplateLoader struct {
	templateConfig *viper.Viper
}

var yamlLoader = new(YamlTemplateLoader)

type Template struct {
	Action  string   `yaml:"action"`
	Content []string `yaml:"content"`
}

func LoadTemplate(key string) (*Template, error) {
	return yamlLoader.LoadTemplate(key)
}

func (y *YamlTemplateLoader) LoadTemplate(key string) (*Template, error) {
	obj := y.templateConfig.Get(key)
	if obj == nil {
		return nil, ErrTemplateNotFound
	}
	result := new(Template)
	if err := y.templateConfig.UnmarshalKey(key, result); err != nil {
		return nil, ErrTemplateNotSupport
	}
	return result, nil
}

func InitTemplateLoader() {
	v := viper.New()
	v.SetConfigType("yaml")
	v.SetConfigName("template")
	v.AddConfigPath(".")
	v.WatchConfig()
	v.OnConfigChange(func(in fsnotify.Event) {
		if err := v.MergeConfig(strings.NewReader(defaultTemplate)); err != nil {
			panic(err)
		}
		if err := v.MergeInConfig(); err != nil {
			logger.Errorf("读取模板配置失败，将使用默认模板，如果您没有使用模板，请忽略本信息: %v", err)
		}
	})
	if err := v.MergeConfig(strings.NewReader(defaultTemplate)); err != nil {
		panic(err)
	}
	if err := v.MergeInConfig(); err != nil {
		logger.Errorf("读取模板配置失败，将使用默认模板，如果您没有使用模板，请忽略本信息: %v", err)
	}
	yamlLoader.templateConfig = v
}
