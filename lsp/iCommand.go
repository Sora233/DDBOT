package lsp

import (
	"errors"
	"fmt"
	localdb "github.com/Sora233/DDBOT/lsp/buntdb"
	"github.com/Sora233/DDBOT/lsp/concern"
	"github.com/Sora233/DDBOT/lsp/concern_type"
	"github.com/Sora233/DDBOT/lsp/mmsg"
	"github.com/Sora233/DDBOT/lsp/permission"
	"github.com/Sora233/DDBOT/utils"
	"github.com/Sora233/sliceutil"
	"github.com/sirupsen/logrus"
	"github.com/tidwall/buntdb"
	"sort"
	"strings"
)

func IList(c *MessageContext, groupCode int64, site string) {
	if c.Lsp.PermissionStateManager.CheckGroupCommandDisabled(groupCode, ListCommand) {
		c.DisabledReply()
		return
	}

	listMsg := mmsg.NewMSG()

	var first = true
	var targetCM []concern.Concern
	var info concern.IdentityInfo

	if len(site) > 0 {
		cm, err := concern.GetConcernByParseSite(site)
		if err != nil {
			c.TextReply(fmt.Sprintf("失败 - %v", err))
			return
		}
		targetCM = append(targetCM, cm)
	} else {
		targetCM = concern.ListConcern()
	}
	for _, c := range targetCM {
		_, ids, ctypes, err := c.GetStateManager().ListConcernState(func(_groupCode int64, _ interface{}, _ concern_type.Type) bool {
			return groupCode == _groupCode
		})
		if err == nil {
			ids, ctypes, err = c.GetStateManager().GroupTypeById(ids, ctypes)
		}
		if err != nil {
			if first {
				first = false
			} else {
				listMsg.Text("\n")
			}
			listMsg.Textf("%v订阅查询失败 - %v", c.Site(), err)
		} else {
			if len(ids) > 0 {
				if first {
					first = false
				} else {
					listMsg.Text("\n")
				}
				listMsg.Textf("%v订阅：", c.Site())
				for index, id := range ids {
					info, err = c.Get(id)
					if err != nil {
						info = concern.NewIdentity(id, "unknown")
					}
					listMsg.Text("\n")
					listMsg.Textf("%v %v %v", info.GetName(), info.GetUid(), ctypes[index].String())
				}

			}
		}
	}

	if len(listMsg.Elements()) == 0 {
		listMsg.Textf("暂无订阅，可以使用%v命令订阅", c.Lsp.CommandShowName(WatchCommand))
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

	cm, err := concern.GetConcernBySiteAndType(site, watchType)
	if err != nil {
		log.Errorf("GetConcernManager error %v", err)
		c.TextReply(fmt.Sprintf("失败 - %v", err))
		return
	}

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
				c.TextReply(fmt.Sprintf("unwatch失败 - %v", err))
			}
		} else {
			if userInfo == nil {
				userInfo = concern.NewIdentity(mid, "未知")
			}
			log.WithField("name", userInfo.GetName()).Debugf("unwatch success")
			c.TextReply(fmt.Sprintf("unwatch成功 - %v用户 %v", site, userInfo.GetName()))
		}
		return
	}
	// watch
	userInfo, err := cm.Add(c, groupCode, mid, watchType)
	if err != nil {
		if err == concern.ErrAlreadyExists {
			log.Errorf("user already watched")
			c.TextReply(fmt.Sprintf("watch失败 - 已经watch过了"))
		} else {
			log.Errorf("watch error %v", err)
			c.TextReply(fmt.Sprintf("watch失败 - %v", err))
		}
		return
	}
	if userInfo == nil {
		userInfo = concern.NewIdentity(mid, "未知")
	}
	log.WithField("name", userInfo.GetName()).Debugf("watch success")
	c.TextReply(fmt.Sprintf("watch成功 - %v用户 %v", site, userInfo.GetName()))
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

	if len(command) == 0 {
		c.TextReply("失败 - 没有指定要操作的命令名")
		log.Errorf("empty command")
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
		err = c.Lsp.PermissionStateManager.EnableGroupCommand(groupCode, command)
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
		if gi := utils.GetBot().FindGroup(groupCode); gi != nil && gi.FindMember(grantTo) != nil {
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

	if !c.Lsp.PermissionStateManager.RequireAny(
		permission.AdminRoleRequireOption(c.Sender.Uin),
		permission.GroupAdminRoleRequireOption(groupCode, c.Sender.Uin),
		permission.QQAdminRequireOption(groupCode, c.Sender.Uin),
	) {
		c.NoPermissionReply()
		return
	}

	if !CheckOperateableCommand(command) {
		log.Errorf("unknown command")
		c.TextReply(fmt.Sprintf("失败 - 【%v】无效命令", command))
		return
	}

	if gi := utils.GetBot().FindGroup(groupCode); gi != nil && gi.FindMember(grantTo) != nil {
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

func ISilenceCmd(c *MessageContext, groupCode int64, delete bool) {
	var err error
	if groupCode == 0 {
		if !c.Lsp.PermissionStateManager.RequireAny(
			permission.AdminRoleRequireOption(c.Sender.Uin),
		) {
			c.NoPermissionReply()
			return
		}
		if delete {
			err = c.Lsp.PermissionStateManager.UndoGlobalSilence()
		} else {
			err = c.Lsp.PermissionStateManager.GlobalSilence()
		}
		if err == nil {
			c.TextReply("成功")
		} else {
			c.TextReply(fmt.Sprintf("失败 - %v", err))
		}
		return
	}

	if !c.Lsp.PermissionStateManager.RequireAny(
		permission.AdminRoleRequireOption(c.Sender.Uin),
		permission.GroupAdminRoleRequireOption(groupCode, c.Sender.Uin),
	) {
		c.NoPermissionReply()
		return
	}

	if c.Lsp.PermissionStateManager.CheckGlobalSilence() {
		c.TextReply("失败 - 管理员已开启全局设置，无法操作")
		return
	}

	if delete {
		err = c.Lsp.PermissionStateManager.UndoGroupSilence(groupCode)
	} else {
		err = c.Lsp.PermissionStateManager.GroupSilence(groupCode)
	}
	if err == nil {
		c.TextReply("成功")
	} else {
		c.TextReply(fmt.Sprintf("失败 - %v", err))
	}
}

func IConfigAtCmd(c *MessageContext, groupCode int64, id string, site string, ctype concern_type.Type, action string, QQ []int64) {
	err := configCmdGroupCommonCheck(c, groupCode)
	if err == nil {
		if action != "show" && action != "clear" && len(QQ) == 0 {
			c.TextReply("失败 - 没有要操作的指定QQ号")
			return
		}
		if action == "add" {
			g := utils.GetBot().FindGroup(groupCode)
			if g == nil {
				c.TextReply("失败 - 无法找到这个群的信息，如果看到这个信息表示bot出现了一些问题")
				// 可能没找到吗
				return
			}
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
		err = iConfigCmd(c, groupCode, id, site, ctype, operateAtConcernConfig(c, ctype, action, QQ))
	}
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
	err := configCmdGroupCommonCheck(c, groupCode)
	if err == nil {
		if len(types) == 0 {
			c.TextReply("失败 - 没有指定过滤类型")
			return
		}
		err = iConfigCmd(c, groupCode, id, site, ctype, func(config concern.IConfig) bool {
			config.GetGroupConcernFilter().Type = concern.FilterTypeType
			filterConfig := &concern.GroupConcernFilterConfigByType{Type: types}
			config.GetGroupConcernFilter().Config = filterConfig.ToString()
			return true
		})
	}
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
	err := configCmdGroupCommonCheck(c, groupCode)
	if err == nil {

		if len(types) == 0 {
			c.TextReply("失败 - 没有指定过滤类型")
			return
		}
		err = iConfigCmd(c, groupCode, id, site, ctype, func(config concern.IConfig) bool {
			config.GetGroupConcernFilter().Type = concern.FilterTypeNotType
			filterConfig := &concern.GroupConcernFilterConfigByType{Type: types}
			config.GetGroupConcernFilter().Config = filterConfig.ToString()
			return true
		})
	}
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
	err := configCmdGroupCommonCheck(c, groupCode)
	if err == nil {
		if len(keywords) == 0 {
			c.TextReply("失败 - 没有指定过滤关键字")
			return
		}
		err = iConfigCmd(c, groupCode, id, site, ctype, func(config concern.IConfig) bool {
			config.GetGroupConcernFilter().Type = concern.FilterTypeText
			filterConfig := &concern.GroupConcernFilterConfigByText{Text: keywords}
			config.GetGroupConcernFilter().Config = filterConfig.ToString()
			return true
		})
	}
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
	if err = configCmdGroupCommonCheck(c, groupCode); err != nil {
		return err
	}
	if !sliceutil.Contains(concern.ListSite(), site) {
		return concern.ErrSiteNotSupported
	}
	cm, err := concern.GetConcernBySiteAndType(site, ctype)
	if err != nil {
		c.GetLog().Errorf("GetConcernManager error %v", err)
		c.TextReply(fmt.Sprintf("失败 - %v", err))
		return
	}
	mid, err := cm.ParseId(id)
	if err != nil {
		return fmt.Errorf("%v解析Id失败 - %v", cm.Site(), err)
	}
	err = cm.GetStateManager().CheckGroupConcern(groupCode, mid, ctype)
	if err != concern.ErrAlreadyExists {
		return errors.New("失败 - 该id尚未watch")
	}
	cfg := cm.GetStateManager().GetGroupConcernConfig(groupCode, mid)
	err = cm.GetStateManager().OperateGroupConcernConfig(groupCode, mid, cfg, f)
	if err != nil && !localdb.IsRollback(err) {
		c.GetLog().Errorf("OperateGroupConcernConfig failed %v", err)
		err = fmt.Errorf("失败 - %v", err)
	}
	return
}

func ReplyUserInfo(c *MessageContext, id string, site string, ctype concern_type.Type) {
	cm, err := concern.GetConcernBySiteAndType(site, ctype)
	if err != nil {
		c.GetLog().Errorf("GetConcernManager error %v", err)
		c.TextReply(fmt.Sprintf("失败 - %v", err))
		return
	}
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
			c.TextReply("失败 - 未知操作")
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

func IAbnormalConcernCheck(c *MessageContext) {
	if !c.Lsp.PermissionStateManager.RequireAny(
		permission.AdminRoleRequireOption(c.Sender.Uin),
	) {
		c.NoPermissionReply()
		return
	}

	var allGroups = make(map[int64]bool)
	for _, groups := range utils.GetBot().GetGroupList() {
		allGroups[groups.Code] = true
	}

	var allConcernGroups = make(map[int64]int)
	for _, cm := range concern.ListConcern() {
		_, _, _, err := cm.GetStateManager().ListConcernState(func(groupCode int64, id interface{}, p concern_type.Type) bool {
			allConcernGroups[groupCode] += 1
			return true
		})
		if err != nil {
			c.TextReply(fmt.Sprintf("失败 - %v", err))
			return
		}
	}
	var unknownGroups [][2]int64

	for groupCode, number := range allConcernGroups {
		if _, found := allGroups[groupCode]; !found {
			unknownGroups = append(unknownGroups, [2]int64{groupCode, int64(number)})
		}
	}

	var m = mmsg.NewMSG()

	if len(utils.GetBot().GetGroupList()) > 900 {
		m.Textf("警告：当前账号加入超过900个群，由于QQ本身的限制，可能会将正常的群显示为异常！请注意确认。")
	}

	if len(unknownGroups) == 0 {
		m.Textf("没有查询到异常群号")
	} else {
		// 让结果稳定
		sort.Slice(unknownGroups, func(i, j int) bool {
			return unknownGroups[i][0] < unknownGroups[j][0]
		})
		m.Textf("共查询到%v个异常群号:\n", len(unknownGroups))
		for _, pair := range unknownGroups {
			m.Textf("群 %v - %v个订阅\n", pair[0], pair[1])
		}
		m.Textf("可以使用<%v --abnormal>命令清除异常群订阅", c.Lsp.CommandShowName(CleanConcern))
	}
	c.Send(m)
}

func ICleanConcern(c *MessageContext, abnormal bool, groupCodes []int64, rawSite string, rawType string) {
	log := c.GetLog()

	log = log.WithFields(logrus.Fields{
		"abnormal":    abnormal,
		"group_codes": groupCodes,
		"site":        rawSite,
		"type":        rawType,
	})

	if abnormal {
		if len(groupCodes) != 0 {
			c.TextReply("失败 - 无法同时清除异常订阅和指定群订阅，请重新操作。")
			return
		}
	} else {
		if len(groupCodes) == 0 {
			c.TextReply("失败 - 请指定要清除的群号码")
			return
		}
	}
	type cleanItem struct {
		groupCode int64
		id        interface{}
		tp        concern_type.Type
	}

	type cleanResult struct {
		site string
		tp   concern_type.Type
	}

	var (
		cleanGroupCode = make(map[int64]bool)
		allGroups      = make(map[int64]bool)
		site           string
		err            error
		tp             concern_type.Type
		itemMap        = make(map[string][]*cleanItem)
		result         []*cleanResult
	)

	for _, code := range groupCodes {
		cleanGroupCode[code] = true
	}

	for _, groups := range utils.GetBot().GetGroupList() {
		allGroups[groups.Code] = true
	}

	for _, cm := range concern.ListConcern() {
		if len(rawSite) > 0 {
			site, err = concern.ParseRawSite(rawSite)
			if err != nil {
				c.TextReply(fmt.Sprintf("失败 - %v", err))
				return
			}
			if site != cm.Site() {
				continue
			}
		} else {
			site = cm.Site()
		}
		if len(rawType) > 0 {
			_, tp, err = concern.ParseRawSiteAndType(site, rawType)
			if err != nil {
				continue
			}
		} else {
			tp = concern_type.Empty.Add(cm.Types()...)
		}
		result = append(result, &cleanResult{
			site: site,
			tp:   tp,
		})
		_, _, _, err = cm.GetStateManager().ListConcernState(func(groupCode int64, id interface{}, p concern_type.Type) bool {
			var itp = p.Intersection(tp)
			if itp.Empty() {
				return true
			}
			if abnormal {
				if _, found := allGroups[groupCode]; found {
					return true
				}
			} else {
				if _, found := cleanGroupCode[groupCode]; !found {
					return true
				}
			}
			itemMap[site] = append(itemMap[site], &cleanItem{
				groupCode: groupCode,
				id:        id,
				tp:        itp,
			})
			return true
		})
		if err != nil {
			c.TextReply(fmt.Sprintf("失败 - %v", err))
			return
		}
	}

	var count int
	for site, items := range itemMap {
		cm, err := concern.GetConcernBySite(site)
		if err != nil {
			c.TextReply(fmt.Sprintf("失败 - %v", err))
			return
		}
		for _, item := range items {
			_, err = cm.Remove(c, item.groupCode, item.id, item.tp)
			if err == buntdb.ErrNotFound {
				continue
			} else if err != nil {
				c.TextReply(fmt.Sprintf("失败 - %v", err))
				return
			}
			count++
		}
	}

	c.TextSend(fmt.Sprintf("成功 - 共清除%v个订阅", count))
}
