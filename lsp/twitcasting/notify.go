package twitcasting

import (
	"fmt"
	"github.com/Sora233/DDBOT/lsp/concern_type"
	"github.com/Sora233/DDBOT/lsp/mmsg"
	"github.com/Sora233/DDBOT/proxy_pool"
	"github.com/Sora233/DDBOT/requests"
	localutils "github.com/Sora233/DDBOT/utils"
	"github.com/Sora233/MiraiGo-Template/config"
	"github.com/nobuf/cas"
	"github.com/sirupsen/logrus"
	"strings"
	"time"
)

type LiveEvent struct {
	Id    string
	Live  bool
	Name  string
	Movie *cas.MovieContainer
}

func (e *LiveEvent) Site() string {
	return Site
}

func (e *LiveEvent) Type() concern_type.Type {
	return Live
}

func (e *LiveEvent) GetUid() interface{} {
	return e.Id
}

func (e *LiveEvent) Logger() *logrus.Entry {
	return logger.WithField("Id", e.Id)
}

type LiveNotify struct {
	groupCode int64
	LiveEvent
}

func (n *LiveNotify) GetGroupCode() int64 {
	return n.groupCode
}

func (n *LiveNotify) ToMessage() *mmsg.MSG {

	user := strings.ReplaceAll(n.Id, "%", ":")
	name := n.Name

	nameStrategy := config.GlobalConfig.GetString("twitcasting.nameStrategy")

	var username string

	switch nameStrategy {
	case "both": // 风控机率大
		username = fmt.Sprintf("%v (%v)", name, user)
	case "name": // 风控机率不大，但还是有机会
		username = name
	case "userid": // 风控机率不大，但还是有机会
		username = user
	default:
		logger.Warnf("未知的显示内容: %v, 将采用用户名称。", nameStrategy)
		username = name
	}

	if !n.Live {
		return mmsg.NewTextf("%v 的 TwitCasting 直播已结束。", username)
	}

	// 无资讯
	if n.Movie == nil {
		return mmsg.NewTextf("%v 正在 TwitCasting 直播: https://twitcasting.tv/%v (直播资讯获取失败)", username, user)
	}

	enabledTitle, enabledCreated, enabledImage :=
		config.GlobalConfig.GetBool("twitcasting.broadcaster.title"),
		config.GlobalConfig.GetBool("twitcasting.broadcaster.created"),
		config.GlobalConfig.GetBool("twitcasting.broadcaster.image")

	message := mmsg.NewTextf("%v 正在 TwitCasting 直播", username)

	if enabledTitle {
		message.Textf("\n标题: %v", n.Movie.Movie.Title)
	}

	if enabledCreated {
		created := time.Unix(int64(n.Movie.Movie.Created), 0)

		startTime := fmt.Sprintf("%v年%v月%v日 - %v时%v分%v秒",
			created.Year(), int(created.Month()), created.Day(),
			created.Hour(), created.Minute(), created.Second(),
		)

		message.Textf("\n开播时间: %v", startTime)
	}

	message.Textf("\n直播间: %v", fmt.Sprintf("https://twitcasting.tv/%v", n.Movie.Broadcaster.ScreenID))

	if enabledImage && n.Movie.Movie.LargeThumbnail != "" {
		message.ImageByUrl(n.Movie.Movie.LargeThumbnail, "\n[直播封面获取失败]", requests.ProxyOption(proxy_pool.PreferOversea))
	}

	return message
}

func (n *LiveNotify) Logger() *logrus.Entry {
	return n.LiveEvent.Logger().WithFields(localutils.GroupLogFields(n.groupCode))
}
