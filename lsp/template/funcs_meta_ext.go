package template

import "github.com/modern-go/gls"

const (
	metaFlag   = "__is__meta__"
	metaConfig = "__meta__config__"
)

type templateMetaConfig struct {
	Trigger string
}

func metaGls(f func()) {
	gls.WithGls(func() {
		gls.Set(metaFlag, true)
		gls.Set(metaConfig, new(templateMetaConfig))
		f()
	})
}

func isInMeta() bool {
	v := gls.Get(metaFlag)
	return v != nil && v.(bool) == true
}

func ifMeta(f func(*templateMetaConfig)) {
	if isInMeta() {
		f(gls.Get(metaConfig).(*templateMetaConfig))
	} else {
		panic("template: meta func should be called only in meta definition")
	}
}

func triggerFullMatch() string {
	return "full_match"
}

func triggerRegex() string {
	return "regex"
}

func triggerCron(cron string) string {
	return ""
}

func trigger(opt string) {
	ifMeta(func(m *templateMetaConfig) {
		m.Trigger = opt
	})
}
