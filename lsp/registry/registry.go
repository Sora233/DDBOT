package registry

import (
	"github.com/Logiase/MiraiGo-Template/utils"
	"github.com/Sora233/DDBOT/lsp/concern"
)

var logger = utils.GetModuleLogger("registry")

type option struct {
}

type OptFunc func(opt *option) *option

type ConcernCenter struct {
	M map[string]map[concern.Type]concern.Concern
}

var globalCenter = newConcernCenter()

func newConcernCenter() *ConcernCenter {
	cc := new(ConcernCenter)
	cc.M = make(map[string]map[concern.Type]concern.Concern)
	return cc
}

func RegisterConcernManager(c concern.Concern, site string, concernType []concern.Type, opts ...OptFunc) {
	if _, found := globalCenter.M[site]; !found {
		globalCenter.M[site] = make(map[concern.Type]concern.Concern)
	}
	for _, ctype := range concernType {
		if lastC, found := globalCenter.M[site][ctype]; !found {
			globalCenter.M[site][ctype] = c
		} else {
			logger.Errorf("Concern %v - Site %v and Type %v is already registered by Concern %v", c.Name(), site, ctype, lastC.Name())
		}
	}
}
