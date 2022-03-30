package template

import (
	"github.com/Sora233/DDBOT/lsp/mmsg"
	"math/rand"
)

type YamlEngine struct {
}

func (t *YamlEngine) Render(template *Template, boilerplate map[string]interface{}) (*mmsg.MSG, error) {
	var (
		p = new(Parser)
		m = mmsg.NewMSG()
	)
	if template == nil || template.Content == nil {
		return nil, nil
	}
	switch template.Action {
	case "", "text":
		for _, format := range template.Content {
			if err := p.Parse(format); err != nil {
				return nil, err
			}
			for p.Next() {
				m.Append(p.Peek().ToElement(boilerplate))
			}
			m.Cut()
		}
	case "roll":
		if err := p.Parse(template.Content[rand.Intn(len(template.Content))]); err != nil {
			return nil, err
		}
		for p.Next() {
			m.Append(p.Peek().ToElement(boilerplate))
		}
		m.Cut()
	default:
		logger.WithField("template", template.Content).Errorf("unknown render action <%v>", template.Action)
		return nil, ErrTemplateActionNotSupport
	}
	return m, nil
}

var yamlEngine = new(YamlEngine)

func YAMLRender(template *Template, boilerplate map[string]interface{}) (*mmsg.MSG, error) {
	return yamlEngine.Render(template, boilerplate)
}

func YAMLRenderByKey(key string, boilerplate map[string]interface{}) (*mmsg.MSG, error) {
	formats, err := yamlLoader.LoadTemplate(key)
	if err != nil {
		return nil, err
	}
	return yamlEngine.Render(formats, boilerplate)
}
