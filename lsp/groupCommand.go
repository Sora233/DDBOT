package lsp

import (
	"bytes"
	"errors"
	"fmt"
	miraiBot "github.com/Logiase/MiraiGo-Template/bot"
	"github.com/Logiase/MiraiGo-Template/config"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Sora233/Sora233-MiraiGo/concern"
	"github.com/Sora233/Sora233-MiraiGo/image_pool"
	"github.com/Sora233/Sora233-MiraiGo/image_pool/lolicon_pool"
	"github.com/Sora233/Sora233-MiraiGo/lsp/aliyun"
	"github.com/Sora233/Sora233-MiraiGo/lsp/bilibili"
	localdb "github.com/Sora233/Sora233-MiraiGo/lsp/buntdb"
	"github.com/Sora233/Sora233-MiraiGo/lsp/douyu"
	"github.com/Sora233/Sora233-MiraiGo/lsp/permission"
	"github.com/Sora233/Sora233-MiraiGo/lsp/youtube"
	"github.com/Sora233/Sora233-MiraiGo/proxy_pool"
	"github.com/Sora233/Sora233-MiraiGo/utils"
	"github.com/Sora233/sliceutil"
	"github.com/alecthomas/kong"
	"github.com/tidwall/buntdb"
	"math/rand"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"
)

type LspGroupCommand struct {
	msg *message.GroupMessage

	*Runtime
}

func NewLspGroupCommand(bot *miraiBot.Bot, l *Lsp, msg *message.GroupMessage) *LspGroupCommand {
	c := &LspGroupCommand{
		Runtime: NewRuntime(bot, l),
		msg:     msg,
	}
	c.Parse(msg.Elements)
	return c
}

func (lgc *LspGroupCommand) DebugCheck() bool {
	var ok bool
	if lgc.debug {
		if sliceutil.Contains(config.GlobalConfig.GetStringSlice("debug.group"), strconv.FormatInt(lgc.groupCode(), 10)) {
			ok = true
		}
		if sliceutil.Contains(config.GlobalConfig.GetStringSlice("debug.uin"), strconv.FormatInt(lgc.msg.Sender.Uin, 10)) {
			ok = true
		}
	} else {
		ok = true
	}
	return ok
}

func (lgc *LspGroupCommand) Execute() {
	defer func() {
		if err := recover(); err != nil {
			logger.WithField("stack", string(debug.Stack())).
				Errorf("panic recovered: %v", err)
			lgc.textReply("エラー発生")
		}
	}()

	if lgc.GetCmd() != "" && !strings.HasPrefix(lgc.GetCmd(), "/") {
		return
	}

	log := logger.WithField("group_code", lgc.groupCode()).WithField("cmd", lgc.GetCmd()).WithField("args", lgc.GetArgs())
	log.Debug("execute")

	if !lgc.DebugCheck() {
		log.Debugf("debug mode, skip execute.")
		return
	}

	if lgc.GetCmd() == "" && len(lgc.GetArgs()) == 0 {
		if !lgc.groupEnabled(ImageContentCommand) {
			logger.WithField("command", ImageContentCommand).Debug("not enabled")
			return
		}
		if lgc.uin() != lgc.bot.Uin {
			lgc.ImageContent()
		}
		return
	}

	switch lgc.GetCmd() {
	case "/lsp":
		if lgc.requireNotDisable(LspCommand) {
			lgc.LspCommand()
		}
	case "/色图":
		if lgc.requireEnable(SetuCommand) {
			lgc.SetuCommand(false)
		}
	case "/黄图":
		if lgc.requireEnable(HuangtuCommand) {
			lgc.SetuCommand(true)
		}
	case "/watch":
		if lgc.requireNotDisable(WatchCommand) {
			if !lgc.requireAnyCommand(WatchCommand, UnwatchCommand) {
				lgc.noPermissionReply()
				return
			}
			lgc.WatchCommand(false)
		}
	case "/unwatch":
		if lgc.requireNotDisable(UnwatchCommand) {
			if !lgc.requireAnyCommand(WatchCommand, UnwatchCommand) {
				lgc.noPermissionReply()
				return
			}
			lgc.WatchCommand(true)
		}
	case "/list":
		if lgc.requireNotDisable(ListCommand) {
			lgc.ListCommand()
		}
	case "/签到":
		if lgc.requireNotDisable(CheckinCommand) {
			lgc.CheckinCommand()
		}
	case "/roll":
		if lgc.requireNotDisable(RollCommand) {
			lgc.RollCommand()
		}
	case "/grant":
		if !lgc.l.PermissionStateManager.RequireAny(
			permission.AdminRoleRequireOption(lgc.uin()),
			permission.GroupAdminRoleRequireOption(lgc.groupCode(), lgc.uin()),
			permission.QQAdminRequireOption(lgc.groupCode(), lgc.uin()),
		) {
			lgc.noPermissionReply()
			return
		}
		lgc.GrantCommand()
	case "/enable":
		if !lgc.l.PermissionStateManager.RequireAny(
			permission.AdminRoleRequireOption(lgc.uin()),
			permission.GroupAdminRoleRequireOption(lgc.groupCode(), lgc.uin()),
		) {
			lgc.noPermissionReply()
			return
		}
		lgc.EnableCommand(false)
	case "/disable":
		if !lgc.l.PermissionStateManager.RequireAny(
			permission.AdminRoleRequireOption(lgc.uin()),
			permission.GroupAdminRoleRequireOption(lgc.groupCode(), lgc.uin()),
		) {
			lgc.noPermissionReply()
			return
		}
		lgc.EnableCommand(true)
	case "/face":
		if lgc.requireNotDisable(FaceCommand) {
			lgc.FaceCommand()
		}
	case "/倒放":
		if lgc.requireNotDisable(ReverseCommand) {
			lgc.ReverseCommand()
		}
	case "/help":
		lgc.HelpCommand()
	default:
		log.Debug("no command matched")
	}
	return
}

func (lgc *LspGroupCommand) LspCommand() {
	msg := lgc.msg
	groupCode := msg.GroupCode

	log := logger.WithField("GroupCode", groupCode)
	log.Infof("run lsp command")
	defer log.Info("lsp command end")

	var lspCmd struct{}
	output := lgc.parseCommandSyntax(&lspCmd, LspCommand)
	if output != "" {
		lgc.textReply(output)
	}
	if lgc.exit {
		return
	}
	lgc.textReply("LSP竟然是你")
	return
}

func (lgc *LspGroupCommand) SetuCommand(r18 bool) {
	msg := lgc.msg
	bot := lgc.bot
	groupCode := msg.GroupCode

	log := logger.WithField("GroupCode", groupCode)
	log.Info("run setu command")
	defer log.Info("setu command end")

	if !lgc.l.status.ImagePoolEnable {
		log.Debug("image pool not setup")
		return
	}

	var setuCmd struct {
		Num int    `arg:"" optional:"" help:"image number"`
		Tag string `optional:"" short:"t" help:"image tag"`
	}
	var name string
	if r18 {
		name = "黄图"
	} else {
		name = "色图"
	}
	output := lgc.parseCommandSyntax(&setuCmd, name)
	if output != "" {
		lgc.textReply(output)
	}
	if lgc.exit {
		return
	}

	num := setuCmd.Num

	if num == 0 {
		num = 1
	}

	if num <= 0 || num > 10 {
		lgc.textReply("失败 - 数量范围为1-10")
		return
	}

	sendingMsg := message.NewSendingMessage()

	var options []image_pool.OptionFunc
	if r18 {
		options = append(options, lolicon_pool.R18Option(lolicon_pool.R18_ON))
	} else {
		options = append(options, lolicon_pool.R18Option(lolicon_pool.R18_OFF))
	}
	if setuCmd.Tag != "" {
		options = append(options, lolicon_pool.KeywordOption(setuCmd.Tag))
	}
	options = append(options, lolicon_pool.NumOption(num))
	imgs, err := lgc.l.GetImageFromPool(options...)
	if err != nil {
		log.Errorf("get from image pool failed %v", err)
		lgc.textReply("获取失败")
		return
	}
	if len(imgs) == 0 {
		log.Errorf("get empty image")
		lgc.textReply("获取失败")
		return
	}
	var imgsBytes = make([][]byte, len(imgs))
	var errs = make([]error, len(imgs))
	var groupImages = make([]*message.GroupImageElement, len(imgs))
	var wg sync.WaitGroup

	for index := range imgs {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			imgsBytes[index], errs[index] = imgs[index].Content()
		}(index)
	}
	wg.Wait()

	for index := range imgs {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			imgBytes, err := imgsBytes[index], errs[index]
			if err != nil {
				errs[index] = fmt.Errorf("get image bytes failed %v", err)
				return
			}
			resizedImage, err := utils.ImageNormSize(imgBytes)
			if err != nil {
				logger.Errorf("resize failed, use raw image")
				groupImages[index], errs[index] = bot.UploadGroupImage(groupCode, bytes.NewReader(imgBytes))
			} else {
				groupImages[index], errs[index] = bot.UploadGroupImage(groupCode, bytes.NewReader(resizedImage))
			}
		}(index)
	}
	wg.Wait()

	imgBatch := 2
	ok := false

	for i := 0; i < len(groupImages); i += imgBatch {
		last := i + imgBatch
		if last > len(groupImages) {
			last = len(groupImages)
		}
		groupPart := groupImages[i:last]

		for index, groupImage := range groupPart {
			if errs[i+index] != nil {
				continue
			}
			ok = true
			img := imgs[i+index]
			sendingMsg.Append(groupImage)
			if loliconImage, ok := img.(*lolicon_pool.Setu); ok {
				log.WithField("author", loliconImage.Author).
					WithField("r18", loliconImage.R18).
					WithField("pid", loliconImage.Pid).
					WithField("tags", loliconImage.Tags).
					WithField("title", loliconImage.Title).
					WithField("upload_url", groupImage.Url).
					Debug("debug image")
				sendingMsg.Append(utils.MessageTextf("标题：%v\n", loliconImage.Title))
				sendingMsg.Append(utils.MessageTextf("作者：%v\n", loliconImage.Author))
				sendingMsg.Append(utils.MessageTextf("PID：%v\n", loliconImage.Pid))
				tagCount := len(loliconImage.Tags)
				if tagCount >= 2 {
					tagCount = 2
				}
				sendingMsg.Append(utils.MessageTextf("TAG：%v\n", strings.Join(loliconImage.Tags[:tagCount], " ")))
				sendingMsg.Append(utils.MessageTextf("R18：%v", loliconImage.R18))
			}
		}
		lgc.reply(sendingMsg)
		sendingMsg = message.NewSendingMessage()
	}

	if !ok {
		lgc.textReply("获取失败")
	}
	return
}

func (lgc *LspGroupCommand) WatchCommand(remove bool) {
	var (
		msg       = lgc.msg
		groupCode = msg.GroupCode
		site      = bilibili.Site
		watchType = concern.BibiliLive
		err       error
	)

	log := logger.WithField("GroupCode", groupCode)
	log.Info("run watch command")
	defer log.Info("watch command end")

	var watchCmd struct {
		Site string `optional:"" short:"s" default:"bilibili" help:"bilibili / douyu / youtube"`
		Type string `optional:"" short:"t" default:"live" help:"news / live"`
		Id   string `arg:""`
	}
	var name string
	if remove {
		name = "unwatch"
	} else {
		name = "watch"
	}
	output := lgc.parseCommandSyntax(&watchCmd, name)
	if output != "" {
		lgc.textReply(output)
	}
	if lgc.exit {
		return
	}

	site, watchType, err = lgc.parseRawSiteAndType(watchCmd.Site, watchCmd.Type)
	if err != nil {
		log.WithField("args", lgc.GetArgs()).Errorf("parse raw concern failed %v", err)
		lgc.textReply(fmt.Sprintf("参数错误 - %v", err))
		return
	}
	log = log.WithField("site", site).WithField("type", watchType)

	id := watchCmd.Id

	switch site {
	case bilibili.Site:
		mid, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			log.WithField("id", id).Errorf("not a int")
			lgc.textReply("失败 - bilibili mid格式错误")
			return
		}
		if remove {
			// unwatch
			if err := lgc.l.bilibiliConcern.Remove(groupCode, mid, watchType); err != nil {
				lgc.textReply(fmt.Sprintf("unwatch失败 - %v", err))
			} else {
				log.WithField("mid", mid).Debugf("unwatch success")
				lgc.textReply("unwatch成功")
			}
			return
		}
		// watch
		userInfo, err := lgc.l.bilibiliConcern.Add(groupCode, mid, watchType)
		if err != nil {
			log.WithField("mid", mid).Errorf("watch error %v", err)
			lgc.textReply(fmt.Sprintf("watch失败 - %v", err))
			return
		}
		log.WithField("mid", mid).Debugf("watch success")
		lgc.textReply(fmt.Sprintf("watch成功 - Bilibili用户 %v", userInfo.Name))
	case douyu.Site:
		mid, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			log.WithField("id", id).Errorf("not a int")
			lgc.textReply("失败 - douyu id格式错误")
			return
		}
		if remove {
			// unwatch
			if err := lgc.l.douyuConcern.Remove(groupCode, mid, watchType); err != nil {
				lgc.textReply(fmt.Sprintf("unwatch失败 - %v", err))
			} else {
				log.WithField("mid", mid).Debugf("unwatch success")
				lgc.textReply("unwatch成功")
			}
			return
		}
		// watch
		userInfo, err := lgc.l.douyuConcern.Add(groupCode, mid, watchType)
		if err != nil {
			log.WithField("mid", mid).Errorf("watch error %v", err)
			lgc.textReply(fmt.Sprintf("watch失败 - %v", err))
			break
		}
		log.WithField("mid", mid).Debugf("watch success")
		lgc.textReply(fmt.Sprintf("watch成功 - 斗鱼用户 %v", userInfo.Nickname))
	case youtube.Site:
		if remove {
			// unwatch
			if err := lgc.l.youtubeConcern.Remove(groupCode, id, watchType); err != nil {
				lgc.textReply(fmt.Sprintf("unwatch失败 - %v", err))
			} else {
				log.WithField("id", id).Debugf("unwatch success")
				lgc.textReply("unwatch成功")
			}
			return
		}
		info, err := lgc.l.youtubeConcern.Add(groupCode, id, watchType)
		if err != nil {
			log.WithField("id", id).Errorf("watch error %v", err)
			lgc.textReply(fmt.Sprintf("watch失败 - %v", err))
			break
		}
		log.WithField("id", id).Debugf("watch success")
		if info.ChannelName == "" {
			lgc.textReply(fmt.Sprintf("watch成功 - YTB用户，该用户未发任何布直播/视频，无法获取名字"))
		} else {
			lgc.textReply(fmt.Sprintf("watch成功 - YTB用户 %v", info.ChannelName))
		}
	default:
		log.WithField("site", site).Error("unsupported")
		lgc.textReply("未支持的网站")
	}
}

func (lgc *LspGroupCommand) ListCommand() {
	msg := lgc.msg
	groupCode := msg.GroupCode

	log := logger.WithField("GroupCode", groupCode)
	log.Info("run list living command")
	defer log.Info("list living command end")

	var listLivingCmd struct {
		Site string `optional:"" short:"s" default:"bilibili" help:"bilibili / douyu / youtube"`
		Type string `optional:"" short:"t" default:"live" help:"news / live"`
		All  bool   `optional:"" short:"a" default:"false" help:"show all"`
	}
	output := lgc.parseCommandSyntax(&listLivingCmd, ListCommand)
	if output != "" {
		lgc.textReply(output)
	}
	if lgc.exit {
		return
	}

	site, ctype, err := lgc.parseRawSiteAndType(listLivingCmd.Site, listLivingCmd.Type)
	if err != nil {
		log.WithField("args", lgc.GetArgs()).Errorf("parse raw site failed %v", err)
		lgc.textReply(fmt.Sprintf("失败 - %v", err))
		return
	}
	log = log.WithField("site", site).WithField("type", ctype)

	all := listLivingCmd.All

	listMsg := message.NewSendingMessage()

	switch ctype {
	case concern.BibiliLive:
		listMsg.Append(message.NewText("当前直播：\n"))
		living, err := lgc.l.bilibiliConcern.ListLiving(groupCode, all)
		if err != nil {
			log.Debugf("list living failed %v", err)
			lgc.textReply(fmt.Sprintf("list living 失败 - %v", err))
			return
		}
		if living == nil {
			lgc.textReply("关注列表为空，可以使用/watch命令关注")
			return
		}
		for idx, liveInfo := range living {
			if idx != 0 {
				listMsg.Append(message.NewText("\n"))
			}
			notifyMsg := lgc.l.NotifyMessage(lgc.bot, liveInfo)
			for _, msg := range notifyMsg {
				listMsg.Append(msg)
			}
		}
		if len(listMsg.Elements) == 1 {
			listMsg.Append(message.NewText("无人直播"))
		}
	case concern.BilibiliNews:
		listMsg.Append(message.NewText("当前关注：\n"))
		news, err := lgc.l.bilibiliConcern.ListNews(groupCode, all)
		if err != nil {
			log.Debugf("list news failed %v", err)
			lgc.textReply(fmt.Sprintf("list news 失败 - %v", err))
			return
		}
		if news == nil {
			lgc.textReply("关注列表为空，可以使用/watch命令关注")
			return
		}
		for idx, newsInfo := range news {
			if idx != 0 {
				listMsg.Append(message.NewText("\n"))
			}
			listMsg.Append(utils.MessageTextf("%v %v", newsInfo.Name, newsInfo.Mid))
		}
	case concern.DouyuLive:
		listMsg.Append(message.NewText("当前直播：\n"))
		living, err := lgc.l.douyuConcern.ListLiving(groupCode, all)
		if err != nil {
			log.Debugf("list living failed %v", err)
			lgc.textReply(fmt.Sprintf("失败 - %v", err))
			return
		}
		if living == nil {
			lgc.textReply("关注列表为空，可以使用/watch命令关注")
			return
		}
		for idx, liveInfo := range living {
			if idx != 0 {
				listMsg.Append(message.NewText("\n"))
			}
			notifyMsg := lgc.l.NotifyMessage(lgc.bot, liveInfo)
			for _, msg := range notifyMsg {
				listMsg.Append(msg)
			}
		}
		if len(listMsg.Elements) == 1 {
			listMsg.Append(message.NewText("无人直播"))
		}
	case concern.YoutubeLive:
		listMsg.Append(message.NewText("当前关注：\n"))
		living, err := lgc.l.youtubeConcern.ListLiving(groupCode, all)
		if err != nil {
			log.Debugf("list living failed %v", err)
			lgc.textReply(fmt.Sprintf("失败 - %v", err))
			return
		}
		if living == nil {
			lgc.textReply("关注列表为空，可以使用/watch命令关注")
			return
		}
		for idx, info := range living {
			if idx != 0 {
				listMsg.Append(message.NewText("\n"))
			}
			listMsg.Append(utils.MessageTextf("%v %v", info.ChannelName, info.ChannelId))
		}
	case concern.YoutubeVideo:
		// TODO
		log.Error("list youtube video not supported yet")
		listMsg.Append(message.NewText("暂不支持"))
	}

	lgc.send(listMsg)
	//lgc.privateAnswer(listMsg)
	//lgc.textReply("该命令较为刷屏，已通过私聊发送")

}

func (lgc *LspGroupCommand) RollCommand() {
	msg := lgc.msg
	groupCode := msg.GroupCode

	log := logger.WithField("GroupCode", groupCode)
	log.Info("run roll command")
	defer log.Info("roll command end")

	var rollCmd struct {
		RangeArg string `arg:"" optional:"" help:"roll range, eg. 100 / 50-100"`
	}
	output := lgc.parseCommandSyntax(&rollCmd, RollCommand)
	if output != "" {
		lgc.textReply(output)
	}
	if lgc.exit {
		return
	}

	var (
		max int64 = 100
		min int64 = 1
		err error
	)

	rollarg := rollCmd.RangeArg
	if rollarg != "" {
		if strings.Contains(rollarg, "-") {
			rolls := strings.Split(rollarg, "-")
			if len(rolls) != 2 {
				lgc.textReply(fmt.Sprintf("参数解析错误 - %v", rollarg))
				return
			}
			min, err = strconv.ParseInt(rolls[0], 10, 64)
			if err != nil {
				lgc.textReply(fmt.Sprintf("参数解析错误 - %v", rollarg))
				return
			}
			max, err = strconv.ParseInt(rolls[1], 10, 64)
			if err != nil {
				lgc.textReply(fmt.Sprintf("参数解析错误 - %v", rollarg))
				return
			}
		} else {
			max, err = strconv.ParseInt(rollarg, 10, 64)
			if err != nil {
				lgc.textReply(fmt.Sprintf("参数解析错误 - %v", rollarg))
				return
			}
		}
	}
	if min > max {
		lgc.textReply(fmt.Sprintf("参数解析错误 - %v", rollarg))
		return
	}
	result := rand.Int63n(max-min+1) + min
	log = log.WithField("roll", result)
	lgc.textReply(strconv.FormatInt(result, 10))
}

func (lgc *LspGroupCommand) CheckinCommand() {
	msg := lgc.msg
	groupCode := msg.GroupCode

	log := logger.WithField("GroupCode", groupCode)
	log.Infof("run checkin command")
	defer log.Info("checkin command end")

	var checkinCmd struct{}
	output := lgc.parseCommandSyntax(&checkinCmd, CheckinCommand)
	if output != "" {
		lgc.textReply(output)
	}
	if lgc.exit {
		return
	}

	db, err := localdb.GetClient()
	if err != nil {
		logger.Errorf("get db failed %v", err)
		return
	}
	date := time.Now().Format("20060102")

	var replyText string

	err = db.Update(func(tx *buntdb.Tx) error {
		var score int64
		key := localdb.Key("Score", groupCode, msg.Sender.Uin)
		dateMarker := localdb.Key("ScoreDate", groupCode, msg.Sender.Uin, date)

		val, err := tx.Get(key)
		if err == buntdb.ErrNotFound {
			score = 0
		} else {
			score, err = strconv.ParseInt(val, 10, 64)
			if err != nil {
				log.WithField("value", val).Errorf("parse score failed %v", err)
				return err
			}
		}
		_, err = tx.Get(dateMarker)
		if err != buntdb.ErrNotFound {
			replyText = fmt.Sprintf("明天再来吧，当前积分为%v", score)
			return nil
		}

		score += 1
		_, _, err = tx.Set(key, strconv.FormatInt(score, 10), nil)
		if err != nil {
			log.WithField("sender", msg.Sender.Uin).Errorf("update score failed %v", err)
			return err
		}

		_, _, err = tx.Set(dateMarker, "1", nil)
		if err != nil {
			log.WithField("sender", msg.Sender.Uin).Errorf("update score marker failed %v", err)
			return err
		}
		replyText = fmt.Sprintf("签到成功！获得1积分，当前积分为%v", score)
		return nil
	})
	lgc.textReply(replyText)
	if err != nil {
		log.Errorf("签到失败")
	}
}

func (lgc *LspGroupCommand) EnableCommand(disable bool) {
	msg := lgc.msg
	groupCode := msg.GroupCode

	log := logger.WithField("GroupCode", groupCode)
	log.Infof("run enable command")
	defer log.Info("enable command end")

	var enableCmd struct {
		Command string `arg:"" help:"command name"`
	}
	name := "enable"
	if disable {
		name = "disable"
	}
	output := lgc.parseCommandSyntax(&enableCmd, name)
	if output != "" {
		lgc.textReply(output)
	}
	if lgc.exit {
		return
	}

	log = log.WithField("command", enableCmd.Command).WithField("disable", disable)
	if !CheckOperateableCommand(enableCmd.Command) {
		log.Errorf("unknown command")
		lgc.textReply("失败 - invalid command name")
		return
	}
	var err error
	if disable {
		err = lgc.l.PermissionStateManager.DisableGroupCommand(groupCode, enableCmd.Command)
	} else {
		err = lgc.l.PermissionStateManager.EnableGroupCommand(groupCode, enableCmd.Command)
	}
	if err != nil {
		log.Errorf("err %v", err)
		if err == permission.ErrPermissionExist {
			if disable {
				lgc.textReply("失败 - 该命令已禁用")
			} else {
				lgc.textReply("失败 - 该命令已启用")
			}
		} else {
			lgc.textReply(fmt.Sprintf("失败 - %v", err))
		}
		return
	}
	lgc.textReply("成功")
}

func (lgc *LspGroupCommand) GrantCommand() {
	msg := lgc.msg
	groupCode := msg.GroupCode

	log := logger.WithField("GroupCode", groupCode)
	log.Infof("run grant command")
	defer log.Info("grant command end")

	var grantCmd struct {
		Command string `optional:"" short:"c" xor:"1" help:"command name"`
		Role    string `optional:"" short:"r" xor:"1" enum:"Admin,GroupAdmin," help:"Admin / GroupAdmin"`
		Delete  bool   `short:"d" help:"perform a ungrant instead"`
		Target  int64  `arg:""`
	}
	output := lgc.parseCommandSyntax(&grantCmd, GrantCommand)
	if output != "" {
		lgc.textReply(output)
	}
	if lgc.exit {
		return
	}

	grantFrom := msg.Sender.Uin
	grantTo := grantCmd.Target
	if grantCmd.Command == "" && grantCmd.Role == "" {
		log.Errorf("command and role both empty")
		lgc.textReply("参数错误 - 必须指定-c / -r")
		return
	}
	del := grantCmd.Delete
	log = log.WithField("grantFrom", grantFrom).WithField("grantTo", grantTo).WithField("delete", del)
	var (
		err error
	)
	if grantCmd.Command != "" {
		log = log.WithField("command", grantCmd.Command)
		if !CheckOperateableCommand(grantCmd.Command) {
			log.Errorf("unknown command")
			lgc.textReply("失败 - invalid command name")
			return
		}
		if !lgc.l.PermissionStateManager.RequireAny(
			permission.AdminRoleRequireOption(lgc.uin()),
			permission.GroupAdminRoleRequireOption(groupCode, lgc.uin()),
			permission.QQAdminRequireOption(groupCode, lgc.uin()),
		) {
			lgc.noPermissionReply()
			return
		}
		if lgc.bot.FindGroup(groupCode).FindMember(grantTo) != nil {
			if del {
				err = lgc.l.PermissionStateManager.UngrantPermission(groupCode, grantTo, grantCmd.Command)
			} else {
				err = lgc.l.PermissionStateManager.GrantPermission(groupCode, grantTo, grantCmd.Command)
			}
		} else {
			log.Errorf("can not find uin")
			err = errors.New("未找到用户")
		}
	} else if grantCmd.Role != "" {
		grantRole := permission.FromString(grantCmd.Role)
		log = log.WithField("role", grantRole.String())
		switch grantRole {
		case permission.GroupAdmin:
			if !lgc.l.PermissionStateManager.RequireAny(
				permission.AdminRoleRequireOption(lgc.uin()),
				permission.GroupAdminRoleRequireOption(groupCode, lgc.uin()),
			) {
				lgc.noPermissionReply()
				return
			}
			if lgc.bot.FindGroup(groupCode).FindMember(grantTo) != nil {
				if del {
					err = lgc.l.PermissionStateManager.UngrantGroupRole(groupCode, grantTo, grantRole)
				} else {
					err = lgc.l.PermissionStateManager.GrantGroupRole(groupCode, grantTo, grantRole)
				}
			} else {
				log.Errorf("can not find uin")
				err = errors.New("未找到用户")
			}
		case permission.Admin:
			if !lgc.l.PermissionStateManager.RequireAny(
				permission.AdminRoleRequireOption(lgc.uin()),
			) {
				lgc.noPermissionReply()
				return
			}
			if lgc.bot.FindGroup(groupCode).FindMember(grantTo) != nil {
				if del {
					err = lgc.l.PermissionStateManager.UngrantRole(grantTo, grantRole)
				} else {
					err = lgc.l.PermissionStateManager.GrantRole(grantTo, grantRole)
				}
			} else {
				log.Errorf("can not find uin")
				err = errors.New("未找到用户")
			}
		default:
			err = errors.New("invalid role")
		}
	} else {
		log.Errorf("unknown grant")
	}
	if err != nil {
		log.Errorf("grant failed %v", err)
		if err == permission.ErrPermissionExist {
			lgc.textReply("失败 - 目标已有该权限")
		} else if err == permission.ErrPermissionNotExist {
			lgc.textReply("失败 - 目标未有该权限")
		} else {
			lgc.textReply(fmt.Sprintf("失败 - %v", err))
		}
		return
	}
	log.Debug("grant success")
	lgc.textReply("成功")
}

func (lgc *LspGroupCommand) FaceCommand() {
	msg := lgc.msg
	groupCode := msg.GroupCode

	log := logger.WithField("GroupCode", groupCode)
	log.Infof("run face command")
	defer log.Info("face command end")

	output := lgc.parseCommandSyntax(&struct{}{}, FaceCommand, kong.Description("电脑使用/face [图片] 或者 回复图片消息+/face触发"))
	if output != "" {
		lgc.textReply(output)
	}
	if lgc.exit {
		return
	}

	for _, e := range msg.Elements {
		if e.Type() == message.Image {
			if ie, ok := e.(*message.ImageElement); ok {
				lgc.faceDetect(ie.Url)
				return
			} else {
				log.Errorf("cast to ImageElement failed")
				lgc.textReply("失败")
				return
			}
		} else if e.Type() == message.Reply {
			if re, ok := e.(*message.ReplyElement); ok {
				urls := lgc.l.LspStateManager.GetMessageImageUrl(groupCode, re.ReplySeq)
				if len(urls) >= 1 {
					lgc.faceDetect(urls[0])
					return
				}
			} else {
				log.Errorf("cast to ReplyElement failed")
				lgc.textReply("失败")
				return
			}
		}
	}
	log.Debug("no image found")
	lgc.textReply("参数错误 - 未找到图片")
}

func (lgc *LspGroupCommand) ReverseCommand() {
	log := logger.WithField("GroupCode", lgc.groupCode())
	log.Info("run reverse command")
	defer log.Info("reverse command end")

	output := lgc.parseCommandSyntax(&struct{}{}, ReverseCommand, kong.Description("电脑使用/倒放 [图片] 或者 回复图片消息+/倒放触发"))
	if output != "" {
		lgc.textReply(output)
	}
	if lgc.exit {
		return
	}

	for _, e := range lgc.msg.Elements {
		if e.Type() == message.Image {
			if ie, ok := e.(*message.ImageElement); ok {
				lgc.reserveGif(ie.Url)
				return
			} else {
				log.Errorf("cast to ImageElement failed")
				lgc.textReply("失败")
				return
			}
		} else if e.Type() == message.Reply {
			if re, ok := e.(*message.ReplyElement); ok {
				urls := lgc.l.LspStateManager.GetMessageImageUrl(lgc.groupCode(), re.ReplySeq)
				if len(urls) >= 1 {
					lgc.reserveGif(urls[0])
					return
				}
			} else {
				log.Errorf("cast to ReplyElement failed")
				lgc.textReply("失败")
				return
			}
		}
	}
	log.Debug("no image found")
	lgc.textReply("参数错误 - 未找到图片")
}

func (lgc *LspGroupCommand) HelpCommand() {
	log := logger.WithField("group_code", lgc.groupCode())
	log.Info("run help command")
	defer log.Info("help command end")

	output := lgc.parseCommandSyntax(&struct{}{}, HelpCommand, kong.Description("print help message"))
	if output != "" {
		lgc.textReply(output)
	}
	if lgc.exit {
		return
	}

	text := "一个多功能DD专用机器人，包括b站直播、动态推送，斗鱼直播推送，油管直播、视频推送\n" +
		"只需添加bot好友，阁下也可为自己的群添加自动推送功能\n" +
		"详细命令请添加好友后私聊发送/help\n" +
		"by Sora233\n" +
		"如果喜欢请点一个Star：https://github.com/Sora233/DDBOT"
	lgc.textReply(text)
}

func (lgc *LspGroupCommand) ImageContent() {
	msg := lgc.msg
	groupCode := msg.GroupCode
	log := logger.WithField("group_code", groupCode)

	if !lgc.l.status.AliyunEnable {
		logger.Debug("aliyun not setup")
		return
	}

	for _, e := range msg.Elements {
		if e.Type() == message.Image {
			if img, ok := e.(*message.ImageElement); ok {
				rating := lgc.l.checkImage(img)
				if rating == aliyun.SceneSexy {
					lgc.textReply("就这")
					return
				} else if rating == aliyun.ScenePorn {
					lgc.textReply("多发点")
					return
				}
			} else {
				log.Error("can not cast element to ImageElement")
			}
		}
	}
}

func (lgc *LspGroupCommand) faceDetect(url string) {
	log := logger.WithField("GroupCode", lgc.groupCode())
	log.WithField("detect_url", url).Debug("face detect")
	img, err := utils.ImageGet(url, proxy_pool.PreferNone)
	if err != nil {
		log.Errorf("get image err %v", err)
		lgc.textReply(fmt.Sprintf("获取图片失败 - %v", err))
		return
	}
	img, err = utils.OpenCvAnimeFaceDetect(img)
	if err == utils.ErrGoCvNotSetUp {
		log.Debug("gocv not setup")
		return
	}
	if err != nil {
		log.Errorf("detect image err %v", err)
		lgc.textReply(fmt.Sprintf("检测失败 - %v", err))
		return
	}
	sendingMsg := message.NewSendingMessage()
	groupImg, err := lgc.bot.UploadGroupImage(lgc.groupCode(), bytes.NewReader(img))
	if err != nil {
		log.Errorf("upload group image failed %v", err)
		lgc.textReply(fmt.Sprintf("上传失败 - %v", err))
		return
	}
	sendingMsg.Append(groupImg)
	lgc.reply(sendingMsg)
}

func (lgc *LspGroupCommand) reserveGif(url string) {
	log := logger.WithField("GroupCode", lgc.groupCode())
	log.WithField("reserve_url", url).Debug("reserve image")
	img, err := utils.ImageGet(url, proxy_pool.PreferNone)
	if err != nil {
		log.Errorf("get image err %v", err)
		lgc.textReply("获取图片失败")
		return
	}
	img, err = utils.ImageReserve(img)
	if err != nil {
		log.Errorf("reserve image err %v", err)
		lgc.textReply(fmt.Sprintf("失败 - %v", err))
		return
	}
	sendingMsg := message.NewSendingMessage()
	groupImage, err := lgc.bot.UploadGroupImage(lgc.groupCode(), bytes.NewReader(img))
	if err != nil {
		log.Errorf("upload group image failed %v", err)
		lgc.textReply(fmt.Sprintf("上传失败 - %v", err))
		return
	}
	sendingMsg.Append(groupImage)
	lgc.reply(sendingMsg)
}

func (lgc *LspGroupCommand) uin() int64 {
	return lgc.msg.Sender.Uin
}

func (lgc *LspGroupCommand) groupCode() int64 {
	return lgc.msg.GroupCode
}

func (lgc *LspGroupCommand) requireAnyCommand(commands ...string) bool {
	var ok = lgc.l.PermissionStateManager.RequireAny(
		permission.AdminRoleRequireOption(lgc.uin()),
		permission.GroupAdminRoleRequireOption(lgc.groupCode(), lgc.uin()),
		permission.QQAdminRequireOption(lgc.groupCode(), lgc.uin()),
	)
	if ok {
		return true
	}
	for _, command := range commands {
		ok = ok || lgc.l.PermissionStateManager.RequireAny(permission.GroupCommandRequireOption(lgc.groupCode(), lgc.uin(), command))
		if ok {
			return true
		}
	}
	return false
}

func (lgc *LspGroupCommand) requireEnable(command string) bool {
	if !lgc.groupEnabled(command) {
		logger.WithField("group_code", lgc.groupCode()).
			WithField("command", command).
			Debug("not enable")
		return false
	}
	return true
}

func (lgc *LspGroupCommand) requireNotDisable(command string) bool {
	if lgc.groupDisabled(command) {
		logger.WithField("group_code", lgc.groupCode()).
			WithField("command", command).
			Debug("disabled")
		return false
	}
	return true
}

// explicit defined and enabled
func (lgc *LspGroupCommand) groupEnabled(command string) bool {
	return lgc.l.PermissionStateManager.CheckGroupCommandEnabled(lgc.groupCode(), command)
}

// explicit defined and disabled
func (lgc *LspGroupCommand) groupDisabled(command string) bool {
	return lgc.l.PermissionStateManager.CheckGroupCommandDisabled(lgc.groupCode(), command)
}

// all send method should not be called from inside a rw transaction

func (lgc *LspGroupCommand) textReply(text string) *message.GroupMessage {
	sendingMsg := message.NewSendingMessage()
	sendingMsg.Append(message.NewText(text))
	return lgc.reply(sendingMsg)
}

func (lgc *LspGroupCommand) textSend(text string) *message.GroupMessage {
	sendingMsg := message.NewSendingMessage()
	sendingMsg.Append(message.NewText(text))
	return lgc.send(sendingMsg)
}

func (lgc *LspGroupCommand) reply(msg *message.SendingMessage) *message.GroupMessage {
	sendingMsg := message.NewSendingMessage()
	sendingMsg.Append(message.NewReply(lgc.msg))
	for _, e := range msg.Elements {
		sendingMsg.Append(e)
	}
	return lgc.send(sendingMsg)
}

func (lgc *LspGroupCommand) send(msg *message.SendingMessage) *message.GroupMessage {
	return lgc.l.sendGroupMessage(lgc.groupCode(), msg)
}

func (lgc *LspGroupCommand) privateSend(msg *message.SendingMessage) {
	if lgc.msg.Sender.IsFriend {
		lgc.bot.SendPrivateMessage(lgc.uin(), msg)
	} else {
		lgc.bot.SendTempMessage(lgc.groupCode(), lgc.uin(), msg)
	}
}

func (lgc *LspGroupCommand) privateTextSend(text string) {
	sendingMsg := message.NewSendingMessage()
	sendingMsg.Append(message.NewText(text))
	lgc.privateSend(sendingMsg)
}

func (lgc *LspGroupCommand) noPermissionReply() *message.GroupMessage {
	return lgc.textReply("权限不够")
}

func (lgc *LspGroupCommand) parseRawSiteAndType(rawSite string, rawType string) (string, concern.Type, error) {
	var (
		site      string
		_type     string
		found     bool
		watchType concern.Type
		err       error
	)
	rawSite = strings.Trim(rawSite, `"`)
	rawType = strings.Trim(rawType, `"`)
	site, err = lgc.parseRawSite(rawSite)
	if err != nil {
		return "", concern.Empty, err
	}
	_type, found = utils.PrefixMatch([]string{"live", "news"}, rawType)
	if !found {
		return "", concern.Empty, errors.New("can not determine type")
	}

	switch _type {
	case "live":
		if site == bilibili.Site {
			watchType = concern.BibiliLive
		} else if site == douyu.Site {
			watchType = concern.DouyuLive
		} else if site == youtube.Site {
			watchType = concern.YoutubeLive
		} else {
			return "", concern.Empty, errors.New("unknown watch type")
		}
	case "news":
		if site == bilibili.Site {
			watchType = concern.BilibiliNews
		} else if site == youtube.Site {
			watchType = concern.YoutubeVideo
		} else {
			return "", concern.Empty, errors.New("unknown watch type")
		}
	default:
		return "", concern.Empty, errors.New("unknown watch type")
	}
	return site, watchType, nil
}

func (lgc *LspGroupCommand) parseRawSite(rawSite string) (string, error) {
	var (
		found bool
		site  string
	)

	site, found = utils.PrefixMatch([]string{bilibili.Site, douyu.Site, youtube.Site}, rawSite)
	if !found {
		return "", errors.New("can not determine site")
	}
	return site, nil
}
