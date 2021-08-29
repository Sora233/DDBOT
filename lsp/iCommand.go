package lsp

import (
	"errors"
	"fmt"
	"github.com/Logiase/MiraiGo-Template/bot"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Sora233/DDBOT/concern"
	"github.com/Sora233/DDBOT/lsp/bilibili"
	localdb "github.com/Sora233/DDBOT/lsp/buntdb"
	"github.com/Sora233/DDBOT/lsp/concern_manager"
	"github.com/Sora233/DDBOT/lsp/douyu"
	"github.com/Sora233/DDBOT/lsp/huya"
	"github.com/Sora233/DDBOT/lsp/permission"
	"github.com/Sora233/DDBOT/lsp/youtube"
	"github.com/Sora233/DDBOT/utils"
	"github.com/sirupsen/logrus"
	"github.com/tidwall/buntdb"
	"strings"
	"time"
)

var errNilParam = errors.New("内部参数错误")

type SourceType int

const (
	SourceTypeGroup SourceType = iota
	SourceTypePrivate
)

type MessageContext struct {
	TextReply           func(text string) interface{}
	Reply               func(sendingMessage *message.SendingMessage) interface{}
	Send                func(sendingMessage *message.SendingMessage) interface{}
	NoPermissionReply   func() interface{}
	DisabledReply       func() interface{}
	GlobalDisabledReply func() interface{}
	Lsp                 *Lsp
	Log                 *logrus.Entry
	Sender              *message.Sender
	Source              SourceType
}

func (c *MessageContext) IsFromPrivate() bool {
	return c.Source == SourceTypePrivate
}

func (c *MessageContext) IsFromGroup() bool {
	return c.Source == SourceTypeGroup
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
		mid, err := bilibili.ParseUid(id)
		if err != nil {
			c.TextReply("失败 - bilibili uid格式错误")
			return
		}
		log = log.WithField("mid", mid)
		if remove {
			// unwatch
			userInfo, _ := c.Lsp.bilibiliConcern.FindUser(mid, false)
			if _, err := c.Lsp.bilibiliConcern.Remove(groupCode, mid, watchType); err != nil {
				if err == buntdb.ErrNotFound {
					c.TextReply(fmt.Sprintf("unwatch失败 - 未找到该用户"))
				} else {
					log.Errorf("concern remove failed %v", err)
					c.TextReply(fmt.Sprintf("unwatch失败 - 内部错误"))
				}
			} else {
				if userInfo == nil {
					c.TextReply("unwatch成功")
				} else {
					log = log.WithField("name", userInfo.Name)
					c.TextReply(fmt.Sprintf("unwatch成功 - bilibili用户 %v", userInfo.Name))
				}
				log.Debugf("unwatch success")
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
		if watchType.ContainAny(concern.BibiliLive) {
			// 其他群关注了同一uid，并且推送过Living，那么给新watch的群也推一份
			liveInfo, _ := c.Lsp.bilibiliConcern.GetLiveInfo(mid)
			if liveInfo != nil && liveInfo.Living() {
				if c.IsFromGroup() {
					defer c.Lsp.bilibiliConcern.GroupWatchNotify(groupCode, mid)
				}
				if c.IsFromPrivate() {
					defer c.TextReply("检测到该用户正在直播，但由于您目前处于私聊模式，因此不会在群内推送本次直播，将在该用户下次直播时推送")
				}
			}
		}
		log = log.WithField("name", userInfo.Name)
		log.Debugf("watch success")
		const followerCap = 50
		if userInfo != nil && userInfo.UserStat != nil && watchType.ContainAny(concern.BibiliLive) && userInfo.UserStat.Follower < followerCap {
			c.TextReply(fmt.Sprintf("watch成功 - Bilibili用户 %v\n注意：检测到该用户粉丝数少于%v，请确认您的订阅目标是否正确，注意使用UID而非直播间ID", userInfo.Name, followerCap))
		} else {
			c.TextReply(fmt.Sprintf("watch成功 - Bilibili用户 %v", userInfo.Name))
		}
	case douyu.Site:
		mid, err := douyu.ParseUid(id)
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
				if err == buntdb.ErrNotFound {
					c.TextReply(fmt.Sprintf("unwatch失败 - 未找到该用户"))
				} else {
					log.Errorf("concern remove failed %v", err)
					c.TextReply(fmt.Sprintf("unwatch失败 - 内部错误"))
				}
			} else {
				if info == nil {
					c.TextReply("unwatch成功")
				} else {
					log = log.WithField("name", info.Nickname)
					c.TextReply(fmt.Sprintf("unwatch成功 - 斗鱼用户 %v", info.Nickname))
				}
				log.Debugf("unwatch success")
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
				if err == buntdb.ErrNotFound {
					c.TextReply(fmt.Sprintf("unwatch失败 - 未找到该用户"))
				} else {
					log.Errorf("concern remove failed %v", err)
					c.TextReply(fmt.Sprintf("unwatch失败 - 内部错误"))
				}
			} else {
				if info == nil {
					c.TextReply("unwatch成功")
				} else {
					log = log.WithField("name", info.ChannelName)
					c.TextReply(fmt.Sprintf("unwatch成功 - YTB用户 %v", info.ChannelName))
				}
				log.Debugf("unwatch success")
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
				if err == buntdb.ErrNotFound {
					c.TextReply(fmt.Sprintf("unwatch失败 - 未找到该用户"))
				} else {
					log.Errorf("concern remove failed %v", err)
					c.TextReply(fmt.Sprintf("unwatch失败 - 内部错误"))
				}
			} else {
				if info == nil {
					c.TextReply("unwatch成功")
				} else {
					log = log.WithField("name", info.Name)
					c.TextReply(fmt.Sprintf("unwatch成功 - 虎牙用户 %v", info.Name))
				}
				log.Debugf("unwatch success")
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

	command = CombineCommand(command)

	if !CheckOperateableCommand(command) {
		log.Errorf("non-operateable command")
		c.TextReply(fmt.Sprintf("失败 - 【%v】无效命令", command))
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
		log.Errorf("enable failed %v", err)
		if err == permission.ErrGlobalDisabled {
			c.GlobalDisabledReply()
			return
		}
		if err == permission.ErrPermissionExist {
			if disable {
				c.TextReply("失败 - 该命令已经禁用过了，请不要重复禁用")
			} else {
				c.TextReply("失败 - 该命令已经启用过了，请不要重复启用")
			}
		} else {
			c.TextReply(fmt.Sprintf("失败 - 内部错误"))
		}
		return
	}
	c.TextReply("成功")
}

func IGrantRole(c *MessageContext, groupCode int64, grantRole permission.RoleType, grantTo int64, del bool) {
	var err error
	log := c.Log.WithField("role", grantRole.String()).WithFields(utils.GroupLogFields(groupCode))
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
		if del {
			err = c.Lsp.PermissionStateManager.UngrantRole(grantTo, grantRole)
		} else {
			err = c.Lsp.PermissionStateManager.GrantRole(grantTo, grantRole)
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
	command = CombineCommand(command)
	log := c.Log.WithField("command", command)
	if !CheckOperateableCommand(command) {
		log.Errorf("unknown command")
		c.TextReply(fmt.Sprintf("失败 - 【%v】无效命令", command))
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
		if err == permission.ErrGlobalDisabled {
			c.GlobalDisabledReply()
			return
		}
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

func IConfigAtCmd(c *MessageContext, groupCode int64, id string, site string, ctype concern.Type, action string, QQ []int64) {
	g := bot.Instance.FindGroup(groupCode)
	if g == nil {
		// 可能没找到吗
		return
	}
	if action != "show" && action != "clear" && len(QQ) == 0 {
		c.TextReply("失败 - 没有要操作的指定QQ号")
		return
	}
	if action == "add" {
		var failed []int64
		for _, qq := range QQ {
			member := g.FindMember(qq)
			if member == nil {
				failed = append(failed, qq)
			}
		}
		if len(failed) != 0 {
			c.TextReply(fmt.Sprintf("失败 - 没有找到QQ号：\n%v", utils.JoinInt64(failed, "\n")))
			return
		}
	}
	err := iConfigCmd(c, groupCode, id, site, ctype, operateAtConcernConfig(c, ctype, action, QQ))
	if localdb.IsRollback(err) || permission.IsPermissionError(err) {
		return
	}
	if err != nil {
		c.TextReply(err.Error())
	} else {
		if action != "show" {
			ReplyUserInfo(c, site, id)
		}
	}
}

func IConfigAtAllCmd(c *MessageContext, groupCode int64, id string, site string, ctype concern.Type, on bool) {
	err := iConfigCmd(c, groupCode, id, site, ctype, operateAtAllConcernConfig(c, ctype, on))
	if localdb.IsRollback(err) || permission.IsPermissionError(err) {
		return
	}
	if err != nil {
		c.TextReply(err.Error())
	} else {
		ReplyUserInfo(c, site, id)
	}
}

func IConfigTitleNotifyCmd(c *MessageContext, groupCode int64, id string, site string, ctype concern.Type, on bool) {
	err := iConfigCmd(c, groupCode, id, site, ctype, operateNotifyConcernConfig(c, ctype, on))
	if localdb.IsRollback(err) || permission.IsPermissionError(err) {
		return
	}
	if err != nil {
		c.TextReply(err.Error())
	} else {
		ReplyUserInfo(c, site, id)
	}
}

func IConfigOfflineNotifyCmd(c *MessageContext, groupCode int64, id string, site string, ctype concern.Type, on bool) {
	err := iConfigCmd(c, groupCode, id, site, ctype, operateOfflineNotifyConcernConfig(c, ctype, on))
	if localdb.IsRollback(err) || permission.IsPermissionError(err) {
		return
	}
	if err != nil {
		c.TextReply(err.Error())
	} else {
		ReplyUserInfo(c, site, id)
	}
}

func IConfigFilterCmdType(c *MessageContext, groupCode int64, id string, site string, ctype concern.Type, types []string) {
	if len(types) == 0 {
		c.TextReply("失败 - 没有指定过滤类型")
		return
	}
	if ctype != concern.BilibiliNews {
		c.TextReply("失败 - 暂不支持")
		return
	}
	var invalid = bilibili.CheckTypeDefine(types)
	if len(invalid) != 0 {
		c.TextReply(fmt.Sprintf("失败 - 未定义的类型：\n%v", strings.Join(invalid, " ")))
		return
	}
	err := iConfigCmd(c, groupCode, id, site, ctype, func(config *concern_manager.GroupConcernConfig) bool {
		config.GroupConcernFilter.Type = concern_manager.FilterTypeType
		filterConfig := &concern_manager.GroupConcernFilterConfigByType{Type: types}
		config.GroupConcernFilter.Config = filterConfig.ToString()
		return true
	})
	if localdb.IsRollback(err) || permission.IsPermissionError(err) {
		return
	}
	if err != nil {
		c.TextReply(err.Error())
	} else {
		ReplyUserInfo(c, site, id)
	}
}

func IConfigFilterCmdNotType(c *MessageContext, groupCode int64, id string, site string, ctype concern.Type, types []string) {
	if len(types) == 0 {
		c.TextReply("失败 - 没有指定过滤类型")
		return
	}
	if ctype != concern.BilibiliNews {
		c.TextReply("失败 - 暂不支持")
		return
	}
	var invalid = bilibili.CheckTypeDefine(types)
	if len(invalid) != 0 {
		c.TextReply(fmt.Sprintf("失败 - 未定义的类型：\n%v", strings.Join(invalid, " ")))
		return
	}

	err := iConfigCmd(c, groupCode, id, site, ctype, func(config *concern_manager.GroupConcernConfig) bool {
		config.GroupConcernFilter.Type = concern_manager.FilterTypeNotType
		filterConfig := &concern_manager.GroupConcernFilterConfigByType{Type: types}
		config.GroupConcernFilter.Config = filterConfig.ToString()
		return true
	})
	if localdb.IsRollback(err) || permission.IsPermissionError(err) {
		return
	}
	if err != nil {
		c.TextReply(err.Error())
	} else {
		ReplyUserInfo(c, site, id)
	}
}

func IConfigFilterCmdText(c *MessageContext, groupCode int64, id string, site string, ctype concern.Type, keywords []string) {
	if len(keywords) == 0 {
		c.TextReply("失败 - 没有指定过滤关键字")
		return
	}
	err := iConfigCmd(c, groupCode, id, site, ctype, func(config *concern_manager.GroupConcernConfig) bool {
		config.GroupConcernFilter.Type = concern_manager.FilterTypeText
		filterConfig := &concern_manager.GroupConcernFilterConfigByText{Text: keywords}
		config.GroupConcernFilter.Config = filterConfig.ToString()
		return true
	})
	if localdb.IsRollback(err) || permission.IsPermissionError(err) {
		return
	}
	if err != nil {
		c.TextReply(err.Error())
	} else {
		ReplyUserInfo(c, site, id)
	}
}

func IConfigFilterCmdClear(c *MessageContext, groupCode int64, id string, site string, ctype concern.Type) {
	err := iConfigCmd(c, groupCode, id, site, ctype, func(config *concern_manager.GroupConcernConfig) bool {
		config.GroupConcernFilter = concern_manager.GroupConcernFilterConfig{}
		return true
	})
	if localdb.IsRollback(err) || permission.IsPermissionError(err) {
		return
	}
	if err != nil {
		c.TextReply(err.Error())
	} else {
		ReplyUserInfo(c, site, id)
	}
}

func IConfigFilterCmdShow(c *MessageContext, groupCode int64, id string, site string, ctype concern.Type) {
	err := iConfigCmd(c, groupCode, id, site, ctype, func(config *concern_manager.GroupConcernConfig) bool {
		if config.GroupConcernFilter.Empty() {
			c.TextReply("当前配置为空")
			return false
		}
		sb := strings.Builder{}
		sb.WriteString("当前配置：\n")
		switch config.GroupConcernFilter.Type {
		case concern_manager.FilterTypeText:
			sb.WriteString("关键字过滤模式：\n")
			filter, err := config.GroupConcernFilter.GetFilterByText()
			if err != nil {
				logger.WithField("filter_config", config.GroupConcernFilter.Config).Errorf("get filter failed %v", err)
				c.TextReply("查询失败 - 内部错误")
				return false
			}
			for _, kw := range filter.Text {
				sb.WriteString(kw)
				sb.WriteRune('\n')
			}
		case concern_manager.FilterTypeType, concern_manager.FilterTypeNotType:
			filter, err := config.GroupConcernFilter.GetFilterByType()
			if err != nil {
				logger.WithField("filter_config", config.GroupConcernFilter.Config).Errorf("get filter failed %v", err)
				c.TextReply("查询失败 - 内部错误")
				return false
			}
			if config.GroupConcernFilter.Type == concern_manager.FilterTypeType {
				sb.WriteString("动态类型过滤模式 - 只推送以下种类的动态：\n")
			} else if config.GroupConcernFilter.Type == concern_manager.FilterTypeNotType {
				sb.WriteString("动态类型过滤模式 - 不推送以下种类的动态：\n")
			}
			for _, tp := range filter.Type {
				sb.WriteString(tp)
				sb.WriteRune('\n')
			}
		}
		c.TextReply(sb.String())
		return false
	})
	if localdb.IsRollback(err) || permission.IsPermissionError(err) {
		return
	}
	if err != nil {
		c.TextReply(err.Error())
	} else {
		ReplyUserInfo(c, site, id)
	}
}

func iConfigCmd(c *MessageContext, groupCode int64, id string, site string, ctype concern.Type, f func(*concern_manager.GroupConcernConfig) bool) (err error) {
	log := c.Log
	if err := configCmdGroupCommonCheck(c, groupCode); err != nil {
		return err
	}
	switch site {
	case bilibili.Site:
		var mid int64
		mid, err = bilibili.ParseUid(id)
		if err != nil {
			log.WithField("id", id).Errorf("parse failed")
			err = errors.New("失败 - bilibili uid格式错误")
			return
		}
		err = c.Lsp.bilibiliConcern.CheckGroupConcern(groupCode, mid, ctype)
		if err != concern_manager.ErrAlreadyExists {
			return errors.New("失败 - 该id尚未watch")
		}
		err = c.Lsp.bilibiliConcern.OperateGroupConcernConfig(groupCode, mid, f)
	case douyu.Site:
		var uid int64
		uid, err = douyu.ParseUid(id)
		if err != nil {
			log.WithField("id", id).Errorf("parse failed")
			return errors.New("失败 - douyu id格式错误")
		}
		err = c.Lsp.douyuConcern.CheckGroupConcern(groupCode, uid, ctype)
		if err != concern_manager.ErrAlreadyExists {
			return errors.New("失败 - 该id尚未watch")
		}
		err = c.Lsp.douyuConcern.OperateGroupConcernConfig(groupCode, uid, f)
	case youtube.Site:
		err = c.Lsp.youtubeConcern.CheckGroupConcern(groupCode, id, ctype)
		if err != concern_manager.ErrAlreadyExists {
			return errors.New("失败 - 该id尚未watch")
		}
		err = c.Lsp.youtubeConcern.OperateGroupConcernConfig(groupCode, id, f)
	case huya.Site:
		err = c.Lsp.huyaConcern.CheckGroupConcern(groupCode, id, ctype)
		if err != concern_manager.ErrAlreadyExists {
			return errors.New("失败 - 该id尚未watch")
		}
		err = c.Lsp.huyaConcern.OperateGroupConcernConfig(groupCode, id, f)
	}
	if err != nil && !localdb.IsRollback(err) {
		log.Errorf("OperateGroupConcernConfig failed %v", err)
		err = errors.New("失败 - 内部错误")
	}
	return
}

func ReplyUserInfo(c *MessageContext, site string, id string) {
	switch site {
	case bilibili.Site:
		mid, err := bilibili.ParseUid(id)
		if err != nil {
			c.Log.Errorf("ReplyUserInfo bilibili got wrong id %v", id)
			return
		}
		userInfo, _ := c.Lsp.bilibiliConcern.GetUserInfo(mid)
		c.TextReply(fmt.Sprintf("成功 - Bilibili用户 %v", userInfo.GetName()))
		c.Log.WithField("name", userInfo.GetName()).Debug("reply user info")
	case douyu.Site:
		mid, err := douyu.ParseUid(id)
		if err != nil {
			c.Log.Errorf("ReplyUserInfo douyu got wrong id %v", id)
			return
		}
		userInfo, _ := c.Lsp.douyuConcern.FindRoom(mid, false)
		c.TextReply(fmt.Sprintf("成功 - 斗鱼用户 %v", userInfo.GetNickname()))
		c.Log.WithField("name", userInfo.GetNickname()).Debug("reply user info")
	case youtube.Site:
		userInfo, _ := c.Lsp.youtubeConcern.FindInfo(id, false)
		c.TextReply(fmt.Sprintf("成功 - YTB用户 %v", userInfo.GetChannelName()))
		c.Log.WithField("name", userInfo.GetChannelName()).Debug("reply user info")
	case huya.Site:
		userInfo, _ := c.Lsp.huyaConcern.FindRoom(id, false)
		c.TextReply(fmt.Sprintf("成功 - 虎牙用户 %v", userInfo.GetName()))
		c.Log.WithField("name", userInfo.GetName()).Debug("reply user info")
	}
}

func configCmdGroupCommonCheck(c *MessageContext, groupCode int64) error {
	if c.Lsp.PermissionStateManager.CheckGroupCommandDisabled(groupCode, ConfigCommand) {
		c.DisabledReply()
		return permission.ErrDisabled
	}

	if !c.Lsp.PermissionStateManager.RequireAny(
		permission.AdminRoleRequireOption(c.Sender.Uin),
		permission.GroupAdminRoleRequireOption(groupCode, c.Sender.Uin),
		permission.QQAdminRequireOption(groupCode, c.Sender.Uin),
		permission.GroupCommandRequireOption(groupCode, c.Sender.Uin, ConfigCommand),
	) {
		c.NoPermissionReply()
		return permission.ErrPermissionDenied
	}
	return nil
}

func operateAtConcernConfig(c *MessageContext, ctype concern.Type, action string, QQ []int64) func(concernConfig *concern_manager.GroupConcernConfig) bool {
	return func(concernConfig *concern_manager.GroupConcernConfig) bool {
		switch action {
		case "add":
			concernConfig.GroupConcernAt.MergeAtSomeoneList(ctype, QQ)
			return true
		case "remove":
			concernConfig.GroupConcernAt.RemoveAtSomeoneList(ctype, QQ)
			return true
		case "clear":
			concernConfig.GroupConcernAt.ClearAtSomeoneList(ctype)
			return true
		case "show":
			qqList := concernConfig.GroupConcernAt.GetAtSomeoneList(ctype)
			if len(qqList) == 0 {
				c.TextReply("当前配置为空")
				return false
			}
			c.TextReply(fmt.Sprintf("当前配置：\n%v", utils.JoinInt64(qqList, "\n")))
			return false
		default:
			c.Log.Errorf("unknown action")
			return false
		}
	}
}

func operateAtAllConcernConfig(c *MessageContext, ctype concern.Type, on bool) func(concernConfig *concern_manager.GroupConcernConfig) bool {
	return func(concernConfig *concern_manager.GroupConcernConfig) bool {
		if concernConfig.GroupConcernAt.CheckAtAll(ctype) {
			if on {
				// 配置@all，但已经配置了
				c.TextReply("失败 - 已经配置过了")
				return false
			} else {
				// 取消配置@all
				concernConfig.GroupConcernAt.AtAll = concernConfig.GroupConcernAt.AtAll.Remove(ctype)
				return true
			}
		} else {
			if !on {
				// 取消配置，但并没有配置
				c.TextReply("失败 - 该配置未设置")
				return false
			} else {
				// 配置@all
				concernConfig.GroupConcernAt.AtAll = concernConfig.GroupConcernAt.AtAll.Add(ctype)
				return true
			}
		}
	}
}

func operateNotifyConcernConfig(c *MessageContext, ctype concern.Type, on bool) func(concernConfig *concern_manager.GroupConcernConfig) bool {
	return func(concernConfig *concern_manager.GroupConcernConfig) bool {
		if concernConfig.GroupConcernNotify.CheckTitleChangeNotify(ctype) {
			if on {
				// 配置推送，但已经配置过了
				c.TextReply("失败 - 已经配置过了")
				return false
			} else {
				// 取消配置推送
				concernConfig.GroupConcernNotify.TitleChangeNotify = concernConfig.GroupConcernNotify.TitleChangeNotify.Remove(ctype)
				return true
			}
		} else {
			if !on {
				// 取消配置，但并没有配置
				c.TextReply("失败 - 该配置未设置")
				return false
			} else {
				// 配置推送
				concernConfig.GroupConcernNotify.TitleChangeNotify = concernConfig.GroupConcernNotify.TitleChangeNotify.Add(ctype)
				return true
			}
		}
	}
}

func operateOfflineNotifyConcernConfig(c *MessageContext, ctype concern.Type, on bool) func(concernConfig *concern_manager.GroupConcernConfig) bool {
	return func(concernConfig *concern_manager.GroupConcernConfig) bool {
		if concernConfig.GroupConcernNotify.CheckOfflineNotify(ctype) {
			if on {
				// 配置推送，但已经配置过了
				c.TextReply("失败 - 已经配置过了")
				return false
			} else {
				// 取消配置推送
				concernConfig.GroupConcernNotify.OfflineNotify = concernConfig.GroupConcernNotify.OfflineNotify.Remove(ctype)
				return true
			}
		} else {
			if !on {
				// 取消配置，但并没有配置
				c.TextReply("失败 - 该配置未设置")
				return false
			} else {
				concernConfig.GroupConcernNotify.OfflineNotify = concernConfig.GroupConcernNotify.OfflineNotify.Add(ctype)
				return true
			}
		}
	}
}

func (c *MessageContext) requireNotNil(param ...interface{}) error {
	for _, p := range param {
		if p == nil {
			return errNilParam
		}
	}
	return nil
}
