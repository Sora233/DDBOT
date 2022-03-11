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
      999
      888
    - |
      777
      666
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
