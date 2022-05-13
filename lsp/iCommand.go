package lsp

import (
	"errors"
	"fmt"
	localdb "github.com/Sora233/DDBOT/lsp/buntdb"
	"github.com/Sora233/DDBOT/lsp/concern"
	"github.com/Sora233/DDBOT/lsp/concern_type"
	"github.com/Sora233/DDBOT/lsp/mmsg"
	"github.com/Sora233/DDBOT/lsp/mmsg/mt"
	"github.com/Sora233/DDBOT/lsp/permission"
	"github.com/Sora233/DDBOT/utils"
	"github.com/Sora233/sliceutil"
	"github.com/tidwall/buntdb"
	"strings"
)

func IList(c *MessageContext, target mt.Target, site string) {
	if c.Lsp.PermissionStateManager.CheckTargetCommandDisabled(target, ListCommand) {
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
		_, ids, ctypes, err := c.GetStateManager().ListConcernState(func(_target mt.Target, _ interface{}, _ concern_type.Type) bool {
			return target.Equal(_target)
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

func IWatch(c *MessageContext, target mt.Target, id string, site string, watchType concern_type.Type, remove bool) {
	log := c.Log

	if c.Lsp.PermissionStateManager.CheckTargetCommandDisabled(target, WatchCommand) {
		c.DisabledReply()
		return
	}

	if !c.Lsp.PermissionStateManager.RequireAny(
		permission.AdminRoleRequireOption(c.Sender.Uin()),
		permission.TargetAdminRoleRequireOption(target, c.Sender.Uin()),
		permission.QQAdminRequireOption(target, c.Sender.Uin()),
		permission.TargetCommandRequireOption(target, c.Sender.Uin(), WatchCommand),
		permission.TargetCommandRequireOption(target, c.Sender.Uin(), UnwatchCommand),
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
		if _, err := cm.Remove(c, target, mid, watchType); err != nil {
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
	userInfo, err := cm.Add(c, target, mid, watchType)
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

func IEnable(c *MessageContext, target mt.Target, command string, disable bool) {
	var err error
	log := c.Log

	if c.IsFromPrivate() {
		c.NotImplReply()
		return
	}

	if !c.Lsp.PermissionStateManager.RequireAny(
		permission.AdminRoleRequireOption(c.Sender.Uin()),
		permission.TargetAdminRoleRequireOption(target, c.Sender.Uin()),
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
		err = c.Lsp.PermissionStateManager.DisableTargetCommand(target, command)
	} else {
		err = c.Lsp.PermissionStateManager.EnableTargetCommand(target, command)
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

func IGrantRole(c *MessageContext, target mt.Target, grantRole permission.RoleType, grantTo int64, del bool) {
	if target.GetTargetType().IsPrivate() {
		c.NotImplReply()
		return
	}
	var err error
	log := c.Log.WithField("role", grantRole.String()).WithFields(utils.TargetFields(target))
	switch grantRole {
	case permission.TargetAdmin:
		if !c.Lsp.PermissionStateManager.RequireAny(
			permission.AdminRoleRequireOption(c.Sender.Uin()),
			permission.TargetAdminRoleRequireOption(target, c.Sender.Uin()),
		) {
			c.NoPermissionReply()
			return
		}
		if utils.GetBot().CheckMember(target, grantTo) {
			if del {
				err = c.Lsp.PermissionStateManager.UngrantTargetRole(target, grantTo, grantRole)
			} else {
				err = c.Lsp.PermissionStateManager.GrantTargetRole(target, grantTo, grantRole)
			}
		} else {
			log.Errorf("can not find uin")
			err = errors.New("未找到用户")
		}
	case permission.Admin:
		if !c.Lsp.PermissionStateManager.RequireAny(
			permission.AdminRoleRequireOption(c.Sender.Uin()),
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

func IGrantCmd(c *MessageContext, target mt.Target, command string, grantTo int64, del bool) {
	if target.GetTargetType().IsPrivate() {
		c.NotImplReply()
		return
	}
	var err error
	command = CombineCommand(command)
	log := c.Log.WithField("command", command)

	if !c.Lsp.PermissionStateManager.RequireAny(
		permission.AdminRoleRequireOption(c.Sender.Uin()),
		permission.TargetAdminRoleRequireOption(target, c.Sender.Uin()),
		permission.QQAdminRequireOption(target, c.Sender.Uin()),
	) {
		c.NoPermissionReply()
		return
	}

	if !CheckOperateableCommand(command) {
		log.Errorf("unknown command")
		c.TextReply(fmt.Sprintf("失败 - 【%v】无效命令", command))
		return
	}
	if utils.GetBot().CheckMember(target, grantTo) {
		if del {
			err = c.Lsp.PermissionStateManager.UngrantTargetCommandPermission(target, grantTo, command)
		} else {
			err = c.Lsp.PermissionStateManager.GrantTargetCommandPermission(target, grantTo, command)
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

func ISilenceCmd(c *MessageContext, target mt.Target, global bool, delete bool) {
	var err error
	if global {
		if !c.Lsp.PermissionStateManager.RequireAny(
			permission.AdminRoleRequireOption(c.Sender.Uin()),
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
		permission.AdminRoleRequireOption(c.Sender.Uin()),
		permission.TargetAdminRoleRequireOption(target, c.Sender.Uin()),
	) {
		c.NoPermissionReply()
		return
	}

	if c.Lsp.PermissionStateManager.CheckGlobalSilence() {
		c.TextReply("失败 - 管理员已开启全局设置，无法操作")
		return
	}

	if delete {
		err = c.Lsp.PermissionStateManager.UndoTargetSilence(target)
	} else {
		err = c.Lsp.PermissionStateManager.TargetSilence(target)
	}
	if err == nil {
		c.TextReply("成功")
	} else {
		c.TextReply(fmt.Sprintf("失败 - %v", err))
	}
}

func IConfigAtCmd(c *MessageContext, target mt.Target, id string, site string, ctype concern_type.Type, action string, QQ []int64) {
	err := configCmdTargetCommonCheck(c, target)
	if err == nil {
		if action != "show" && action != "clear" && len(QQ) == 0 {
			c.TextReply("失败 - 没有要操作的指定QQ号")
			return
		}
		if action == "add" {
			if !utils.GetBot().CheckTarget(target) {
				c.TextReply("失败 - 无法找到此源的信息，如果看到这个信息表示bot出现了一些问题")
				// 可能没找到吗
				return
			}
			var failed []int64
			for _, qq := range QQ {
				if !utils.GetBot().CheckMember(target, qq) {
					failed = append(failed, qq)
				}
			}
			if len(failed) != 0 {
				c.TextReply(fmt.Sprintf("失败 - 没有找到QQ号：\n%v", utils.JoinInt64(failed, "\n")))
				return
			}
		}
		err = iConfigCmd(c, target, id, site, ctype, operateAtConcernConfig(c, ctype, action, QQ))
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

func IConfigAtAllCmd(c *MessageContext, target mt.Target, id string, site string, ctype concern_type.Type, on bool) {
	err := iConfigCmd(c, target, id, site, ctype, operateAtAllConcernConfig(c, ctype, on))
	if localdb.IsRollback(err) || permission.IsPermissionError(err) {
		return
	}
	if err != nil {
		c.TextReply(err.Error())
	} else {
		ReplyUserInfo(c, id, site, ctype)
	}
}

func IConfigTitleNotifyCmd(c *MessageContext, target mt.Target, id string, site string, ctype concern_type.Type, on bool) {
	err := iConfigCmd(c, target, id, site, ctype, operateNotifyConcernConfig(c, ctype, on))
	if localdb.IsRollback(err) || permission.IsPermissionError(err) {
		return
	}
	if err != nil {
		c.TextReply(err.Error())
	} else {
		ReplyUserInfo(c, id, site, ctype)
	}
}

func IConfigOfflineNotifyCmd(c *MessageContext, target mt.Target, id string, site string, ctype concern_type.Type, on bool) {
	err := iConfigCmd(c, target, id, site, ctype, operateOfflineNotifyConcernConfig(c, ctype, on))
	if localdb.IsRollback(err) || permission.IsPermissionError(err) {
		return
	}
	if err != nil {
		c.TextReply(err.Error())
	} else {
		ReplyUserInfo(c, id, site, ctype)
	}
}

func IConfigFilterCmdType(c *MessageContext, target mt.Target, id string, site string, ctype concern_type.Type, types []string) {
	err := configCmdTargetCommonCheck(c, target)
	if err == nil {
		if len(types) == 0 {
			c.TextReply("失败 - 没有指定过滤类型")
			return
		}
		err = iConfigCmd(c, target, id, site, ctype, func(config concern.IConfig) bool {
			config.GetConcernFilter().Type = concern.FilterTypeType
			filterConfig := &concern.GroupConcernFilterConfigByType{Type: types}
			config.GetConcernFilter().Config = filterConfig.ToString()
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

func IConfigFilterCmdNotType(c *MessageContext, target mt.Target, id string, site string, ctype concern_type.Type, types []string) {
	err := configCmdTargetCommonCheck(c, target)
	if err == nil {

		if len(types) == 0 {
			c.TextReply("失败 - 没有指定过滤类型")
			return
		}
		err = iConfigCmd(c, target, id, site, ctype, func(config concern.IConfig) bool {
			config.GetConcernFilter().Type = concern.FilterTypeNotType
			filterConfig := &concern.GroupConcernFilterConfigByType{Type: types}
			config.GetConcernFilter().Config = filterConfig.ToString()
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

func IConfigFilterCmdText(c *MessageContext, target mt.Target, id string, site string, ctype concern_type.Type, keywords []string) {
	err := configCmdTargetCommonCheck(c, target)
	if err == nil {
		if len(keywords) == 0 {
			c.TextReply("失败 - 没有指定过滤关键字")
			return
		}
		err = iConfigCmd(c, target, id, site, ctype, func(config concern.IConfig) bool {
			config.GetConcernFilter().Type = concern.FilterTypeText
			filterConfig := &concern.GroupConcernFilterConfigByText{Text: keywords}
			config.GetConcernFilter().Config = filterConfig.ToString()
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

func IConfigFilterCmdClear(c *MessageContext, target mt.Target, id string, site string, ctype concern_type.Type) {
	err := iConfigCmd(c, target, id, site, ctype, func(config concern.IConfig) bool {
		*config.GetConcernFilter() = concern.ConcernFilterConfig{}
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

func IConfigFilterCmdShow(c *MessageContext, target mt.Target, id string, site string, ctype concern_type.Type) {
	err := iConfigCmd(c, target, id, site, ctype, func(config concern.IConfig) bool {
		if config.GetConcernFilter().Empty() {
			c.TextReply("当前配置为空")
			return false
		}
		sb := strings.Builder{}
		sb.WriteString("当前配置：\n")
		switch config.GetConcernFilter().Type {
		case concern.FilterTypeText:
			sb.WriteString("关键字过滤模式：\n")
			filter, err := config.GetConcernFilter().GetFilterByText()
			if err != nil {
				logger.WithField("filter_config", config.GetConcernFilter().Config).
					Errorf("get filter failed %v", err)
				c.TextReply("查询失败 - 内部错误")
				return false
			}
			for _, kw := range filter.Text {
				sb.WriteString(kw)
				sb.WriteRune('\n')
			}
		case concern.FilterTypeType, concern.FilterTypeNotType:
			filter, err := config.GetConcernFilter().GetFilterByType()
			if err != nil {
				logger.WithField("filter_config", config.GetConcernFilter().Config).
					Errorf("get filter failed %v", err)
				c.TextReply("查询失败 - 内部错误")
				return false
			}
			if config.GetConcernFilter().Type == concern.FilterTypeType {
				sb.WriteString("动态类型过滤模式 - 只推送以下种类的动态：\n")
			} else if config.GetConcernFilter().Type == concern.FilterTypeNotType {
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

func iConfigCmd(c *MessageContext, target mt.Target, id string, site string, ctype concern_type.Type,
	f func(config concern.IConfig) bool) (err error) {
	if err = configCmdTargetCommonCheck(c, target); err != nil {
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
	err = cm.GetStateManager().CheckTargetConcern(target, mid, ctype)
	if err != concern.ErrAlreadyExists {
		return errors.New("失败 - 该id尚未watch")
	}
	cfg := cm.GetStateManager().GetConcernConfig(target, mid)
	err = cm.GetStateManager().OperateConcernConfig(target, mid, cfg, f)
	if err != nil && !localdb.IsRollback(err) {
		c.GetLog().Errorf("OperateConcernConfig failed %v", err)
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

func configCmdTargetCommonCheck(c *MessageContext, target mt.Target) error {
	if c.Lsp.PermissionStateManager.CheckTargetCommandDisabled(target, ConfigCommand) {
		c.DisabledReply()
		return permission.ErrDisabled
	}

	if !c.Lsp.PermissionStateManager.RequireAny(
		permission.AdminRoleRequireOption(c.Sender.Uin()),
		permission.TargetAdminRoleRequireOption(target, c.Sender.Uin()),
		permission.QQAdminRequireOption(target, c.Sender.Uin()),
		permission.TargetCommandRequireOption(target, c.Sender.Uin(), ConfigCommand),
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
			concernConfig.GetConcernAt().MergeAtSomeoneList(ctype, QQ)
			return true
		case "remove":
			concernConfig.GetConcernAt().RemoveAtSomeoneList(ctype, QQ)
			return true
		case "clear":
			concernConfig.GetConcernAt().ClearAtSomeoneList(ctype)
			return true
		case "show":
			qqList := concernConfig.GetConcernAt().GetAtSomeoneList(ctype)
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
		if concernConfig.GetConcernAt().CheckAtAll(ctype) {
			if on {
				// 配置@all，但已经配置了
				c.TextReply("失败 - 已经配置过了")
				return false
			} else {
				// 取消配置@all
				concernConfig.GetConcernAt().AtAll = concernConfig.GetConcernAt().AtAll.Remove(ctype)
				return true
			}
		} else {
			if !on {
				// 取消配置，但并没有配置
				c.TextReply("失败 - 该配置未设置")
				return false
			} else {
				// 配置@all
				concernConfig.GetConcernAt().AtAll = concernConfig.GetConcernAt().AtAll.Add(ctype)
				return true
			}
		}
	}
}

func operateNotifyConcernConfig(c *MessageContext, ctype concern_type.Type, on bool) func(concernConfig concern.IConfig) bool {
	return func(concernConfig concern.IConfig) bool {
		if concernConfig.GetConcernNotify().CheckTitleChangeNotify(ctype) {
			if on {
				// 配置推送，但已经配置过了
				c.TextReply("失败 - 已经配置过了")
				return false
			} else {
				// 取消配置推送
				concernConfig.GetConcernNotify().TitleChangeNotify = concernConfig.GetConcernNotify().TitleChangeNotify.Remove(ctype)
				return true
			}
		} else {
			if !on {
				// 取消配置，但并没有配置
				c.TextReply("失败 - 该配置未设置")
				return false
			} else {
				// 配置推送
				concernConfig.GetConcernNotify().TitleChangeNotify = concernConfig.GetConcernNotify().TitleChangeNotify.Add(ctype)
				return true
			}
		}
	}
}

func operateOfflineNotifyConcernConfig(c *MessageContext, ctype concern_type.Type, on bool) func(concernConfig concern.IConfig) bool {
	return func(concernConfig concern.IConfig) bool {
		if concernConfig.GetConcernNotify().CheckOfflineNotify(ctype) {
			if on {
				// 配置推送，但已经配置过了
				c.TextReply("失败 - 已经配置过了")
				return false
			} else {
				// 取消配置推送
				concernConfig.GetConcernNotify().OfflineNotify = concernConfig.GetConcernNotify().OfflineNotify.Remove(ctype)
				return true
			}
		} else {
			if !on {
				// 取消配置，但并没有配置
				c.TextReply("失败 - 该配置未设置")
				return false
			} else {
				concernConfig.GetConcernNotify().OfflineNotify = concernConfig.GetConcernNotify().OfflineNotify.Add(ctype)
				return true
			}
		}
	}
}
