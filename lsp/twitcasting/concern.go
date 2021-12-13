package twitcasting

import (
	"fmt"
	"github.com/Sora233/DDBOT/lsp/concern"
	"github.com/Sora233/DDBOT/lsp/concern_type"
	"github.com/Sora233/DDBOT/lsp/mmsg"
	"github.com/Sora233/MiraiGo-Template/config"
	"github.com/Sora233/MiraiGo-Template/utils"
	"github.com/nobuf/cas"
	"strings"
)

var logger = utils.GetModuleLogger("twitcasting-concern")

const (
	Site                   = "twitcasting"
	Live concern_type.Type = "live"
)

type tcStateManager struct {
	*concern.StateManager
}

func (tc *tcStateManager) GetGroupConcernConfig(groupCode int64, id interface{}) concern.IConfig {
	return NewGroupConcernConfig(tc.StateManager.GetGroupConcernConfig(groupCode, id))
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

func (tc *TwitCastConcern) tcFresh() concern.FreshFunc {

	lastStatus := make(map[string]bool)

	return tc.EmitQueueFresher(func(p concern_type.Type, id interface{}) ([]concern.Event, error) {
		userId := id.(string)

		// 恢复原本 userId
		userId = strings.ReplaceAll(userId, "%", ":")

		logger.Tracef("正在检测 Twitcasting 用户 (%v) 的直播状态..", userId)

		userInfo, err := tc.client.User(userId)

		if err != nil {
			return nil, err
		}

		if p.ContainAny(Live) {
			var currentLive *cas.MovieContainer

			if userInfo.User.Live {

				logger.Tracef("检测到 %v 正在直播", userInfo.User.ScreenID)

				if last, ok := lastStatus[userId]; ok && last {

					logger.Tracef("%v 的直播状态与上次相同，已略过", userInfo.User.ScreenID)
					return []concern.Event{}, nil

				} else {

					movie, err := tc.client.Movie(userInfo.User.LastMovieID)

					if err != nil {
						return nil, err
					} else if !movie.Movie.Live { // 直播中，但上一次的视频资讯不是直播中
						return nil, fmt.Errorf("无法获取 %v 的直播资讯", userInfo.User.Name)
					}

					currentLive = movie

				}

			}

			lastStatus[userId] = userInfo.User.Live

			return []concern.Event{
				&LiveEvent{
					Live:  userInfo.User.Live,
					Movie: currentLive,
					Id:    id.(string),
				},
			}, nil
		}

		return nil, nil
	})
}

func (tc *TwitCastConcern) tcNotifyGenerator() concern.NotifyGeneratorFunc {
	return func(groupCode int64, event concern.Event) []concern.Notify {

		if liveEvent, ok := event.(*LiveEvent); ok {
			return []concern.Notify{
				&LiveNotify{
					groupCode,
					*liveEvent,
				},
			}
		}

		return []concern.Notify{}
	}
}

func (tc *TwitCastConcern) Start() error {

	tc.client = cas.New(config.GlobalConfig.GetString("twitcasting.clientId"), config.GlobalConfig.GetString("twitcasting.clientSecret"))

	tc.UseFreshFunc(tc.tcFresh())
	tc.UseNotifyGeneratorFunc(tc.tcNotifyGenerator())

	// 检查 config 中的 twitcasting token 是否有效
	if _, err := tc.client.RecommendedLives(); err != nil {
		// 无效 token
		if ifError(err.Error(), "401") {
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

// ifError 因为 cas api 写死用 resp.Status 作为错误传递 (懒得自己包装 tc api)
func ifError(s, err string) bool {
	return strings.HasPrefix(s, fmt.Sprintf("Error: %s", err))
}

func (tc *TwitCastConcern) Add(ctx mmsg.IMsgCtx, groupCode int64, id interface{}, ctype concern_type.Type) (concern.IdentityInfo, error) {

	userId := id.(string)

	if _, err := tc.client.User(strings.ReplaceAll(userId, "%", ":")); err != nil {
		msg := err.Error()
		switch {
		case ifError(msg, "404"):
			return nil, fmt.Errorf("找不到用户 %v", userId)
		case ifError(msg, "403"):
			return nil, fmt.Errorf("无效token或者请求过度频繁")
		case ifError(msg, "400"):
			return nil, fmt.Errorf("无效请求")
		case ifError(msg, "500"):
			return nil, fmt.Errorf("API服务器错误")
		}
	}

	_, err := tc.GetStateManager().AddGroupConcern(groupCode, userId, ctype)

	if err != nil {
		return nil, err
	}

	return tc.Get(id)
}

// Remove 实现删除一个订阅
func (tc *TwitCastConcern) Remove(ctx mmsg.IMsgCtx, groupCode int64, id interface{}, ctype concern_type.Type) (concern.IdentityInfo, error) {
	_, err := tc.GetStateManager().RemoveGroupConcern(groupCode, id.(string), ctype)
	if err != nil {
		return nil, err
	}
	return tc.Get(id)
}

// Get 实现查询单个订阅的信息
func (tc *TwitCastConcern) Get(id interface{}) (concern.IdentityInfo, error) {
	return concern.NewIdentity(id, id.(string)), nil
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
	sm.UseEmitQueue()
	return &TwitCastConcern{sm, nil}
}

// init 向框架注册这个插件，引用这个插件即可使用
func init() {
	concern.RegisterConcern(NewConcern(concern.GetNotifyChan()))
}
