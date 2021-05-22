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

func IList(c *CommandContext, groupCode int64) {
	err := c.requireNotNil(c.TextReply, c.Send, c.Log, c.Lsp)
	if err != nil {
		c.Log.Errorf("params error %v", err)
		return
	}
	log := c.Log

	listMsg := message.NewSendingMessage()

	{
		// bilibili
		userInfos, ctypes, err := c.Lsp.bilibiliConcern.ListWatching(groupCode, concern.BilibiliNews|concern.BibiliLive)
		if err != nil {
			listMsg.Append(message.NewText("bilibili订阅：\n查询失败\n"))
			log.Errorf("bilibili ListWatching error %v ", err)
		} else if len(userInfos) != 0 {
			listMsg.Append(message.NewText("bilibili订阅：\n"))
			for index := range userInfos {
				listMsg.Append(utils.MessageTextf("%v %v %v\n", userInfos[index].Name, userInfos[index].Mid, ctypes[index].Description()))
			}
		}

	}

	{
		// douyu
		info, ctypes, err := c.Lsp.douyuConcern.ListWatching(groupCode, concern.DouyuLive)
		if err != nil {
			listMsg.Append(message.NewText("douyu订阅：\n查询失败\n"))
			log.Errorf("douyu ListWatching error %v ", err)
		} else if len(info) != 0 {
			listMsg.Append(message.NewText("douyu订阅：\n"))
			for index := range info {
				listMsg.Append(utils.MessageTextf("%v %v %v\n", info[index].Nickname, info[index].RoomId, ctypes[index].Description()))
			}
		}
	}

	{
		// huya
		info, ctypes, err := c.Lsp.huyaConcern.ListWatching(groupCode, concern.HuyaLive)
		if err != nil {
			listMsg.Append(message.NewText("huya订阅：\n查询失败\n"))
			log.Errorf("huya ListWatching error %v ", err)
		} else if len(info) != 0 {
			listMsg.Append(message.NewText("huya订阅：\n"))
			for index := range info {
				listMsg.Append(utils.MessageTextf("%v %v %v\n", info[index].Name, info[index].RoomId, ctypes[index].Description()))
			}
		}
	}

	{
		// youtube
		info, ctypes, err := c.Lsp.youtubeConcern.ListWatching(groupCode, concern.YoutubeLive|concern.YoutubeVideo)
		if err != nil {
			listMsg.Append(message.NewText("ytb订阅：\n查询失败\n"))
			log.Errorf("youtube ListWatching error %v ", err)
		} else if len(info) != 0 {
			listMsg.Append(message.NewText("ytb订阅：\n"))
			for index := range info {
				listMsg.Append(utils.MessageTextf("%v %v %v\n", info[index].ChannelName, info[index].ChannelId, ctypes[index].Description()))
			}
		}
	}
	if len(listMsg.Elements) == 0 {
		listMsg.Append(message.NewText("暂无订阅，可以使用/watch命令订阅"))
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
