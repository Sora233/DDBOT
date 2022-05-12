package twitcasting

import (
	"fmt"
	"github.com/Sora233/DDBOT/lsp/buntdb"
	"github.com/Sora233/DDBOT/lsp/concern"
	"github.com/Sora233/DDBOT/lsp/concern_type"
	"github.com/Sora233/DDBOT/lsp/mmsg"
	"github.com/Sora233/DDBOT/lsp/mmsg/mt"
	"github.com/Sora233/MiraiGo-Template/config"
	"github.com/Sora233/MiraiGo-Template/utils"
	"github.com/nobuf/cas"
	"strings"
	"sync"
)

var logger = utils.GetModuleLogger("twitcasting-concern")

var (
	Site                        = "twitcasting"
	Live      concern_type.Type = "live"
	userCache                   = sync.Map{}
)

func (tc *TwitCastConcern) getUserName(id string) (*string, error) {

	if data, ok := userCache.Load(id); ok {
		name := data.(string)
		return &name, nil
	}

	userInfo, err := tc.client.User(id)

	if err != nil {
		return nil, err
	}

	name := userInfo.User.Name

	userCache.Store(id, name)

	return &name, nil

}

func (tc *TwitCastConcern) compareAndUpdateUsername(id string, name string) {

	if data, ok := userCache.Load(id); ok {

		cacheName := data.(string)

		if name == cacheName { // 名字相同
			return
		}

	}

	// 否則更新
	userCache.Store(id, name)

}

type tcStateManager struct {
	*concern.StateManager
}

func (tc *tcStateManager) GetGroupConcernConfig(target mt.Target, id interface{}) concern.IConfig {
	return NewGroupConcernConfig(tc.StateManager.GetConcernConfig(target, id))
}

type TwitCastConcern struct {
	*tcStateManager
	client *cas.Client
}

func (tc *TwitCastConcern) Site() string {
	return Site
}

func (tc *TwitCastConcern) Types() []concern_type.Type {
	return []concern_type.Type{Live}
}

type LastStatus struct {
	Live      bool
	LastMovie int
}

func (tc *TwitCastConcern) getLastStatus(id string) (*LastStatus, bool) {

	// 转换
	id = fmt.Sprintf("%v_%v", strings.ReplaceAll(id, ":", "%"), "lastStatus")

	var status = &LastStatus{}
	err := tc.StateManager.GetJson(id, &status)
	if err != nil {
		return nil, false
	}
	return status, true
}

func (tc *TwitCastConcern) updateLastStatus(id string, status *LastStatus) error {

	// 转换
	id = fmt.Sprintf("%v_%v", strings.ReplaceAll(id, ":", "%"), "lastStatus")

	return tc.StateManager.SetJson(id, status)
}

func (tc *TwitCastConcern) removeLastStatus(id string) error {

	// 转换
	id = fmt.Sprintf("%v_%v", strings.ReplaceAll(id, ":", "%"), "lastStatus")

	_, err := tc.Delete(id, buntdb.IgnoreNotFoundOpt())
	return err
}

func (tc *TwitCastConcern) tcFresh() concern.FreshFunc {

	return tc.EmitQueueFresher(func(p concern_type.Type, id interface{}) ([]concern.Event, error) {
		userId := id.(string)

		// 恢复原本 userId
		userId = strings.ReplaceAll(userId, "%", ":")

		logger.Tracef("正在检测 Twitcasting 用户 (%v) 的直播状态..", userId)

		liveStatus, err := tc.GetIsLive(userId)

		var movieId int

		// 有直播才有目前的直播录像ID
		if liveStatus.Living {
			movieId = liveStatus.Movie.Movie.ID.Int()
		} else {
			movieId = -1
		}

		if err != nil {
			return nil, err
		}

		if p.ContainAny(Live) {

			var currentLive *cas.MovieContainer

			last, ok := tc.getLastStatus(userId)

			// LastMovieId 用来获取先前的直播资讯是否与现在相同
			if ok && last.Live == liveStatus.Living && last.LastMovie == movieId {

				logger.Tracef("%v 的直播状态与上次相同，已略过", userId)
				return nil, nil

			}

			// 以下则需要更新到资料库
			defer func() {

				err := tc.updateLastStatus(userId, &LastStatus{
					liveStatus.Living,
					movieId,
				})

				if err != nil {
					logger.Errorf("更新db时出现错误: %v", err)
					// 之后仍要推送
				} else {
					logger.Tracef("成功在数据库更新 %v 的直播状态", userId)
				}

			}()

			if !ok && !liveStatus.Living { // 沒有先前記錄 + 下播狀態
				logger.Tracef("%v 的初始状态为下播，已略过。", userId)
				return nil, nil
			}

			var username string

			if liveStatus.Living {

				logger.Tracef("检测到 %v 正在直播", userId)

				movie := liveStatus.Movie

				currentLive = movie

				// 每次開播的時候比較快取的名稱和用戶名稱
				tc.compareAndUpdateUsername(userId, movie.Broadcaster.Name)

				username = movie.Broadcaster.Name

			} else {

				logger.Tracef("检测到 %v 已停止直播", userId)

				if name, ok := userCache.Load(userId); ok {
					username = name.(string)
				} else {

					user, err := tc.client.User(userId)

					if err != nil {
						logger.Warnf("刷新用户名时出现错误: %v", err)
						username = userId
					} else {
						username = user.User.Name

						tc.compareAndUpdateUsername(userId, user.User.Name)
					}

				}

			}

			return []concern.Event{
				&LiveEvent{
					Live:  liveStatus.Living,
					Name:  username,
					Movie: currentLive,
					Id:    id.(string),
				},
			}, nil
		}

		return nil, nil
	})
}

func (tc *TwitCastConcern) tcNotifyGenerator() concern.NotifyGeneratorFunc {
	return func(target mt.Target, event concern.Event) []concern.Notify {

		if liveEvent, ok := event.(*LiveEvent); ok {
			return []concern.Notify{
				&LiveNotify{
					target:    target,
					LiveEvent: *liveEvent,
				},
			}
		}

		return nil
	}
}

func (tc *TwitCastConcern) Start() error {
	tc.UseEmitQueue()

	if config.GlobalConfig.Get("twitcasting") == nil {
		return fmt.Errorf("找不到 TwitCasting 配置， TC 订阅将不会启动。")
	}

	tc.client = cas.New(config.GlobalConfig.GetString("twitcasting.clientId"), config.GlobalConfig.GetString("twitcasting.clientSecret"))

	tc.UseFreshFunc(tc.tcFresh())
	tc.UseNotifyGeneratorFunc(tc.tcNotifyGenerator())

	// 检查 config 中的 twitcasting token 是否有效
	if _, err := tc.client.RecommendedLives(); err != nil {
		// 无效 token
		if tcErr, ok := err.(*cas.RequestError); ok && tcErr.Content.Code == 1000 {
			return fmt.Errorf("无效的 TwitCasting API Token, 你请确保你填写了正确的 Twitcasting token 资料")
		} else {
			return err
		}
	}

	return tc.StateManager.Start()
}

func (tc *TwitCastConcern) Stop() {
	tc.StateManager.Stop()
}

func (tc *TwitCastConcern) ParseId(s string) (interface{}, error) {
	// 因为 DDBOT 用 : 作为转换，因此要改成其他
	return strings.ReplaceAll(s, ":", "%"), nil
}

func (tc *TwitCastConcern) Add(ctx mmsg.IMsgCtx, target mt.Target, id interface{}, ctype concern_type.Type) (concern.IdentityInfo, error) {

	userId := id.(string)

	// getUserName 以預先加載到快取
	if _, err := tc.getUserName(strings.ReplaceAll(userId, "%", ":")); err != nil {

		// 屬於 twitcasting 官方的 error response
		if tcErr, ok := err.(*cas.RequestError); ok {
			switch tcErr.Content.Code {
			case 1000:
				return nil, fmt.Errorf("无效的 TwitCasting API Token, 你请确保你填写了正确的 Twitcasting token 资料")
			case 2000:
				return nil, fmt.Errorf("请求过度频繁")
			case 400:
				return nil, fmt.Errorf("请求内容无效")
			case 500:
				return nil, fmt.Errorf("API服务器错误")
			case 404:
				return nil, fmt.Errorf("找不到用户 %v", userId)
			default:
				return nil, fmt.Errorf(tcErr.Content.Message)
			}
		} else {
			return nil, err
		}

	}

	_, err := tc.GetStateManager().AddTargetConcern(target, userId, ctype)

	if err != nil {
		return nil, err
	}

	return tc.Get(id)
}

// Remove 实现删除一个订阅
func (tc *TwitCastConcern) Remove(ctx mmsg.IMsgCtx, target mt.Target, id interface{}, ctype concern_type.Type) (concern.IdentityInfo, error) {
	_, err := tc.GetStateManager().RemoveTargetConcern(target, id.(string), ctype)
	if err != nil {
		return nil, err
	}
	// 移除 数据库中的 状态
	if err = tc.removeLastStatus(id.(string)); err != nil {
		logger.Tracef("数据库移除 %v 的数据失败 (可能不存在)", id.(string))
	} else {
		logger.Tracef("数据库移除 %v 的数据成功", id.(string))
	}
	return tc.Get(id)
}

// Get 实现查询单个订阅的信息
func (tc *TwitCastConcern) Get(id interface{}) (concern.IdentityInfo, error) {
	userId := strings.ReplaceAll(id.(string), "%", ":")
	channelName, err := tc.getUserName(userId)
	if err != nil {
		return nil, err
	}
	name := fmt.Sprintf("%v(%v)", *channelName, userId)
	return concern.NewIdentity(id, name), nil
}

// GetStateManager 返回我们自定义修改过的 concern.IStateManager，让所有修改对框架生效
func (tc *TwitCastConcern) GetStateManager() concern.IStateManager {
	return tc.StateManager
}

// NewConcern 返回一个新的 TwitCastConcern， 推荐像这样将 notify channel 通过参数传进来，方便编写单元测试
// 此处使用的 concern.NewStateManagerWithStringID 适用于 string 类型的id
// 如果 ParseId 中选择了int64类型， 则此处可以选择 concern.NewStateManagerWithInt64ID
func NewConcern(notify chan<- concern.Notify) *TwitCastConcern {
	sm := &tcStateManager{concern.NewStateManagerWithStringID(Site, notify)}
	return &TwitCastConcern{sm, nil}
}

// init 向框架注册这个插件，引用这个插件即可使用
func init() {
	concern.RegisterConcern(NewConcern(concern.GetNotifyChan()))
}
