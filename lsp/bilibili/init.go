package bilibili

import (
	"time"

	"github.com/Sora233/DDBOT/v2/lsp/concern"
)

func init() {
	concern.RegisterConcern(NewConcern(concern.GetNotifyChan()))
	refreshCookieJar()
	refreshNavWbi()
	go func() {
		for range time.Tick(time.Minute * 60) {
			refreshCookieJar()
		}
	}()
	go func() {
		for range time.Tick(2 * time.Minute) {
			refreshNavWbi()
		}
	}()
}
