package lsp

import (
	"errors"
	"fmt"
	"github.com/Logiase/MiraiGo-Template/bot"
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

type MessageContext struct {
	TextReply         func(text string) interface{}
	Reply             func(sendingMessage *message.SendingMessage) interface{}
	Send              func(sendingMessage *message.SendingMessage) interface{}
	NoPermissionReply func() interface{}
	DisabledReply     func() interface{}
	Lsp               *Lsp
	Log               *logrus.Entry
	Sender            *message.Sender
}

func NewMessageContext() *MessageContext {
	return new(MessageContext)
}

func IList(c *MessageContext, groupCode int64) {
	log := c.Log

	if c.Lsp.PermissionStateManager.CheckGroupCommandDisabled(groupCode, ListCommand) {
		c.DisabledReply()
		return
	}

	var success bool

	listMsg := message.NewSendingMessage()

	{
		// bilibili
		userInfos, ctypes, err := c.Lsp.bilibiliConcern.ListWatching(groupCode, concern.BilibiliNews|concern.BibiliLive)
		if err != nil {
			listMsg.Append(message.NewText("bilibili订阅：\n查询失败\n"))
			log.Errorf("bilibili ListWatching error %v ", err)
		} else {
			success = true
			if len(userInfos) != 0 {
				listMsg.Append(message.NewText("bilibili订阅：\n"))
				for index := range userInfos {
					listMsg.Append(utils.MessageTextf("%v %v %v\n", userInfos[index].Name, userInfos[index].Mid, ctypes[index].Description()))
				}
			}
		}

	}

	{
		// douyu
		info, ctypes, err := c.Lsp.douyuConcern.ListWatching(groupCode, concern.DouyuLive)
		if err != nil {
			listMsg.Append(message.NewText("douyu订阅：\n查询失败\n"))
			log.Errorf("douyu ListWatching error %v ", err)
		} else {
			success = true
			if len(info) != 0 {
				listMsg.Append(message.NewText("douyu订阅：\n"))
				for index := range info {
					listMsg.Append(utils.MessageTextf("%v %v %v\n", info[index].Nickname, info[index].RoomId, ctypes[index].Description()))
				}
			}
		}
	}

	{
		// huya
		info, ctypes, err := c.Lsp.huyaConcern.ListWatching(groupCode, concern.HuyaLive)
		if err != nil {
			listMsg.Append(message.NewText("huya订阅：\n查询失败\n"))
			log.Errorf("huya ListWatching error %v ", err)
		} else {
			success = true
			if len(info) != 0 {
				listMsg.Append(message.NewText("huya订阅：\n"))
				for index := range info {
					listMsg.Append(utils.MessageTextf("%v %v %v\n", info[index].Name, info[index].RoomId, ctypes[index].Description()))
				}
			}
		}
	}

	{
		// youtube
		info, ctypes, err := c.Lsp.youtubeConcern.ListWatching(groupCode, concern.YoutubeLive|concern.YoutubeVideo)
		if err != nil {
			listMsg.Append(message.NewText("ytb订阅：\n查询失败\n"))
			log.Errorf("youtube ListWatching error %v ", err)
		} else {
			success = true
			if len(info) != 0 {
				listMsg.Append(message.NewText("ytb订阅：\n"))
				for index := range info {
					listMsg.Append(utils.MessageTextf("%v %v %v\n", info[index].ChannelName, info[index].ChannelId, ctypes[index].Description()))
				}
			}
		}
	}
	if !success {
		c.TextReply("查询失败，请重试")
	} else {
		if len(listMsg.Elements) == 0 {
			listMsg.Append(message.NewText("暂无订阅，可以使用/watch命令订阅"))
		}
		c.Send(listMsg)
	}
}

func IWatch(c *MessageContext, groupCode int64, id string, site string, watchType concern.Type, remove bool) {
	log := c.Log

	if c.Lsp.PermissionStateManager.CheckGroupCommandDisabled(groupCode, WatchCommand) {
		c.DisabledReply()
		return
	}

	if !c.Lsp.PermissionStateManager.RequireAny(
		permission.AdminRoleRequireOption(c.Sender.Uin),
		permission.GroupAdminRoleRequireOption(groupCode, c.Sender.Uin),
		permission.QQAdminRequireOption(groupCode, c.Sender.Uin),
		permission.GroupCommandRequireOption(groupCode, c.Sender.Uin, WatchCommand),
		permission.GroupCommandRequireOption(groupCode, c.Sender.Uin, UnwatchCommand),
	) {
		c.NoPermissionReply()
		return
	}

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
			userInfo, _ := c.Lsp.bilibiliConcern.FindUser(mid, false)
			if _, err := c.Lsp.bilibiliConcern.Remove(groupCode, mid, watchType); err != nil {
				c.TextReply(fmt.Sprintf("unwatch失败 - %v", err))
			} else {
				log.Debugf("unwatch success")
				if userInfo == nil {
					c.TextReply("unwatch成功")
				} else {
					c.TextReply(fmt.Sprintf("unwatch成功 - bilibili用户 %v", userInfo.Name))
				}
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
		// 其他群关注了同一mid，并且推送过Living，那么给新watch的群也推一份
		defer c.Lsp.bilibiliConcern.GroupWatchNotify(groupCode, mid, watchType)
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
			info, _ := c.Lsp.douyuConcern.FindRoom(mid, false)
			if _, err := c.Lsp.douyuConcern.RemoveGroupConcern(groupCode, mid, watchType); err != nil {
				c.TextReply(fmt.Sprintf("unwatch失败 - %v", err))
			} else {
				log.Debugf("unwatch success")
				if info == nil {
					c.TextReply("unwatch成功")
				} else {
					c.TextReply(fmt.Sprintf("unwatch成功 - 斗鱼用户 %v", info.Nickname))
				}
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
			info, _ := c.Lsp.youtubeConcern.FindInfo(id, false)
			if _, err := c.Lsp.youtubeConcern.RemoveGroupConcern(groupCode, id, watchType); err != nil {
				c.TextReply(fmt.Sprintf("unwatch失败 - %v", err))
			} else {
				log.WithField("id", id).Debugf("unwatch success")
				if info == nil {
					c.TextReply("unwatch成功")
				} else {
					c.TextReply(fmt.Sprintf("unwatch成功 - YTB用户 %v", info.ChannelName))
				}
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
			info, _ := c.Lsp.huyaConcern.FindRoom(id, false)
			if _, err := c.Lsp.huyaConcern.RemoveGroupConcern(groupCode, id, watchType); err != nil {
				c.TextReply(fmt.Sprintf("unwatch失败 - %v", err))
			} else {
				log.WithField("id", id).Debugf("unwatch success")
				if info == nil {
					c.TextReply("unwatch成功")
				} else {
					c.TextReply(fmt.Sprintf("unwatch成功 - 虎牙用户 %v", info.Name))
				}
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

func IEnable(c *MessageContext, groupCode int64, command string, disable bool) {
	var err error
	log := c.Log

	if !c.Lsp.PermissionStateManager.RequireAny(
		permission.AdminRoleRequireOption(c.Sender.Uin),
		permission.GroupAdminRoleRequireOption(groupCode, c.Sender.Uin),
	) {
		c.NoPermissionReply()
		return
	}

	if command == UnwatchCommand {
		command = WatchCommand
	}

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

func IGrantRole(c *MessageContext, groupCode int64, role string, grantTo int64, del bool) {
	var err error
	grantRole := permission.FromString(role)
	log := c.Log.WithField("role", grantRole.String())
	switch grantRole {
	case permission.GroupAdmin:
		if !c.Lsp.PermissionStateManager.RequireAny(
			permission.AdminRoleRequireOption(c.Sender.Uin),
			permission.GroupAdminRoleRequireOption(groupCode, c.Sender.Uin),
		) {
			c.NoPermissionReply()
			return
		}
		if bot.Instance.FindGroup(groupCode).FindMember(grantTo) != nil {
			if del {
				err = c.Lsp.PermissionStateManager.UngrantGroupRole(groupCode, grantTo, grantRole)
			} else {
				err = c.Lsp.PermissionStateManager.GrantGroupRole(groupCode, grantTo, grantRole)
			}
		} else {
			log.Errorf("can not find uin")
			err = errors.New("未找到用户")
		}
	case permission.Admin:
		if !c.Lsp.PermissionStateManager.RequireAny(
			permission.AdminRoleRequireOption(c.Sender.Uin),
		) {
			c.NoPermissionReply()
			return
		}
		if bot.Instance.FindGroup(groupCode).FindMember(grantTo) != nil {
			if del {
				err = c.Lsp.PermissionStateManager.UngrantRole(grantTo, grantRole)
			} else {
				err = c.Lsp.PermissionStateManager.GrantRole(grantTo, grantRole)
			}
		} else {
			log.Errorf("can not find uin")
			err = errors.New("未找到用户")
		}
	default:
		err = errors.New("invalid role")
	}
	if err != nil {
		log.Errorf("grant failed %v", err)
		if err == permission.ErrPermissionExist {
			c.TextReply("失败 - 目标已有该权限")
		} else if err == permission.ErrPermissionNotExist {
			c.TextReply("失败 - 目标未有该权限")
		} else {
			c.TextReply(fmt.Sprintf("失败 - %v", err))
		}
		return
	}
	log.Debug("grant success")
	c.TextReply("成功")
}

func IGrantCmd(c *MessageContext, groupCode int64, command string, grantTo int64, del bool) {
	var err error
	if command == UnwatchCommand {
		command = WatchCommand
	}
	log := c.Log.WithField("command", command)
	if !CheckOperateableCommand(command) {
		log.Errorf("unknown command")
		c.TextReply("失败 - invalid command name")
		return
	}
	if !c.Lsp.PermissionStateManager.RequireAny(
		permission.AdminRoleRequireOption(c.Sender.Uin),
		permission.GroupAdminRoleRequireOption(groupCode, c.Sender.Uin),
		permission.QQAdminRequireOption(groupCode, c.Sender.Uin),
	) {
		c.NoPermissionReply()
		return
	}
	if bot.Instance.FindGroup(groupCode).FindMember(grantTo) != nil {
		if del {
			err = c.Lsp.PermissionStateManager.UngrantPermission(groupCode, grantTo, command)
		} else {
			err = c.Lsp.PermissionStateManager.GrantPermission(groupCode, grantTo, command)
		}
	} else {
		log.Errorf("can not find uin")
		err = errors.New("未找到用户")
	}
	if err != nil {
		log.Errorf("grant failed %v", err)
		if err == permission.ErrPermissionExist {
			c.TextReply("失败 - 目标已有该权限")
		} else if err == permission.ErrPermissionNotExist {
			c.TextReply("失败 - 目标未有该权限")
		} else {
			c.TextReply(fmt.Sprintf("失败 - %v", err))
		}
		return
	}
	log.Debug("grant success")
	c.TextReply("成功")
}

func (ic *MessageContext) requireNotNil(param ...interface{}) error {
	for _, p := range param {
		if p == nil {
			return errNilParam
		}
	}
	return nil
}
