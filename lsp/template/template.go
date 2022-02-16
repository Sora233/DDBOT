package template

import (
	"github.com/Sora233/DDBOT/lsp/mmsg"
)

//type RenderConfig struct{}

//type Engine interface {
//	Render(format string, boilerplate map[string]interface{}) ([]*mmsg.MSG, error)
//}

type YamlEngine struct {
}

func (t *YamlEngine) Render(formats []string, boilerplate map[string]interface{}) ([]*mmsg.MSG, error) {
	var (
		p      = new(Parser)
		result []*mmsg.MSG
	)
	for _, format := range formats {
		m := mmsg.NewMSG()
		if err := p.Parse(format); err != nil {
			return nil, err
		}
		for p.Next() {
			m.Append(p.Peek().ToElement(boilerplate))
		}
		result = append(result, m)
	}
	return result, nil
}

var yamlEngine = new(YamlEngine)

func YAMLRender(formats []string, boilerplate map[string]interface{}) ([]*mmsg.MSG, error) {
	return yamlEngine.Render(formats, boilerplate)
}

func YAMLRenderByKey(key string, boilerplate map[string]interface{}) ([]*mmsg.MSG, error) {
	formats, err := yamlLoader.LoadTemplate(key)
	if err != nil {
		return nil, err
	}
	return yamlEngine.Render(formats, boilerplate)
}
