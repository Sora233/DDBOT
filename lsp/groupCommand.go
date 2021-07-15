package lsp

import (
	"bytes"
	"fmt"
	miraiBot "github.com/Logiase/MiraiGo-Template/bot"
	"github.com/Logiase/MiraiGo-Template/config"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Sora233/DDBOT/concern"
	"github.com/Sora233/DDBOT/image_pool"
	"github.com/Sora233/DDBOT/image_pool/lolicon_pool"
	"github.com/Sora233/DDBOT/lsp/aliyun"
	"github.com/Sora233/DDBOT/lsp/bilibili"
	localdb "github.com/Sora233/DDBOT/lsp/buntdb"
	"github.com/Sora233/DDBOT/lsp/permission"
	"github.com/Sora233/DDBOT/lsp/youtube"
	"github.com/Sora233/DDBOT/proxy_pool"
	"github.com/Sora233/DDBOT/utils"
	"github.com/Sora233/sliceutil"
	"github.com/alecthomas/kong"
	"github.com/sirupsen/logrus"
	"github.com/tidwall/buntdb"
	"math/rand"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
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
		if sliceutil.Contains(config.GlobalConfig.GetStringSlice("debug.uin"), strconv.FormatInt(lgc.uin(), 10)) {
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
			lgc.textReply("エラー発生：看到该信息表示BOT出了一些问题，该问题已记录")
		}
	}()

	if lgc.GetCmd() != "" && !strings.HasPrefix(lgc.GetCmd(), "/") {
		return
	}

	log := lgc.DefaultLogger().WithField("cmd", lgc.GetCmdArgs())

	if lgc.l.PermissionStateManager.CheckBlockList(lgc.uin()) {
		log.Debug("blocked")
		return
	}

	if !lgc.DebugCheck() {
		log.Debugf("debug mode, skip execute.")
		return
	}

	if lgc.GetCmd() == "" && len(lgc.GetArgs()) == 0 {
		if !lgc.groupEnabled(ImageContentCommand) {
			//log.WithField("command", ImageContentCommand).Trace("not enabled")
			return
		}
		if lgc.uin() != lgc.bot.Uin {
			lgc.ImageContent()
		}
		return
	}

	log.Debug("execute command")

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
			if lgc.l.PermissionStateManager.RequireAny(permission.AdminRoleRequireOption(lgc.uin())) {
				lgc.SetuCommand(true)
			}
		}
	case "/watch":
		if lgc.requireNotDisable(WatchCommand) {
			lgc.WatchCommand(false)
		}
	case "/unwatch":
		if lgc.requireNotDisable(WatchCommand) {
			lgc.WatchCommand(true)
		}
	case "/list":
		if lgc.requireNotDisable(ListCommand) {
			lgc.ListCommand()
		}
	case "/config":
		if lgc.requireNotDisable(ConfigCommand) {
			lgc.ConfigCommand()
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
		lgc.GrantCommand()
	case "/enable":
		lgc.EnableCommand(false)
	case "/disable":
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
		if lgc.requireNotDisable(HelpCommand) {
			lgc.HelpCommand()
		}
	default:
		log.Debug("no command matched")
	}
	return
}

func (lgc *LspGroupCommand) LspCommand() {
	log := lgc.DefaultLoggerWithCommand(LspCommand)
	log.Infof("run lsp command")
	defer func() { log.Info("lsp command end") }()

	var lspCmd struct{}
	_, output := lgc.parseCommandSyntax(&lspCmd, LspCommand)
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
	log := lgc.DefaultLoggerWithCommand(SetuCommand)
	log.Info("run setu command")
	defer func() { log.Info("setu command end") }()

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
	_, output := lgc.parseCommandSyntax(&setuCmd, name)
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

	if !lgc.l.PermissionStateManager.RequireAny(permission.AdminRoleRequireOption(lgc.uin())) {
		if num != 1 {
			lgc.textReply("失败 - 数量限制为1")
			return
		}
		if setuCmd.Tag != "" {
			lgc.textReply("失败 - tag搜索已禁用")
			return
		}
	}

	if num <= 0 || num > 10 {
		lgc.textReply("失败 - 数量范围为1-10")
		return
	}

	var options []image_pool.OptionFunc
	if r18 {
		options = append(options, lolicon_pool.R18Option(lolicon_pool.R18On))
	} else {
		options = append(options, lolicon_pool.R18Option(lolicon_pool.R18Off))
	}
	if setuCmd.Tag != "" {
		options = append(options, lolicon_pool.KeywordOption(setuCmd.Tag))
	}
	options = append(options, lolicon_pool.NumOption(num))
	imgs, err := lgc.l.GetImageFromPool(options...)
	if err != nil {
		if err == lolicon_pool.ErrNotFound {
			lgc.textReply(err.Error())
		} else if err == lolicon_pool.ErrQuotaExceed {
			lgc.textReply("达到调用限制")
		} else {
			lgc.textReply("获取失败")
		}
		log.Errorf("get from image pool failed %v", err)
		return
	}
	if len(imgs) == 0 {
		log.Errorf("get empty image")
		lgc.textReply("获取失败")
		return
	}
	searchNum := len(imgs)
	var (
		imgsBytes   = make([][]byte, len(imgs))
		errs        = make([]error, len(imgs))
		groupImages = make([]*message.GroupImageElement, len(imgs))
		wg          = new(sync.WaitGroup)
	)

	for index := range imgs {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			imgsBytes[index], errs[index] = imgs[index].Content()
		}(index)
	}
	wg.Wait()

	log.Debug("all image requested")

	for index := range imgs {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			imgBytes, err := imgsBytes[index], errs[index]
			if err != nil || len(imgBytes) == 0 {
				errs[index] = fmt.Errorf("get image bytes failed %v", err)
				return
			}
			groupImages[index], errs[index] = utils.UploadGroupImage(lgc.groupCode(), imgBytes, true)
		}(index)
	}
	wg.Wait()

	log.Debug("all image uploaded")

	imgBatch := 10

	if r18 {
		imgBatch = 5
	}

	var missCount int32 = 0

	for i := 0; i < len(groupImages); i += imgBatch {
		last := i + imgBatch
		if last > len(groupImages) {
			last = len(groupImages)
		}
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			sendingMsg := message.NewSendingMessage()
			var imgSubCount int32 = 0
			for index, groupImage := range groupImages[i:last] {
				if errs[i+index] != nil {
					log.Errorf("upload failed %v", errs[i+index])
					atomic.AddInt32(&missCount, 1)
					continue
				}
				imgSubCount += 1
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
					sendingMsg.Append(utils.MessageTextf("PID：%v P%v\n", loliconImage.Pid, loliconImage.P))
					tagCount := len(loliconImage.Tags)
					if tagCount >= 2 {
						tagCount = 2
					}
					sendingMsg.Append(utils.MessageTextf("TAG：%v\n", strings.Join(loliconImage.Tags[:tagCount], " ")))
					sendingMsg.Append(utils.MessageTextf("R18：%v", loliconImage.R18))
				}
			}
			if len(sendingMsg.Elements) == 0 {
				return
			}
			if lgc.reply(sendingMsg).Id == -1 {
				atomic.AddInt32(&missCount, imgSubCount)
			}
		}(i)
	}

	wg.Wait()

	log = log.WithField("search_num", searchNum).WithField("miss", missCount)
	if searchNum != num || missCount != 0 {
		lgc.textReplyF("本次共查询到%v张图片，有%v张图片被吞了哦", searchNum, missCount)
	}

	return
}

func (lgc *LspGroupCommand) WatchCommand(remove bool) {
	var (
		groupCode = lgc.groupCode()
		site      = bilibili.Site
		watchType = concern.BibiliLive
		err       error
	)

	log := lgc.DefaultLoggerWithCommand(WatchCommand)
	log.Info("run watch command")
	defer func() { log.Info("watch command end") }()

	var watchCmd struct {
		Site string `optional:"" short:"s" default:"bilibili" help:"bilibili / douyu / youtube / huya"`
		Type string `optional:"" short:"t" default:"live" help:"news / live"`
		Id   string `arg:""`
	}
	var name string
	if remove {
		name = "unwatch"
	} else {
		name = "watch"
	}
	_, output := lgc.parseCommandSyntax(&watchCmd, name)
	if output != "" {
		lgc.textReply(output)
	}
	if lgc.exit {
		return
	}

	site, watchType, err = lgc.ParseRawSiteAndType(watchCmd.Site, watchCmd.Type)
	if err != nil {
		log = log.WithField("args", lgc.GetArgs())
		log.Errorf("parse raw concern failed %v", err)
		lgc.textReply(fmt.Sprintf("参数错误 - %v", err))
		return
	}
	log = log.WithField("site", site).WithField("type", watchType)

	id := watchCmd.Id

	IWatch(lgc.NewMessageContext(log), groupCode, id, site, watchType, remove)
}

func (lgc *LspGroupCommand) ListCommand() {
	groupCode := lgc.groupCode()

	log := lgc.DefaultLoggerWithCommand(ListCommand)
	log.Info("run list command")
	defer func() { log.Info("list command end") }()

	var listCmd struct {
		Site string `optional:"" short:"s" help:"已弃用"`
		Type string `optional:"" short:"t" help:"已弃用"`
	}
	_, output := lgc.parseCommandSyntax(&listCmd, ListCommand)
	if output != "" {
		lgc.textReply(output)
	}
	if lgc.exit {
		return
	}

	if listCmd.Site != "" || listCmd.Type != "" {
		lgc.textReply("命令已更新，请直接输入/list即可")
		return
	}

	IList(lgc.NewMessageContext(log), groupCode)
}

func (lgc *LspGroupCommand) RollCommand() {
	log := lgc.DefaultLoggerWithCommand(RollCommand)
	log.Info("run roll command")
	defer func() { log.Info("roll command end") }()

	var rollCmd struct {
		RangeArg []string `arg:"" optional:"" help:"roll range, eg. 100 / 50-100 / opt1 opt2 opt3"`
	}
	_, output := lgc.parseCommandSyntax(&rollCmd, RollCommand)
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

	if len(rollCmd.RangeArg) <= 1 {
		var rollarg string
		if len(rollCmd.RangeArg) == 1 {
			rollarg = rollCmd.RangeArg[0]
		}
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
	} else {
		result := rollCmd.RangeArg[rand.Intn(len(rollCmd.RangeArg))]
		log = log.WithField("choice", result)
		lgc.textReply(result)
	}
}

func (lgc *LspGroupCommand) CheckinCommand() {
	log := lgc.DefaultLoggerWithCommand(CheckinCommand)
	log.Infof("run checkin command")
	defer func() { log.Info("checkin command end") }()

	var checkinCmd struct{}
	_, output := lgc.parseCommandSyntax(&checkinCmd, CheckinCommand)
	if output != "" {
		lgc.textReply(output)
	}
	if lgc.exit {
		return
	}

	date := time.Now().Format("20060102")

	var replyText string

	err := localdb.RWTxCover(func(tx *buntdb.Tx) error {
		var score int64
		key := localdb.Key("Score", lgc.groupCode(), lgc.uin())
		dateMarker := localdb.Key("ScoreDate", lgc.groupCode(), lgc.uin(), date)

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
			log.WithField("sender", lgc.uin()).Errorf("update score failed %v", err)
			return err
		}

		_, _, err = tx.Set(dateMarker, "1", localdb.ExpireOption(time.Hour*24*3))
		if err != nil {
			log.WithField("sender", lgc.uin()).Errorf("update score marker failed %v", err)
			return err
		}
		log = log.WithField("new score", score)
		replyText = fmt.Sprintf("签到成功！获得1积分，当前积分为%v", score)
		return nil
	})
	lgc.textReply(replyText)
	if err != nil {
		log.Errorf("checkin error %v", err)
	}
}

func (lgc *LspGroupCommand) EnableCommand(disable bool) {
	groupCode := lgc.groupCode()
	log := lgc.DefaultLoggerWithCommand(EnableCommand).WithField("disable", disable)
	log.Infof("run enable command")
	defer func() { log.Info("enable command end") }()

	name := "enable"
	if disable {
		name = "disable"
	}

	var enableCmd struct {
		Command string `arg:"" help:"command name"`
	}
	_, output := lgc.parseCommandSyntax(&enableCmd, name)
	if output != "" {
		lgc.textReply(output)
	}
	if lgc.exit {
		return
	}

	log = log.WithField("targetCommand", enableCmd.Command)

	IEnable(lgc.NewMessageContext(log), groupCode, enableCmd.Command, disable)
}

func (lgc *LspGroupCommand) GrantCommand() {
	log := lgc.DefaultLoggerWithCommand(GrantCommand)
	log.Infof("run grant command")
	defer func() { log.Info("grant command end") }()

	var grantCmd struct {
		Command string `optional:"" short:"c" xor:"1" help:"command name"`
		Role    string `optional:"" short:"r" xor:"1" enum:"Admin,GroupAdmin," help:"Admin / GroupAdmin"`
		Delete  bool   `short:"d" help:"perform a ungrant instead"`
		Target  int64  `arg:""`
	}
	_, output := lgc.parseCommandSyntax(&grantCmd, GrantCommand)
	if output != "" {
		lgc.textReply(output)
	}
	if lgc.exit {
		return
	}

	grantFrom := lgc.uin()
	grantTo := grantCmd.Target
	if grantCmd.Command == "" && grantCmd.Role == "" {
		log.Errorf("command and role both empty")
		lgc.textReply("参数错误 - 必须指定-c / -r")
		return
	}
	del := grantCmd.Delete
	log = log.WithField("grantFrom", grantFrom).WithField("grantTo", grantTo).WithField("delete", del)

	if grantCmd.Command != "" {
		IGrantCmd(lgc.NewMessageContext(log), lgc.groupCode(), grantCmd.Command, grantTo, del)
	} else if grantCmd.Role != "" {
		IGrantRole(lgc.NewMessageContext(log), lgc.groupCode(), permission.FromString(grantCmd.Role), grantTo, del)
	}
}

func (lgc *LspGroupCommand) ConfigCommand() {
	log := lgc.DefaultLoggerWithCommand(ConfigCommand)
	log.Infof("run config command")
	defer func() { log.Info("config command end") }()

	var configCmd struct {
		At struct {
			Site   string  `optional:"" short:"s" default:"bilibili" help:"bilibili / douyu / youtube / huya"`
			Id     string  `arg:"" help:"配置的主播id"`
			Action string  `arg:"" enum:"add,remove,clear,show" help:"add / remove / clear / show"`
			QQ     []int64 `arg:"" optional:"" help:"需要@的成员QQ号码"`
		} `cmd:"" help:"配置推送时的@人员列表，默认为空" name:"at"`
		AtAll struct {
			Site   string `optional:"" short:"s" default:"bilibili" help:"bilibili / douyu / youtube / huya"`
			Id     string `arg:"" help:"配置的主播id"`
			Switch string `arg:"" default:"on" enum:"on,off" help:"on / off"`
		} `cmd:"" help:"配置推送时@全体成员，默认关闭，需要管理员权限" name:"at_all"`
		TitleNotify struct {
			Site   string `optional:"" short:"s" default:"bilibili" help:"bilibili / douyu / youtube / huya"`
			Id     string `arg:"" help:"配置的主播id"`
			Switch string `arg:"" default:"off" enum:"on,off" help:"on / off"`
		} `cmd:"" help:"配置直播间标题发生变化时是否进行推送，默认不推送" name:"title_notify"`
		OfflineNotify struct {
			Site   string `optional:"" short:"s" default:"bilibili" help:"bilibili / douyu / youtube / huya"`
			Id     string `arg:"" help:"配置的主播id"`
			Switch string `arg:"" default:"off" enum:"on,off," help:"on / off"`
		} `cmd:"" help:"配置下播时是否进行推送，默认不推送" name:"offline_notify"`
		Filter struct {
			Site string `optional:"" short:"s" default:"bilibili" help:"bilibili"`
			Type struct {
				Id   string   `arg:"" help:"配置的主播id"`
				Type []string `arg:"" optional:"" help:"指定的种类"`
			} `cmd:"" help:"只推送指定种类的动态" name:"type" group:"filter"`
			NotType struct {
				Id   string   `arg:"" help:"配置的主播id"`
				Type []string `arg:"" optional:"" help:"指定不推送的种类"`
			} `cmd:"" help:"不推送指定种类的动态" name:"not_type" group:"filter"`
			Text struct {
				Id      string   `arg:"" help:"配置的主播id"`
				Keyword []string `arg:"" optional:"" help:"指定的关键字"`
			} `cmd:"" help:"当动态内容里出现关键字时进行推送" name:"text" group:"filter"`
			Clear struct {
				Id string `arg:"" help:"配置的主播id"`
			} `cmd:"" help:"清除过滤器" name:"clear" group:"filter"`
			Show struct {
				Id string `arg:"" help:"配置的主播id"`
			} `cmd:"" help:"查看当前过滤器" name:"show" group:"filter"`
		} `cmd:"" help:"配置动态过滤器，目前只支持b站动态" name:"filter"`
	}

	kongCtx, output := lgc.parseCommandSyntax(&configCmd, ConfigCommand, kong.Description("管理BOT的配置，目前支持配置@成员、@全体成员、开启下播推送、开启标题推送"))
	if output != "" {
		lgc.textReply(output)
	}
	if lgc.exit || len(kongCtx.Path) <= 1 {
		return
	}

	kongPath := strings.Split(kongCtx.Command(), " ")

	cmd := kongPath[0]
	log = log.WithField("sub_command", cmd)

	switch cmd {
	case "at":
		site, ctype, err := lgc.ParseRawSiteAndType(configCmd.At.Site, "live")
		if err != nil {
			log.WithField("site", configCmd.At.Site).Errorf("ParseRawSiteAndType failed %v", err)
			lgc.textSend(fmt.Sprintf("失败 - %v", err.Error()))
			return
		}
		log = log.WithField("site", site).WithField("id", configCmd.At.Id).WithField("action", configCmd.At.Action).WithField("QQ", configCmd.At.QQ)
		IConfigAtCmd(lgc.NewMessageContext(log), lgc.groupCode(), configCmd.At.Id, site, ctype, configCmd.At.Action, configCmd.At.QQ)
	case "at_all":
		site, ctype, err := lgc.ParseRawSiteAndType(configCmd.AtAll.Site, "live")
		if err != nil {
			log.WithField("site", configCmd.AtAll.Site).Errorf("ParseRawSiteAndType failed %v", err)
			lgc.textSend(fmt.Sprintf("失败 - %v", err.Error()))
			return
		}
		var on = utils.Switch2Bool(configCmd.AtAll.Switch)
		log = log.WithField("site", site).WithField("id", configCmd.AtAll.Id).WithField("on", on)
		IConfigAtAllCmd(lgc.NewMessageContext(log), lgc.groupCode(), configCmd.AtAll.Id, site, ctype, on)
	case "title_notify":
		site, ctype, err := lgc.ParseRawSiteAndType(configCmd.TitleNotify.Site, "live")
		if err != nil {
			log.WithField("site", configCmd.TitleNotify.Site).Errorf("ParseRawSiteAndType failed %v", err)
			lgc.textSend(fmt.Sprintf("失败 - %v", err.Error()))
			return
		}
		var on = utils.Switch2Bool(configCmd.TitleNotify.Switch)
		log = log.WithField("site", site).WithField("id", configCmd.TitleNotify.Id).WithField("on", on)
		IConfigTitleNotifyCmd(lgc.NewMessageContext(log), lgc.groupCode(), configCmd.TitleNotify.Id, site, ctype, on)
	case "offline_notify":
		site, ctype, err := lgc.ParseRawSiteAndType(configCmd.OfflineNotify.Site, "live")
		if err != nil {
			log.WithField("site", configCmd.OfflineNotify.Site).Errorf("ParseRawSiteAndType failed %v", err)
			lgc.textSend(fmt.Sprintf("失败 - %v", err.Error()))
			return
		}
		if site == youtube.Site {
			log.WithField("site", configCmd.OfflineNotify.Site).Errorf("not supported")
			lgc.textSend(fmt.Sprintf("失败 - %v", "暂不支持YTB"))
			return
		}
		var on = utils.Switch2Bool(configCmd.OfflineNotify.Switch)
		log = log.WithField("site", site).WithField("id", configCmd.OfflineNotify.Id).WithField("on", on)
		IConfigOfflineNotifyCmd(lgc.NewMessageContext(log), lgc.groupCode(), configCmd.OfflineNotify.Id, site, ctype, on)
	case "filter":
		filterCmd := kongPath[1]
		site, ctype, err := lgc.ParseRawSiteAndType(configCmd.Filter.Site, "news")
		if err != nil {
			log.WithField("site", configCmd.Filter.Site).Errorf("ParseRawSiteAndType failed %v", err)
			lgc.textSend(fmt.Sprintf("失败 - %v", err.Error()))
			return
		}
		switch filterCmd {
		case "type":
			IConfigFilterCmdType(lgc.NewMessageContext(log), lgc.groupCode(), configCmd.Filter.Type.Id, site, ctype, configCmd.Filter.Type.Type)
		case "not_type":
			IConfigFilterCmdNotType(lgc.NewMessageContext(log), lgc.groupCode(), configCmd.Filter.NotType.Id, site, ctype, configCmd.Filter.NotType.Type)
		case "text":
			IConfigFilterCmdText(lgc.NewMessageContext(log), lgc.groupCode(), configCmd.Filter.Text.Id, site, ctype, configCmd.Filter.Text.Keyword)
		case "clear":
			IConfigFilterCmdClear(lgc.NewMessageContext(log), lgc.groupCode(), configCmd.Filter.Clear.Id, site, ctype)
		case "show":
			IConfigFilterCmdShow(lgc.NewMessageContext(log), lgc.groupCode(), configCmd.Filter.Show.Id, site, ctype)
		default:
			log.WithField("filter_cmd", filterCmd).Errorf("unknown filter command")
			lgc.textSend("未知的filter子命令")
		}
	default:
		lgc.textSend("暂未支持，你可以催作者GKD")
	}
}

func (lgc *LspGroupCommand) FaceCommand() {
	log := lgc.DefaultLoggerWithCommand(FaceCommand)
	log.Infof("run face command")
	defer func() { log.Info("face command end") }()

	_, output := lgc.parseCommandSyntax(&struct{}{}, FaceCommand, kong.Description("电脑使用/face [图片] 或者 回复图片消息+/face触发"))
	if output != "" {
		lgc.textReply(output)
	}
	if lgc.exit {
		return
	}

	for _, e := range lgc.msg.Elements {
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
				urls := lgc.l.LspStateManager.GetMessageImageUrl(lgc.groupCode(), re.ReplySeq)
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
	log := lgc.DefaultLoggerWithCommand(ReverseCommand)
	log.Info("run reverse command")
	defer func() { log.Info("reverse command end") }()

	_, output := lgc.parseCommandSyntax(&struct{}{}, ReverseCommand, kong.Description("电脑使用/倒放 [图片] 或者 回复图片消息+/倒放触发"))
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
	log := lgc.DefaultLoggerWithCommand(HelpCommand)
	log.Info("run help command")
	defer func() { log.Info("help command end") }()

	_, output := lgc.parseCommandSyntax(&struct{}{}, HelpCommand, kong.Description("print help message"))
	if output != "" {
		lgc.textReply(output)
	}
	if lgc.exit {
		return
	}

	text := "一个多功能DD专用机器人，包括b站直播、动态推送，斗鱼直播推送，油管直播、视频推送，虎牙直播推送\n" +
		"只需添加bot好友，阁下也可为自己的群添加自动推送功能\n" +
		"详细命令请添加好友后私聊发送/help\n" +
		"如果有用法问题或者功能改进可以在此提出：https://www.bilibili.com/read/cv10602230\n" +
		"本项目为开源项目，如果喜欢请点一个Star：https://github.com/Sora233/DDBOT"
	lgc.textReply(text)
}

func (lgc *LspGroupCommand) ImageContent() {
	log := lgc.DefaultLoggerWithCommand(ImageContentCommand)

	if !lgc.l.status.AliyunEnable {
		logger.Debug("aliyun not setup")
		return
	}

	for _, e := range lgc.msg.Elements {
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

func (lgc *LspGroupCommand) DefaultLogger() *logrus.Entry {
	return logger.WithField("Name", lgc.displayName()).
		WithField("Uin", lgc.uin()).
		WithFields(utils.GroupLogFields(lgc.groupCode()))
}

func (lgc *LspGroupCommand) DefaultLoggerWithCommand(command string) *logrus.Entry {
	return lgc.DefaultLogger().WithField("Command", command)
}

func (lgc *LspGroupCommand) faceDetect(url string) {
	log := lgc.DefaultLoggerWithCommand(FaceCommand)
	log.WithField("detect_url", url).Debug("face detect")
	img, err := utils.ImageGet(url, proxy_pool.PreferMainland)
	if err != nil {
		log.Errorf("get image err %v", err)
		lgc.textReply("获取图片失败")
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
		lgc.textReply("上传失败")
		return
	}
	sendingMsg.Append(groupImg)
	lgc.reply(sendingMsg)
}

func (lgc *LspGroupCommand) reserveGif(url string) {
	log := lgc.DefaultLoggerWithCommand(ReverseCommand)
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
		lgc.textReply("上传失败")
		return
	}
	sendingMsg.Append(groupImage)
	lgc.reply(sendingMsg)
}

func (lgc *LspGroupCommand) uin() int64 {
	return lgc.msg.Sender.Uin
}

func (lgc *LspGroupCommand) displayName() string {
	return lgc.msg.Sender.DisplayName()
}

func (lgc *LspGroupCommand) groupCode() int64 {
	return lgc.msg.GroupCode
}

func (lgc *LspGroupCommand) groupName() string {
	return lgc.msg.GroupName
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
		lgc.DefaultLoggerWithCommand(command).Debug("not enable")
		return false
	}
	return true
}

func (lgc *LspGroupCommand) requireNotDisable(command string) bool {
	if lgc.groupDisabled(command) {
		lgc.DefaultLoggerWithCommand(command).Debug("disabled")
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

func (lgc *LspGroupCommand) textReplyF(format string, args ...interface{}) *message.GroupMessage {
	sendingMsg := message.NewSendingMessage()
	sendingMsg.Append(utils.MessageTextf(format, args...))
	return lgc.reply(sendingMsg)
}

func (lgc *LspGroupCommand) textSend(text string) *message.GroupMessage {
	sendingMsg := message.NewSendingMessage()
	sendingMsg.Append(message.NewText(text))
	return lgc.send(sendingMsg)
}

func (lgc *LspGroupCommand) textSendF(format string, args ...interface{}) *message.GroupMessage {
	sendingMsg := message.NewSendingMessage()
	sendingMsg.Append(utils.MessageTextf(format, args...))
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

func (lgc *LspGroupCommand) sender() *message.Sender {
	return lgc.msg.Sender
}

func (lgc *LspGroupCommand) privateSend(msg *message.SendingMessage) {
	if lgc.msg.Sender.IsFriend {
		lgc.bot.SendPrivateMessage(lgc.uin(), msg)
	} else {
		lgc.bot.SendGroupTempMessage(lgc.groupCode(), lgc.uin(), msg)
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

func (lgc *LspGroupCommand) NewMessageContext(log *logrus.Entry) *MessageContext {
	ctx := NewMessageContext()
	ctx.Source = SourceTypeGroup
	ctx.Lsp = lgc.l
	ctx.Log = log
	ctx.TextReply = func(text string) interface{} {
		return lgc.textReply(text)
	}
	ctx.Send = func(msg *message.SendingMessage) interface{} {
		return lgc.send(msg)
	}
	ctx.Reply = func(sendingMessage *message.SendingMessage) interface{} {
		return lgc.reply(sendingMessage)
	}
	ctx.NoPermissionReply = func() interface{} {
		return lgc.noPermissionReply()
	}
	ctx.DisabledReply = func() interface{} {
		ctx.Log.Debugf("disabled")
		return nil
	}
	ctx.GlobalDisabledReply = func() interface{} {
		ctx.Log.Debugf("global disabled")
		return lgc.textReply("失败 - 无法操作该命令，该命令已被管理员禁用")
		return nil
	}
	ctx.Sender = lgc.sender()
	return ctx
}
