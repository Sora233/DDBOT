package registry

import "github.com/Sora233/DDBOT/lsp/concern"

type option struct {
	TypeName map[string]concern.Type
}

type OptFunc func(opt *option) *option

func MapConcernType(name string, p concern.Type) OptFunc {
	return func(opt *option) *option {
		opt.TypeName[name] = p
		return opt
	}
}

func RegisterConcernManager(c concern.Concern, site string, opts ...OptFunc) {

}
