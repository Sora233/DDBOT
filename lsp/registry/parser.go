package registry

import (
	"github.com/Sora233/DDBOT/lsp/concern_type"
	"github.com/Sora233/DDBOT/utils"
	"strings"
)

func ParseRawSite(rawSite string) (string, error) {
	var (
		found bool
		site  string
	)

	rawSite = strings.Trim(rawSite, `"`)
	site, found = utils.PrefixMatch(ListSite(), rawSite)
	if !found {
		return "", ErrSiteNotSupported
	}
	return site, nil
}

func ParseRawSiteAndType(rawSite string, rawType string) (string, concern_type.Type, error) {
	var (
		site  string
		_type string
		found bool
		err   error
	)
	rawSite = strings.Trim(rawSite, `"`)
	rawType = strings.Trim(rawType, `"`)
	site, err = ParseRawSite(rawSite)
	if err != nil {
		return "", concern_type.Empty, err
	}
	var sTypes []string
	for _, t := range ListType(site) {
		sTypes = append(sTypes, t.String())
	}
	_type, found = utils.PrefixMatch(sTypes, rawType)
	if !found {
		return "", concern_type.Empty, ErrTypeNotSupported
	}
	return site, concern_type.Type(_type), nil
}
