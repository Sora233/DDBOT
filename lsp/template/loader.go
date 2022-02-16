package template

import (
	"errors"
	"github.com/spf13/viper"
)

var (
	ErrTemplateNotFound = errors.New("template not found")
	ErrNotInit          = errors.New("not init")
)

//type Loader interface {
//	LoadTemplate(key string) ([]string, error)
//}

var defaultTemplate = map[string][]string{
	"test": {
		"test",
	},
}

type YamlTemplateLoader struct {
	templateConfig *viper.Viper
}

var yamlLoader = new(YamlTemplateLoader)

func (y *YamlTemplateLoader) loadDefault(key string) ([]string, error) {
	if formats, found := defaultTemplate[key]; found {
		return formats, nil
	}
	return nil, ErrTemplateNotFound
}

func (y *YamlTemplateLoader) LoadTemplate(key string) ([]string, error) {
	if y.templateConfig == nil {
		return y.loadDefault(key)
	}
	formats := y.templateConfig.GetStringSlice(key)
	if formats != nil {
		return formats, nil
	}
	return y.loadDefault(key)
}

func InitTemplateLoader() {
	v := viper.New()
	v.SetConfigName("template")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.WatchConfig()
	if err := v.ReadInConfig(); err != nil {
		logger.Infof("读取模板配置失败，将使用默认模板，如果您没有使用模板，请忽略本信息: %v", err)
	}
	yamlLoader.templateConfig = v
}
