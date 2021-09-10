package lsp

import (
	"errors"
	"fmt"
	"github.com/Logiase/MiraiGo-Template/bot"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Sora233/DDBOT/lsp/bilibili"
	localdb "github.com/Sora233/DDBOT/lsp/buntdb"
	"github.com/Sora233/DDBOT/lsp/concern"
	"github.com/Sora233/DDBOT/lsp/concern_type"
	"github.com/Sora233/DDBOT/lsp/msg"
	"github.com/Sora233/DDBOT/lsp/permission"
	"github.com/Sora233/DDBOT/lsp/registry"
	"github.com/Sora233/DDBOT/utils"
	"github.com/tidwall/buntdb"
	"strings"
	"time"
)

var errNilParam = errors.New("内部参数错误")

func IList(c *MessageContext, groupCode int64) {
	if c.Lsp.PermissionStateManager.CheckGroupCommandDisabled(groupCode, ListCommand) {
		c.DisabledReply()
		return
	}

	var empty = true

	listMsg := msg.NewMSG()

	for _, c := range registry.ListConcernManager() {
		infos, ctypes, err := c.List(groupCode, concern_type.Empty)
		if err != nil {
			listMsg.Textf("%v订阅查询失败 - %v\n", c.Site(), err)
		} else {
			if len(infos) > 0 {
				empty = false
				listMsg.Textf("%v订阅：\n", c.Site())
				for index, info := range infos {
					listMsg.Textf("%v %v %v\n", info.GetName(), info.GetUid(), ctypes[index].String())
				}
				listMsg.Text("\n")
			}
		}
	}

	if empty {
		listMsg.Append(message.NewText("暂无订阅，可以使用/watch命令订阅"))
	}
	c.Send(listMsg)
}

func IWatch(c *MessageContext, groupCode int64, id string, site string, watchType concern_type.Type, remove bool) {
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

	cm := registry.GetConcernManager(site, watchType)

	mid, err := cm.ParseId(id)
	if err != nil {
		log.Errorf("Parseid error %v", err)
		c.TextReply(fmt.Sprintf("失败 - 解析%v id格式错误", cm.Site()))
		return
	}
	log = log.WithField("mid", mid)
	if remove {
		// unwatch
		userInfo, _ := cm.Get(mid)
		if _, err := cm.Remove(c, groupCode, mid, watchType); err != nil {
			if err == buntdb.ErrNotFound {
				c.TextReply(fmt.Sprintf("unwatch失败 - 未找到该用户"))
			} else {
				log.Errorf("site %v remove failed %v", site, err)
				c.TextReply(fmt.Sprintf("unwatch失败 - 内部错误"))
			}
		} else {
			if userInfo == nil {
				c.TextReply("unwatch成功")
			} else {
				log = log.WithField("name", userInfo.GetName())
				c.TextReply(fmt.Sprintf("unwatch成功 - %v用户 %v", site, userInfo.GetName()))
			}
			log.Debugf("unwatch success")
		}
		return
	}
	// watch
	userInfo, err := cm.Add(c, groupCode, mid, watchType)
	if err != nil {
		log.Errorf("watch error %v", err)
		c.TextReply(fmt.Sprintf("watch失败 - %v", err))
		return
	}
	c.TextReply(fmt.Sprintf("watch成功 - %v用户 %v", site, userInfo))
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

func IConfigAtCmd(c *MessageContext, groupCode int64, id string, site string, ctype concern_type.Type, action string, QQ []int64) {
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
			ReplyUserInfo(c, id, site, ctype)
		}
	}
}

func IConfigAtAllCmd(c *MessageContext, groupCode int64, id string, site string, ctype concern_type.Type, on bool) {
	err := iConfigCmd(c, groupCode, id, site, ctype, operateAtAllConcernConfig(c, ctype, on))
	if localdb.IsRollback(err) || permission.IsPermissionError(err) {
		return
	}
	if err != nil {
		c.TextReply(err.Error())
	} else {
		ReplyUserInfo(c, id, site, ctype)
	}
}

func IConfigTitleNotifyCmd(c *MessageContext, groupCode int64, id string, site string, ctype concern_type.Type, on bool) {
	err := iConfigCmd(c, groupCode, id, site, ctype, operateNotifyConcernConfig(c, ctype, on))
	if localdb.IsRollback(err) || permission.IsPermissionError(err) {
		return
	}
	if err != nil {
		c.TextReply(err.Error())
	} else {
		ReplyUserInfo(c, id, site, ctype)
	}
}

func IConfigOfflineNotifyCmd(c *MessageContext, groupCode int64, id string, site string, ctype concern_type.Type, on bool) {
	err := iConfigCmd(c, groupCode, id, site, ctype, operateOfflineNotifyConcernConfig(c, ctype, on))
	if localdb.IsRollback(err) || permission.IsPermissionError(err) {
		return
	}
	if err != nil {
		c.TextReply(err.Error())
	} else {
		ReplyUserInfo(c, id, site, ctype)
	}
}

func IConfigFilterCmdType(c *MessageContext, groupCode int64, id string, site string, ctype concern_type.Type, types []string) {
	if len(types) == 0 {
		c.TextReply("失败 - 没有指定过滤类型")
		return
	}
	if ctype != concern_manager.BilibiliNews {
		c.TextReply("失败 - 暂不支持")
		return
	}
	var invalid = bilibili.CheckTypeDefine(types)
	if len(invalid) != 0 {
		c.TextReply(fmt.Sprintf("失败 - 未定义的类型：\n%v", strings.Join(invalid, " ")))
		return
	}
	err := iConfigCmd(c, groupCode, id, site, ctype, func(config concern.IConfig) bool {
		config.GetGroupConcernFilter().Type = concern.FilterTypeType
		filterConfig := &concern.GroupConcernFilterConfigByType{Type: types}
		config.GetGroupConcernFilter().Config = filterConfig.ToString()
		return true
	})
	if localdb.IsRollback(err) || permission.IsPermissionError(err) {
		return
	}
	if err != nil {
		c.TextReply(err.Error())
	} else {
		ReplyUserInfo(c, id, site, ctype)
	}
}

func IConfigFilterCmdNotType(c *MessageContext, groupCode int64, id string, site string, ctype concern_type.Type, types []string) {
	if len(types) == 0 {
		c.TextReply("失败 - 没有指定过滤类型")
		return
	}
	if ctype != concern_manager.BilibiliNews {
		c.TextReply("失败 - 暂不支持")
		return
	}
	var invalid = bilibili.CheckTypeDefine(types)
	if len(invalid) != 0 {
		c.TextReply(fmt.Sprintf("失败 - 未定义的类型：\n%v", strings.Join(invalid, " ")))
		return
	}

	err := iConfigCmd(c, groupCode, id, site, ctype, func(config concern.IConfig) bool {
		config.GetGroupConcernFilter().Type = concern.FilterTypeNotType
		filterConfig := &concern.GroupConcernFilterConfigByType{Type: types}
		config.GetGroupConcernFilter().Config = filterConfig.ToString()
		return true
	})
	if localdb.IsRollback(err) || permission.IsPermissionError(err) {
		return
	}
	if err != nil {
		c.TextReply(err.Error())
	} else {
		ReplyUserInfo(c, id, site, ctype)
	}
}

func IConfigFilterCmdText(c *MessageContext, groupCode int64, id string, site string, ctype concern_type.Type, keywords []string) {
	if len(keywords) == 0 {
		c.TextReply("失败 - 没有指定过滤关键字")
		return
	}
	err := iConfigCmd(c, groupCode, id, site, ctype, func(config concern.IConfig) bool {
		config.GetGroupConcernFilter().Type = concern.FilterTypeText
		filterConfig := &concern.GroupConcernFilterConfigByText{Text: keywords}
		config.GetGroupConcernFilter().Config = filterConfig.ToString()
		return true
	})
	if localdb.IsRollback(err) || permission.IsPermissionError(err) {
		return
	}
	if err != nil {
		c.TextReply(err.Error())
	} else {
		ReplyUserInfo(c, id, site, ctype)
	}
}

func IConfigFilterCmdClear(c *MessageContext, groupCode int64, id string, site string, ctype concern_type.Type) {
	err := iConfigCmd(c, groupCode, id, site, ctype, func(config concern.IConfig) bool {
		*config.GetGroupConcernFilter() = concern.GroupConcernFilterConfig{}
		return true
	})
	if localdb.IsRollback(err) || permission.IsPermissionError(err) {
		return
	}
	if err != nil {
		c.TextReply(err.Error())
	} else {
		ReplyUserInfo(c, id, site, ctype)
	}
}

func IConfigFilterCmdShow(c *MessageContext, groupCode int64, id string, site string, ctype concern_type.Type) {
	err := iConfigCmd(c, groupCode, id, site, ctype, func(config concern.IConfig) bool {
		if config.GetGroupConcernFilter().Empty() {
			c.TextReply("当前配置为空")
			return false
		}
		sb := strings.Builder{}
		sb.WriteString("当前配置：\n")
		switch config.GetGroupConcernFilter().Type {
		case concern.FilterTypeText:
			sb.WriteString("关键字过滤模式：\n")
			filter, err := config.GetGroupConcernFilter().GetFilterByText()
			if err != nil {
				logger.WithField("filter_config", config.GetGroupConcernFilter().Config).Errorf("get filter failed %v", err)
				c.TextReply("查询失败 - 内部错误")
				return false
			}
			for _, kw := range filter.Text {
				sb.WriteString(kw)
				sb.WriteRune('\n')
			}
		case concern.FilterTypeType, concern.FilterTypeNotType:
			filter, err := config.GetGroupConcernFilter().GetFilterByType()
			if err != nil {
				logger.WithField("filter_config", config.GetGroupConcernFilter().Config).Errorf("get filter failed %v", err)
				c.TextReply("查询失败 - 内部错误")
				return false
			}
			if config.GetGroupConcernFilter().Type == concern.FilterTypeType {
				sb.WriteString("动态类型过滤模式 - 只推送以下种类的动态：\n")
			} else if config.GetGroupConcernFilter().Type == concern.FilterTypeNotType {
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
		ReplyUserInfo(c, id, site, ctype)
	}
}

func iConfigCmd(c *MessageContext, groupCode int64, id string, site string, ctype concern_type.Type, f func(config concern.IConfig) bool) (err error) {
	if err := configCmdGroupCommonCheck(c, groupCode); err != nil {
		return err
	}
	cm := registry.GetConcernManager(site, ctype)
	if cm == nil {
		return errors.New("<nil>")
	}
	mid, err := cm.ParseId(id)
	if err != nil {
		return fmt.Errorf("%v解析Id失败 - %v", cm.Site(), err)
	}
	cm.GetStateManager().OperateGroupConcernConfig(groupCode, mid, f)

	//switch site {
	//case bilibili.Site:
	//	var mid int64
	//	mid, err = bilibili.ParseUid(id)
	//	if err != nil {
	//		log.WithField("id", id).Errorf("parse failed")
	//		err = errors.New("失败 - bilibili uid格式错误")
	//		return
	//	}
	//	err = c.Lsp.bilibiliConcern.CheckGroupConcern(groupCode, mid, ctype)
	//	if err != concern.ErrAlreadyExists {
	//		return errors.New("失败 - 该id尚未watch")
	//	}
	//	err = c.Lsp.bilibiliConcern.OperateGroupConcernConfig(groupCode, mid, f)
	//case douyu.Site:
	//	var uid int64
	//	uid, err = douyu.ParseUid(id)
	//	if err != nil {
	//		log.WithField("id", id).Errorf("parse failed")
	//		return errors.New("失败 - douyu id格式错误")
	//	}
	//	err = c.Lsp.douyuConcern.CheckGroupConcern(groupCode, uid, ctype)
	//	if err != concern.ErrAlreadyExists {
	//		return errors.New("失败 - 该id尚未watch")
	//	}
	//	err = c.Lsp.douyuConcern.OperateGroupConcernConfig(groupCode, uid, f)
	//case youtube.Site:
	//	err = c.Lsp.youtubeConcern.CheckGroupConcern(groupCode, id, ctype)
	//	if err != concern.ErrAlreadyExists {
	//		return errors.New("失败 - 该id尚未watch")
	//	}
	//	err = c.Lsp.youtubeConcern.OperateGroupConcernConfig(groupCode, id, f)
	//case huya.Site:
	//	err = c.Lsp.huyaConcern.CheckGroupConcern(groupCode, id, ctype)
	//	if err != concern.ErrAlreadyExists {
	//		return errors.New("失败 - 该id尚未watch")
	//	}
	//	err = c.Lsp.huyaConcern.OperateGroupConcernConfig(groupCode, id, f)
	//}
	//if err != nil && !localdb.IsRollback(err) {
	//	log.Errorf("OperateGroupConcernConfig failed %v", err)
	//	err = errors.New("失败 - 内部错误")
	//}
	return
}

func ReplyUserInfo(c *MessageContext, id string, site string, ctype concern_type.Type) {
	cm := registry.GetConcernManager(site, ctype)
	mid, err := cm.ParseId(id)
	if err != nil {
		c.Log.Errorf("ReplyUserInfo %v got wrong id %v", site, id)
		c.TextReply(fmt.Sprintf("成功 - %v用户", site))
		return
	}
	info, err := cm.Get(mid)
	if err != nil || info == nil {
		c.Log.Errorf("ReplyUserInfo %v Get IdentityInfo error %v", site, err)
		c.TextReply(fmt.Sprintf("成功 - %v用户", site))
		return
	}
	c.Log.WithField("name", info.GetName()).Debug("reply user info")
	c.TextReply(fmt.Sprintf("成功 - %v用户 %v", site, info.GetName()))
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

func operateAtConcernConfig(c *MessageContext, ctype concern_type.Type, action string, QQ []int64) func(concernConfig concern.IConfig) bool {
	return func(concernConfig concern.IConfig) bool {
		switch action {
		case "add":
			concernConfig.GetGroupConcernAt().MergeAtSomeoneList(ctype, QQ)
			return true
		case "remove":
			concernConfig.GetGroupConcernAt().RemoveAtSomeoneList(ctype, QQ)
			return true
		case "clear":
			concernConfig.GetGroupConcernAt().ClearAtSomeoneList(ctype)
			return true
		case "show":
			qqList := concernConfig.GetGroupConcernAt().GetAtSomeoneList(ctype)
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

func operateAtAllConcernConfig(c *MessageContext, ctype concern_type.Type, on bool) func(concernConfig concern.IConfig) bool {
	return func(concernConfig concern.IConfig) bool {
		if concernConfig.GetGroupConcernAt().CheckAtAll(ctype) {
			if on {
				// 配置@all，但已经配置了
				c.TextReply("失败 - 已经配置过了")
				return false
			} else {
				// 取消配置@all
				concernConfig.GetGroupConcernAt().AtAll = concernConfig.GetGroupConcernAt().AtAll.Remove(ctype)
				return true
			}
		} else {
			if !on {
				// 取消配置，但并没有配置
				c.TextReply("失败 - 该配置未设置")
				return false
			} else {
				// 配置@all
				concernConfig.GetGroupConcernAt().AtAll = concernConfig.GetGroupConcernAt().AtAll.Add(ctype)
				return true
			}
		}
	}
}

func operateNotifyConcernConfig(c *MessageContext, ctype concern_type.Type, on bool) func(concernConfig concern.IConfig) bool {
	return func(concernConfig concern.IConfig) bool {
		if concernConfig.GetGroupConcernNotify().CheckTitleChangeNotify(ctype) {
			if on {
				// 配置推送，但已经配置过了
				c.TextReply("失败 - 已经配置过了")
				return false
			} else {
				// 取消配置推送
				concernConfig.GetGroupConcernNotify().TitleChangeNotify = concernConfig.GetGroupConcernNotify().TitleChangeNotify.Remove(ctype)
				return true
			}
		} else {
			if !on {
				// 取消配置，但并没有配置
				c.TextReply("失败 - 该配置未设置")
				return false
			} else {
				// 配置推送
				concernConfig.GetGroupConcernNotify().TitleChangeNotify = concernConfig.GetGroupConcernNotify().TitleChangeNotify.Add(ctype)
				return true
			}
		}
	}
}

func operateOfflineNotifyConcernConfig(c *MessageContext, ctype concern_type.Type, on bool) func(concernConfig concern.IConfig) bool {
	return func(concernConfig concern.IConfig) bool {
		if concernConfig.GetGroupConcernNotify().CheckOfflineNotify(ctype) {
			if on {
				// 配置推送，但已经配置过了
				c.TextReply("失败 - 已经配置过了")
				return false
			} else {
				// 取消配置推送
				concernConfig.GetGroupConcernNotify().OfflineNotify = concernConfig.GetGroupConcernNotify().OfflineNotify.Remove(ctype)
				return true
			}
		} else {
			if !on {
				// 取消配置，但并没有配置
				c.TextReply("失败 - 该配置未设置")
				return false
			} else {
				concernConfig.GetGroupConcernNotify().OfflineNotify = concernConfig.GetGroupConcernNotify().OfflineNotify.Add(ctype)
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
