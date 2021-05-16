package lsp

import (
	"errors"
	"fmt"
	"github.com/Sora233/DDBOT/concern"
	"github.com/Sora233/DDBOT/lsp/bilibili"
	"github.com/Sora233/DDBOT/lsp/douyu"
	"github.com/Sora233/DDBOT/lsp/youtube"
	"github.com/sirupsen/logrus"
	"strconv"
	"strings"
)

var errNilParam = errors.New("内部参数错误")

type CommandContext struct {
	TextReply func(text string) interface{}
	Lsp       *Lsp
	Log       *logrus.Entry
}

func NewCommandContext() *CommandContext {
	return new(CommandContext)
}

func IWatchCommand(c *CommandContext, groupCode int64, id string, site string, watchType concern.Type, remove bool) {
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
	default:
		log.WithField("site", site).Error("unsupported")
		c.TextReply("未支持的网站")
	}
	return
}

func (ic *CommandContext) requireNotNil(param ...interface{}) error {
	for _, p := range param {
		if p == nil {
			return errNilParam
		}
	}
	return nil
}
