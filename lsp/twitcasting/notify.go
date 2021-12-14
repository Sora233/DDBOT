package twitcasting

import (
	"fmt"
	"github.com/Sora233/DDBOT/lsp/concern_type"
	"github.com/Sora233/DDBOT/lsp/mmsg"
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
	User  *cas.User
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
	name := n.User.Name

	username := fmt.Sprintf("%v (%v)", name, user)

	if !n.Live {

		return mmsg.NewTextf("%v 的 TwitCasting 直播已结束。", username)
	}

	// 无资讯
	if n.Movie == nil {
		return mmsg.NewTextf("%v 正在 TwitCasting 直播: https://twitcasting.tv/%v (直播资讯获取失败)", username, user)
	}

	enableTitle, enabledCreated, enabledImage :=
		config.GlobalConfig.GetBool("twitcasting.broadcaster.title"),
		config.GlobalConfig.GetBool("twitcasting.broadcaster.created"),
		config.GlobalConfig.GetBool("twitcasting.broadcaster.image")

	message := fmt.Sprintf("%v 正在 TwitCasting 直播", username)

	if enableTitle {
		message += fmt.Sprintf("\n标题: %v", n.Movie.Movie.Title)
	}

	if enabledCreated {
		created := time.Unix(int64(n.Movie.Movie.Created), 0)

		startTime := fmt.Sprintf("%v年%v月%v日 - %v时%v分%v秒",
			created.Year(), int(created.Month()), created.Day(),
			created.Hour(), created.Minute(), created.Second(),
		)

		message += fmt.Sprintf("\n开播时间: %v", startTime)
	}

	message += fmt.Sprintf("\n直播间: %v", fmt.Sprintf("https://twitcasting.tv/%v", n.Movie.Broadcaster.ScreenID))

	msg := mmsg.NewText(message)

	if enabledImage && n.Movie.Movie.LargeThumbnail != "" {
		msg.Append(mmsg.NewImageByUrl(n.Movie.Movie.LargeThumbnail))
	}

	return msg
}

func (n *LiveNotify) Logger() *logrus.Entry {
	return n.LiveEvent.Logger().WithFields(localutils.GroupLogFields(n.groupCode))
}
