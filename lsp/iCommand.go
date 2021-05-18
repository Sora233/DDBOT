package lsp

import (
	"errors"
	"fmt"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Sora233/DDBOT/concern"
	"github.com/Sora233/DDBOT/lsp/bilibili"
	"github.com/Sora233/DDBOT/lsp/douyu"
	"github.com/Sora233/DDBOT/lsp/huya"
	"github.com/Sora233/DDBOT/lsp/permission"
	"github.com/Sora233/DDBOT/lsp/youtube"
	"github.com/Sora233/DDBOT/utils"
	"github.com/sirupsen/logrus"
	"strconv"
	"strings"
	"time"
)

var errNilParam = errors.New("内部参数错误")

type CommandContext struct {
	TextReply func(text string) interface{}
	Reply     func(sendingMessage *message.SendingMessage) interface{}
	Send      func(sendingMessage *message.SendingMessage) interface{}
	Lsp       *Lsp
	Log       *logrus.Entry
}

func NewCommandContext() *CommandContext {
	return new(CommandContext)
}

func IList(c *CommandContext, groupCode int64, site string, ctype concern.Type) {
	err := c.requireNotNil(c.TextReply, c.Send, c.Log, c.Lsp)
	if err != nil {
		c.Log.Errorf("params error %v", err)
		return
	}
	log := c.Log

	listMsg := message.NewSendingMessage()

	switch ctype {
	case concern.BibiliLive, concern.BilibiliNews:
		listMsg.Append(message.NewText("当前关注：\n"))
		userInfos, err := c.Lsp.bilibiliConcern.ListWatching(groupCode, ctype)
		if err != nil {
			log.Debugf("list failed %v", err)
			c.TextReply(fmt.Sprintf("list living 失败 - %v", err))
			return
		}
		if len(userInfos) == 0 {
			c.TextReply("关注列表为空，可以使用/watch命令关注")
			return
		}
		for idx, userInfo := range userInfos {
			if idx != 0 {
				listMsg.Append(message.NewText("\n"))
			}
			listMsg.Append(utils.MessageTextf("%v %v", userInfo.Name, userInfo.Mid))
		}
	case concern.DouyuLive:
		listMsg.Append(message.NewText("当前关注：\n"))
		living, err := c.Lsp.douyuConcern.ListLiving(groupCode, true)
		if err != nil {
			log.Debugf("list living failed %v", err)
			c.TextReply(fmt.Sprintf("失败 - %v", err))
			return
		}
		if living == nil {
			c.TextReply("关注列表为空，可以使用/watch命令关注")
			return
		}
		for idx, liveInfo := range living {
			if idx != 0 {
				listMsg.Append(message.NewText("\n"))
			}
			listMsg.Append(utils.MessageTextf("%v %v", liveInfo.Nickname, liveInfo.RoomId))
		}
	case concern.YoutubeLive, concern.YoutubeVideo:
		listMsg.Append(message.NewText("当前关注：\n"))
		userInfos, err := c.Lsp.youtubeConcern.ListWatching(groupCode, ctype)
		if err != nil {
			log.Debugf("list failed %v", err)
			c.TextReply(fmt.Sprintf("失败 - %v", err))
			return
		}
		if len(userInfos) == 0 {
			c.TextReply("关注列表为空，可以使用/watch命令关注")
			return
		}
		for idx, info := range userInfos {
			if idx != 0 {
				listMsg.Append(message.NewText("\n"))
			}
			listMsg.Append(utils.MessageTextf("%v %v", info.ChannelName, info.ChannelId))
		}
	case concern.HuyaLive:
		listMsg.Append(message.NewText("当前关注：\n"))
		living, err := c.Lsp.huyaConcern.ListLiving(groupCode, true)
		if err != nil {
			log.Debugf("list living failed %v", err)
			c.TextReply(fmt.Sprintf("失败 - %v", err))
			return
		}
		if living == nil {
			c.TextReply("关注列表为空，可以使用/watch命令关注")
			return
		}
		for idx, liveInfo := range living {
			if idx != 0 {
				listMsg.Append(message.NewText("\n"))
			}
			listMsg.Append(utils.MessageTextf("%v %v", liveInfo.Name, liveInfo.RoomId))
		}
	}

	c.Send(listMsg)
}

func IWatch(c *CommandContext, groupCode int64, id string, site string, watchType concern.Type, remove bool) {
	err := c.requireNotNil(c.TextReply, c.Log, c.Lsp)
	if err != nil {
		c.Log.Errorf("params error %v", err)
		return
	}
	log := c.Log
	switch site {
	case bilibili.Site:
		id = strings.TrimLeft(id, "UID:")
		mid, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			log.WithField("id", id).Errorf("not a int")
			c.TextReply("失败 - bilibili mid格式错误")
			return
		}
		log = log.WithField("mid", mid)
		if remove {
			// unwatch
			if _, err := c.Lsp.bilibiliConcern.Remove(groupCode, mid, watchType); err != nil {
				c.TextReply(fmt.Sprintf("unwatch失败 - %v", err))
			} else {
				log.Debugf("unwatch success")
				c.TextReply("unwatch成功")
			}
			return
		}
		// watch
		userInfo, err := c.Lsp.bilibiliConcern.Add(groupCode, mid, watchType)
		if err != nil {
			log.Errorf("watch error %v", err)
			c.TextReply(fmt.Sprintf("watch失败 - %v", err))
			return
		}
		log = log.WithField("name", userInfo.Name)
		log.Debugf("watch success")
		c.TextReply(fmt.Sprintf("watch成功 - Bilibili用户 %v", userInfo.Name))
	case douyu.Site:
		mid, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			log.WithField("id", id).Errorf("not a int")
			c.TextReply("失败 - douyu id格式错误")
			return
		}
		log = log.WithField("mid", mid)
		if remove {
			// unwatch
			if _, err := c.Lsp.douyuConcern.RemoveGroupConcern(groupCode, mid, watchType); err != nil {
				c.TextReply(fmt.Sprintf("unwatch失败 - %v", err))
			} else {
				log.Debugf("unwatch success")
				c.TextReply("unwatch成功")
			}
			return
		}
		// watch
		userInfo, err := c.Lsp.douyuConcern.Add(groupCode, mid, watchType)
		if err != nil {
			log.Errorf("watch error %v", err)
			c.TextReply(fmt.Sprintf("watch失败 - %v", err))
			break
		}
		log = log.WithField("name", userInfo.Nickname)
		log.Debugf("watch success")
		c.TextReply(fmt.Sprintf("watch成功 - 斗鱼用户 %v", userInfo.Nickname))
	case youtube.Site:
		log = log.WithField("id", id)
		if remove {
			// unwatch
			if _, err := c.Lsp.youtubeConcern.RemoveGroupConcern(groupCode, id, watchType); err != nil {
				c.TextReply(fmt.Sprintf("unwatch失败 - %v", err))
			} else {
				log.WithField("id", id).Debugf("unwatch success")
				c.TextReply("unwatch成功")
			}
			return
		}
		info, err := c.Lsp.youtubeConcern.Add(groupCode, id, watchType)
		if err != nil {
			log.Errorf("watch error %v", err)
			c.TextReply(fmt.Sprintf("watch失败 - %v", err))
			break
		}
		log = log.WithField("name", info.ChannelName)
		log.Debugf("watch success")
		if info.ChannelName == "" {
			c.TextReply(fmt.Sprintf("watch成功 - YTB用户，该用户未发任何布直播/视频，无法获取名字"))
		} else {
			c.TextReply(fmt.Sprintf("watch成功 - YTB用户 %v", info.ChannelName))
		}
	case huya.Site:
		log = log.WithField("id", id)
		if remove {
			// unwatch
			if _, err := c.Lsp.huyaConcern.RemoveGroupConcern(groupCode, id, watchType); err != nil {
				c.TextReply(fmt.Sprintf("unwatch失败 - %v", err))
			} else {
				log.WithField("id", id).Debugf("unwatch success")
				c.TextReply("unwatch成功")
			}
			return
		}
		info, err := c.Lsp.huyaConcern.Add(groupCode, id, watchType)
		if err != nil {
			log.Errorf("watch error %v", err)
			c.TextReply(fmt.Sprintf("watch失败 - %v", err))
			break
		}
		log = log.WithField("name", info.Name)
		log.Debugf("watch success")
		c.TextReply(fmt.Sprintf("watch成功 - 虎牙用户 %v", info.Name))
	default:
		log.WithField("site", site).Error("unsupported")
		c.TextReply("未支持的网站")
	}
	return
}

func IEnable(c *CommandContext, groupCode int64, command string, disable bool) {
	err := c.requireNotNil(c.TextReply, c.Log, c.Lsp)
	if err != nil {
		c.Log.Errorf("params error %v", err)
		return
	}
	log := c.Log
	if !CheckOperateableCommand(command) {
		log.Errorf("unknown command")
		c.TextReply("失败 - invalid command name")
		return
	}
	if disable {
		err = c.Lsp.PermissionStateManager.DisableGroupCommand(groupCode, command)
	} else {
		if command == ImageContentCommand {
			// 要收钱了，不能白嫖了
			err = c.Lsp.PermissionStateManager.EnableGroupCommand(groupCode, command, permission.ExpireOption(time.Hour*24))
		} else {
			err = c.Lsp.PermissionStateManager.EnableGroupCommand(groupCode, command)
		}
	}
	if err != nil {
		log.Errorf("err %v", err)
		if err == permission.ErrPermissionExist {
			if disable {
				c.TextReply("失败 - 该命令已禁用")
			} else {
				c.TextReply("失败 - 该命令已启用")
			}
		} else {
			c.TextReply(fmt.Sprintf("失败 - %v", err))
		}
		return
	}
	c.TextReply("成功")
}

func (ic *CommandContext) requireNotNil(param ...interface{}) error {
	for _, p := range param {
		if p == nil {
			return errNilParam
		}
	}
	return nil
}
