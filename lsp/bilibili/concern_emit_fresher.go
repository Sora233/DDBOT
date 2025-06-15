package bilibili

import (
	"time"

	"github.com/Sora233/DDBOT/v2/lsp/concern"
	"github.com/Sora233/DDBOT/v2/lsp/concern_type"
)

func (c *Concern) emitQueueFresher() concern.FreshFunc {
	return c.EmitQueueFresher(func(p concern_type.Type, id interface{}) ([]concern.Event, error) {
		c.SetLastFreshTime(time.Now().Unix())
		mid := id.(int64)
		var result []concern.Event
		for _, subType := range p.Split() {
			if subType.ContainAny(Live) {
				oldInfo, _ := c.FindUserLiving(mid, false)
				newInfo, err := c.FindUserLiving(mid, true)
				if err != nil {
					logger.WithField("mid", mid).Errorf("FindUserLiving error %v", err)
					continue
				}
				c.ClearNotLiveCount(mid)
				if oldInfo == nil {
					newInfo.liveStatusChanged = true
				} else {
					if oldInfo.Living() != newInfo.Living() {
						newInfo.liveStatusChanged = true
					}
					if oldInfo.LiveTitle != newInfo.LiveTitle {
						newInfo.liveTitleChanged = true
					}
				}
				if newInfo.Living() {
					c.MarkLatestActive(mid, time.Now().Unix())
				}
				result = append(result, newInfo)
			}
			if subType.ContainAny(News) {
				newsInfo, err := c.FindUserNews(mid, true)
				if err != nil {
					logger.WithField("mid", mid).Errorf("FindUserNews error %v", err)
					continue
				}
				var cards []*Card
				for _, card := range newsInfo.Cards {
					if c.filterCard(card) {
						cards = append(cards, card)
					}
				}
				newsInfo.Cards = cards
				result = append(result, newsInfo)
			}
		}
		return result, nil
	})
}
