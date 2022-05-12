package bilibili

import (
	"context"
	"fmt"
	"github.com/Sora233/DDBOT/lsp/cfg"
	"github.com/Sora233/DDBOT/lsp/concern"
	"github.com/Sora233/DDBOT/lsp/concern_type"
	"github.com/Sora233/DDBOT/lsp/mmsg/mt"
	"github.com/Sora233/MiraiGo-Template/config"
	"github.com/sirupsen/logrus"
	"github.com/tidwall/buntdb"
	"go.uber.org/atomic"
	"golang.org/x/sync/errgroup"
	"strconv"
	"strings"
	"time"
)

// fresh 这个fresh不能启动多个
func (c *Concern) fresh() concern.FreshFunc {
	return func(ctx context.Context, eventChan chan<- concern.Event) {
		t := time.NewTimer(time.Second * 3)
		var interval time.Duration
		if config.GlobalConfig != nil {
			interval = config.GlobalConfig.GetDuration("bilibili.interval")
		}
		if interval == 0 {
			interval = time.Second * 20
		}
		var freshCount atomic.Int32
		if !cfg.GetBilibiliOnlyOnlineNotify() {
			freshCount.Store(1000)
		}
		for {
			select {
			case <-t.C:
			case <-ctx.Done():
				return
			}
			start := time.Now()
			var errGroup errgroup.Group

			errGroup.Go(func() error {
				defer func() {
					logger.WithField("cost", time.Now().Sub(start)).
						Tracef("watchCore dynamic fresh done")
				}()
				newsList, err := c.freshDynamicNew()
				if err != nil {
					logger.Errorf("freshDynamicNew failed %v", err)
					return err
				} else {
					for _, news := range newsList {
						eventChan <- news
					}
				}
				return nil
			})

			errGroup.Go(func() error {
				defer func() {
					logger.WithField("cost", time.Now().Sub(start)).
						Tracef("watchCore live fresh done")
				}()
				liveInfo, err := c.freshLive()
				if err != nil {
					logger.Errorf("freshLive error %v", err)
					return err
				}
				// liveInfoMap内是所有正在直播的列表，没有直播的不应该放进去
				var liveInfoMap = make(map[int64]*LiveInfo)
				for _, info := range liveInfo {
					liveInfoMap[info.Mid] = info
				}

				_, ids, types, err := c.StateManager.ListConcernState(
					func(target mt.Target, id interface{}, p concern_type.Type) bool {
						return p.ContainAny(Live)
					})
				if err != nil {
					logger.Errorf("ListConcernState error %v", err)
					return err
				}
				ids, types, err = c.GroupTypeById(ids, types)
				if err != nil {
					logger.Errorf("GroupTypeById error %v", err)
					return err
				}

				sendLiveInfo := func(info *LiveInfo) {
					addLiveInfoErr := c.AddLiveInfo(info)
					if addLiveInfoErr != nil {
						// 如果因为系统原因add失败，会造成重复推送
						// 按照ddbot的原则，选择不推送，而非重复推送
						logger.WithField("mid", info.Mid).Errorf("add live info error %v", err)
						return
					}
					if (info.Living() && freshCount.Load() < 1) || (!info.Living() && freshCount.Load() < 3) {
						return
					}
					eventChan <- info
				}

				selfUid := accountUid.Load()
				for _, id := range ids {
					mid := id.(int64)
					if selfUid != 0 && selfUid == mid {
						// 特殊处理下关注自己
						accResp, err := XSpaceAccInfo(selfUid)
						if err != nil {
							logger.Errorf("freshLive self-fresh %v error %v", selfUid, err)
							return err
						}
						liveRoom := accResp.GetData().GetLiveRoom()
						selfLiveInfo := NewLiveInfo(
							NewUserInfo(selfUid, liveRoom.GetRoomid(), accResp.GetData().GetName(), liveRoom.GetUrl()),
							liveRoom.GetTitle(),
							liveRoom.GetCover(),
							liveRoom.GetLiveStatus(),
						)
						if selfLiveInfo.Living() {
							liveInfoMap[selfUid] = selfLiveInfo
						}
					}
					oldInfo, _ := c.GetLiveInfo(mid)
					if oldInfo == nil {
						// first live info
						if newInfo, found := liveInfoMap[mid]; found {
							newInfo.liveStatusChanged = true
							sendLiveInfo(newInfo)
						}
						continue
					}
					if oldInfo.Status == LiveStatus_NoLiving {
						if newInfo, found := liveInfoMap[mid]; found {
							// notliving -> living
							newInfo.liveStatusChanged = true
							sendLiveInfo(newInfo)
						}
					} else if oldInfo.Status == LiveStatus_Living {
						if newInfo, found := liveInfoMap[mid]; !found {
							// living -> notliving
							if count := c.IncNotLiveCount(mid); count < 3 {
								logger.WithField("uid", mid).WithField("name", oldInfo.UserInfo.Name).
									WithField("notlive_count", count).
									Trace("notlive counting")
								continue
							} else {
								logger.WithField("uid", mid).WithField("name", oldInfo.UserInfo.Name).
									Debug("notlive count done, notlive confirmed")
							}
							if err := c.ClearNotLiveCount(mid); err != nil {
								logger.WithField("uid", mid).WithField("name", oldInfo.UserInfo.Name).
									Errorf("clear notlive count error %v", err)
							}
							resp, err := XSpaceAccInfo(mid)
							if err != nil {
								logger.WithField("uid", mid).WithField("name", oldInfo.UserInfo.Name).
									Errorf("XSpaceAccInfo error %v", err)
								continue
							}
							if resp.GetData().GetLiveRoom().GetLiveStatus() == LiveStatus_Living {
								continue
							} else {
								logger.WithField("uid", mid).WithField("name", oldInfo.UserInfo.Name).
									Debug("XSpaceAccInfo notlive confirmed")
							}
							newInfo = NewLiveInfo(&oldInfo.UserInfo, oldInfo.LiveTitle,
								oldInfo.Cover, LiveStatus_NoLiving)
							newInfo.liveStatusChanged = true
							sendLiveInfo(newInfo)
						} else {
							if newInfo.LiveTitle == "bilibili主播的直播间" {
								newInfo.LiveTitle = oldInfo.LiveTitle
							}
							if err := c.ClearNotLiveCount(mid); err != nil {
								logger.WithField("uid", mid).WithField("name", oldInfo.UserInfo.Name).
									Errorf("clear notlive count error %v", err)
							}
							if newInfo.LiveTitle != oldInfo.LiveTitle {
								// live title change
								newInfo.liveTitleChanged = true
								sendLiveInfo(newInfo)
							}
						}
					}
				}
				return nil
			})
			err := errGroup.Wait()
			freshCount.Inc()
			end := time.Now()
			if err == nil {
				logger.WithField("cost", end.Sub(start)).Tracef("watchCore loop done")
				c.SetLastFreshTime(time.Now().Unix())
			} else {
				logger.WithField("cost", end.Sub(start)).Errorf("watchCore error %v", err)
			}
			t.Reset(interval)
		}
	}
}

func (c *Concern) freshDynamicNew() ([]*NewsInfo, error) {
	var start = time.Now()
	resp, err := DynamicSvrDynamicNew()
	if err != nil {
		logger.Errorf("DynamicSvrDynamicNew error %v", err)
		return nil, err
	}
	var newsMap = make(map[int64][]*Card)
	if resp.GetCode() != 0 {
		logger.WithField("RespCode", resp.GetCode()).
			WithField("RespMsg", resp.GetMessage()).
			Errorf("DynamicSvrDynamicNew failed")
		return nil, fmt.Errorf("DynamicSvrDynamicNew failed %v - %v", resp.GetCode(), resp.GetMessage())
	}
	var cards []*Card
	cards = append(cards, resp.GetData().GetCards()...)
	// 尝试刷一下历史动态，看看能不能捞一下被审核的动态
	if len(resp.GetData().GetCards()) > 0 {
		var historyResp *DynamicSvrDynamicHistoryResponse
		var lastDynamicId = resp.GetData().GetCards()[len(resp.GetData().GetCards())-1].GetDesc().GetDynamicIdStr()
		for i := 0; i < 2; i++ {
			if len(lastDynamicId) == 0 {
				break
			}
			historyResp, err = DynamicSvrDynamicHistory(lastDynamicId)
			if err != nil {
				logger.WithField("lastDynamicId", lastDynamicId).
					Errorf("DynamicSvrDynamicHistory error %v", err)
				break
			}
			if historyResp.GetCode() != 0 {
				logger.WithField("RespCode", resp.GetCode()).
					WithField("RespMsg", resp.GetMessage()).
					Errorf("DynamicSvrDynamicHistory failed")
				return nil, fmt.Errorf("DynamicSvrDynamicHistory failed %v - %v",
					historyResp.GetCode(), historyResp.GetMessage())
			}
			cards = append(cards, historyResp.GetData().GetCards()...)
			if len(historyResp.GetData().GetCards()) > 0 {
				cardSize := len(historyResp.GetData().GetCards())
				lastDynamicId = historyResp.GetData().GetCards()[cardSize-1].GetDesc().GetDynamicIdStr()
			} else {
				lastDynamicId = ""
			}
		}
	}

	logger.WithField("cost", time.Now().Sub(start)).Trace("freshDynamicNew cost 1")
	for _, card := range cards {
		uid := card.GetDesc().GetUid()
		if c.filterCard(card) {
			newsMap[uid] = append(newsMap[uid], card)
		}
	}
	logger.WithField("cost", time.Now().Sub(start)).Trace("freshDynamicNew cost 2")
	var result []*NewsInfo
	for uid, cards := range newsMap {
		userInfo, err := c.StateManager.GetUserInfo(uid)
		if err == buntdb.ErrNotFound {
			continue
		} else if err != nil {
			logger.WithField("mid", uid).Debugf("find user info error %v", err)
			continue
		}
		if len(cards) > 0 {
			// 如果更新了名字，有机会在这里捞回来
			userInfo.Name = cards[0].GetDesc().GetUserProfile().GetInfo().GetUname()
		}
		if len(cards) > 3 {
			// 有时候b站抽风会刷屏
			cards = cards[:3]
		}
		result = append(result, NewNewsInfoWithDetail(userInfo, cards))
	}
	for _, news := range result {
		_ = c.MarkLatestActive(news.Mid, news.Timestamp)
	}
	logger.WithField("cost", time.Now().Sub(start)).
		WithField("NewsInfo Size", len(result)).
		Trace("freshDynamicNew done")
	return result, nil
}

// return all LiveInfo in LiveStatus_Living
func (c *Concern) freshLive() ([]*LiveInfo, error) {
	var start = time.Now()
	var liveInfo []*LiveInfo
	var infoSet = make(map[int64]bool)
	var page = 1
	var maxPage int32 = 1
	var zeroCount = 0
	for {
		resp, err := FeedList(FeedPageOpt(page))
		if err != nil {
			logger.Errorf("freshLive FeedList error %v", err)
			return nil, err
		} else if resp.GetCode() != 0 {
			if resp.GetCode() == -101 && strings.Contains(resp.GetMessage(), "未登录") {
				logger.Errorf("刷新直播列表失败，可能是cookie失效，将尝试重新获取cookie")
				ClearCookieInfo(username)
				atomicVerifyInfo.Store(new(VerifyInfo))
			} else if resp.GetCode() == -400 {
				logger.Errorf("刷新直播列表失败，可能是自动登陆失败，请查看文档尝试手动设置b站cookie")
			} else {
				logger.Errorf("freshLive FeedList code %v msg %v", resp.GetCode(), resp.GetMessage())
			}
			return nil, fmt.Errorf("freshLive FeedList error code %v msg %v", resp.GetCode(), resp.GetMessage())
		}
		var (
			dataSize    = len(resp.GetData().GetList())
			pageSize, _ = strconv.ParseInt(resp.GetData().GetPagesize(), 10, 32)
			curTotal    = resp.GetData().GetResults()
			curMaxPage  = (curTotal-1)/int32(pageSize) + 1
		)
		logger.WithFields(logrus.Fields{
			"CurTotal":   curTotal,
			"PageSize":   pageSize,
			"CurMaxPage": curMaxPage,
			"maxPage":    maxPage,
			"page":       page,
		}).Trace("freshLive debug")
		if curMaxPage > maxPage {
			maxPage = curMaxPage
		}
		for _, l := range resp.GetData().GetList() {
			if infoSet[l.GetUid()] {
				continue
			}
			infoSet[l.GetUid()] = true
			info := NewLiveInfo(
				NewUserInfo(l.GetUid(), l.GetRoomid(), l.GetUname(), l.GetLink()),
				l.GetTitle(),
				l.GetPic(),
				LiveStatus_Living,
			)
			if info.Cover == "" {
				info.Cover = l.GetCover()
			}
			if info.Cover == "" {
				info.Cover = l.GetFace()
			}
			liveInfo = append(liveInfo, info)
		}
		if dataSize != 0 {
			zeroCount = 0
			page++
		} else {
			zeroCount += 1
		}
		if int32(page) > maxPage {
			break
		}
		if zeroCount >= 3 {
			// 认为是真的无人在直播，可能是关注比较少
			if maxPage > 1 {
				logger.WithFields(logrus.Fields{
					"Page":          page,
					"MaxPage":       maxPage,
					"LiveInfo Size": len(liveInfo),
				}).Errorf("直播信息刷新异常结束，如果该信息没有频繁出现，可以忽略。")
			}
			break
		}
	}
	ts := time.Now().Unix()
	for _, info := range liveInfo {
		_ = c.MarkLatestActive(info.Mid, ts)
	}
	logger.WithFields(logrus.Fields{
		"cost":          time.Since(start),
		"Page":          page,
		"MaxPage":       maxPage,
		"LiveInfo Size": len(liveInfo),
	}).Tracef("freshLive done")
	return liveInfo, nil
}
