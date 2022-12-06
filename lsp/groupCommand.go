package lsp

import (
	"fmt"
	"math/rand"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.uber.org/atomic"

	mirai_client "github.com/Mrs4s/MiraiGo/client"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Sora233/DDBOT/image_pool"
	"github.com/Sora233/DDBOT/image_pool/lolicon_pool"
	localdb "github.com/Sora233/DDBOT/lsp/buntdb"
	"github.com/Sora233/DDBOT/lsp/concern"
	"github.com/Sora233/DDBOT/lsp/concern_type"
	"github.com/Sora233/DDBOT/lsp/mmsg"
	"github.com/Sora233/DDBOT/lsp/permission"
	"github.com/Sora233/DDBOT/lsp/template"
	"github.com/Sora233/DDBOT/utils"
	"github.com/Sora233/MiraiGo-Template/config"
	"github.com/Sora233/sliceutil"
	"github.com/alecthomas/kong"
	"github.com/sirupsen/logrus"
)

type LspGroupCommand struct {
	msg *message.GroupMessage

	*Runtime
}

func NewLspGroupCommand(l *Lsp, msg *message.GroupMessage) *LspGroupCommand {
	c := &LspGroupCommand{
		Runtime: NewRuntime(l, l.PermissionStateManager.CheckGroupSilence(msg.GroupCode)),
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

	if len(lgc.CommandName()) == 0 {
		return
	}

	log := lgc.DefaultLogger().WithField("cmd", lgc.GetCmdArgs())

	if lgc.l.PermissionStateManager.CheckBlockList(lgc.uin()) ||
		lgc.l.PermissionStateManager.CheckBlockList(lgc.groupCode()) {
		log.Debug("blocked")
		return
	}

	if !lgc.DebugCheck() {
		log.Debugf("debug mode, skip execute.")
		return
	}

	if !lgc.AtCheck() {
		log.Debugf("at check fail, skip execute")
		return
	}

	log.Debug("execute command")

	switch lgc.CommandName() {
	case LspCommand:
		if lgc.requireNotDisable(LspCommand) {
			lgc.LspCommand()
		}
	case SetuCommand:
		if lgc.requireEnable(SetuCommand) {
			lgc.SetuCommand(false)
		}
	case HuangtuCommand:
		if lgc.requireEnable(HuangtuCommand) {
			if lgc.l.PermissionStateManager.RequireAny(
				permission.AdminRoleRequireOption(lgc.uin()),
				permission.GroupCommandRequireOption(lgc.groupCode(), lgc.uin(), HuangtuCommand),
			) {
				lgc.SetuCommand(true)
			}
		}
	case WatchCommand:
		if lgc.requireNotDisable(WatchCommand) {
			lgc.WatchCommand(false)
		}
	case UnwatchCommand:
		if lgc.requireNotDisable(WatchCommand) {
			lgc.WatchCommand(true)
		}
	case ListCommand:
		if lgc.requireNotDisable(ListCommand) {
			lgc.ListCommand()
		}
	case ConfigCommand:
		if lgc.requireNotDisable(ConfigCommand) {
			lgc.ConfigCommand()
		}
	case CheckinCommand:
		if lgc.requireNotDisable(CheckinCommand) {
			lgc.CheckinCommand()
		}
	case RollCommand:
		if lgc.requireNotDisable(RollCommand) {
			lgc.RollCommand()
		}
	case ScoreCommand:
		if lgc.requireNotDisable(ScoreCommand) {
			lgc.ScoreCommand()
		}
	case GrantCommand:
		lgc.GrantCommand()
	case EnableCommand:
		lgc.EnableCommand(false)
	case DisableCommand:
		lgc.EnableCommand(true)
	case SilenceCommand:
		lgc.SilenceCommand()
	case ReverseCommand:
		if lgc.requireNotDisable(ReverseCommand) {
			lgc.ReverseCommand()
		}
	case HelpCommand:
		if lgc.requireNotDisable(HelpCommand) {
			lgc.HelpCommand()
		}
	case CleanConcern:
		if lgc.requireNotDisable(CleanConcern) {
			if lgc.l.PermissionStateManager.RequireAny(
				permission.AdminRoleRequireOption(lgc.uin()),
				permission.GroupAdminRoleRequireOption(lgc.groupCode(), lgc.uin()),
			) {
				lgc.CleanConcernCommand()
			}
		}
	case WaifuCommand:
		if lgc.requireNotDisable(WaifuCommand) {
			lgc.WaifuCommand()
		}
	case DivinationCommand:
		if lgc.requireNotDisable(DivinationCommand) {
			lgc.DivinationCommand()
		}
	case FightCommand:
		if lgc.requireNotDisable(FightCommand) {
			lgc.FightCommand()
		}
	default:
		if CheckCustomGroupCommand(lgc.CommandName()) {
			if lgc.requireNotDisable(lgc.CommandName()) {
				func() {
					log := lgc.DefaultLoggerWithCommand(lgc.CommandName()).WithField("CustomCommand", true)
					log.Infof("run %v command", lgc.CommandName())
					defer func() { log.Infof("%v command end", lgc.CommandName()) }()
					lgc.sendChain(
						lgc.templateMsg(fmt.Sprintf("custom.command.group.%s.tmpl", lgc.CommandName()),
							map[string]interface{}{
								"cmd":        lgc.CommandName(),
								"args":       lgc.GetArgs(),
								"full_args":  strings.Join(lgc.GetArgs(), " "),
								"at_targets": lgc.GetAtArgs(),
							},
						),
					)
				}()
			}
		} else {
			log.Debug("no command matched")
		}
	}
	return
}

func (lgc *LspGroupCommand) LspCommand() {
	log := lgc.DefaultLoggerWithCommand(lgc.CommandName())
	log.Infof("run %v command", lgc.CommandName())
	defer func() { log.Infof("%v command end", lgc.CommandName()) }()

	var lspCmd struct{}
	_, output := lgc.parseCommandSyntax(&lspCmd, lgc.CommandName())
	if output != "" {
		lgc.textReply(output)
	}
	if lgc.exit {
		return
	}
	lgc.sendChain(lgc.templateMsg("command.group.lsp.tmpl", nil))
}

func (lgc *LspGroupCommand) SetuCommand(r18 bool) {
	log := lgc.DefaultLoggerWithCommand(lgc.CommandName())
	log.Infof("run %v command", lgc.CommandName())
	defer func() { log.Infof("%v command end", lgc.CommandName()) }()

	if !lgc.l.status.ImagePoolEnable {
		log.Debug("image pool not setup")
		return
	}

	var setuCmd struct {
		Num int    `arg:"" optional:"" help:"image number"`
		Tag string `optional:"" short:"t" help:"image tag"`
	}

	_, output := lgc.parseCommandSyntax(&setuCmd, lgc.CommandName())
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
	options = append(options, image_pool.NumOption(num))
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

	var missCount atomic.Int32

	for i := 0; i < len(groupImages); i += imgBatch {
		last := i + imgBatch
		if last > len(groupImages) {
			last = len(groupImages)
		}
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			msg := mmsg.NewMSG()
			var imgSubCount int32 = 0
			for index, groupImage := range groupImages[i:last] {
				if errs[i+index] != nil {
					log.Errorf("upload failed %v", errs[i+index])
					missCount.Add(1)
					continue
				}
				imgSubCount += 1
				img := imgs[i+index]
				msg.Append(groupImage)
				if loliconImage, ok := img.(*lolicon_pool.Setu); ok {
					log.WithFields(logrus.Fields{
						"Author":    loliconImage.Author,
						"R18":       loliconImage.R18,
						"Pid":       loliconImage.Pid,
						"Tags":      loliconImage.Tags,
						"Title":     loliconImage.Title,
						"UploadUrl": groupImage.Url,
					}).Debug("debug image")
					msg.Textf("标题：%v\n", loliconImage.Title)
					msg.Textf("作者：%v\n", loliconImage.Author)
					msg.Textf("PID：%v P%v\n", loliconImage.Pid, loliconImage.P)
					tagCount := len(loliconImage.Tags)
					if tagCount >= 2 {
						tagCount = 2
					}
					msg.Textf("TAG：%v\n", strings.Join(loliconImage.Tags[:tagCount], " "))
					msg.Textf("R18：%v", loliconImage.R18)
				}
			}
			if len(msg.Elements()) == 0 {
				return
			}
			if lgc.reply(msg).Id == -1 {
				missCount.Add(imgSubCount)
			}
		}(i)
	}

	wg.Wait()

	log = log.WithField("search_num", searchNum).WithField("miss", missCount)
	if searchNum != num || missCount.Load() != 0 {
		lgc.textReplyF("本次共查询到%v张图片，有%v张图片被吞了哦", searchNum, missCount.Load())
	}

	return
}

func (lgc *LspGroupCommand) WatchCommand(remove bool) {
	var (
		groupCode = lgc.groupCode()
		site      string
		watchType = concern_type.Type("live")
		err       error
	)

	log := lgc.DefaultLoggerWithCommand(lgc.CommandName())
	log.Infof("run %v command", lgc.CommandName())
	defer func() { log.Infof("%v command end", lgc.CommandName()) }()

	var watchCmd struct {
		Site string `optional:"" short:"s" default:"bilibili" help:"网站参数"`
		Type string `optional:"" short:"t" default:"" help:"类型参数"`
		Id   string `arg:""`
	}

	_, output := lgc.parseCommandSyntax(&watchCmd, lgc.CommandName(), kong.Description(
		fmt.Sprintf("当前支持的网站：%v", strings.Join(concern.ListSite(), "/"))),
	)
	if output != "" {
		lgc.textReply(output)
	}
	if lgc.exit {
		return
	}

	site, watchType, err = lgc.ParseRawSiteAndType(watchCmd.Site, watchCmd.Type)
	if err != nil {
		log = log.WithField("args", lgc.GetArgs())
		log.Errorf("ParseRawSiteAndType failed %v", err)
		lgc.textReply(fmt.Sprintf("参数错误 - %v", err))
		return
	}
	log = log.WithField("site", site).WithField("type", watchType)

	id := watchCmd.Id

	IWatch(lgc.NewMessageContext(log), groupCode, id, site, watchType, remove)
}

func (lgc *LspGroupCommand) ListCommand() {
	groupCode := lgc.groupCode()

	log := lgc.DefaultLoggerWithCommand(lgc.CommandName())
	log.Infof("run %v command", lgc.CommandName())
	defer func() { log.Infof("%v command end", lgc.CommandName()) }()

	var listCmd struct {
		Site string `optional:"" short:"s" help:"网站参数"`
	}
	_, output := lgc.parseCommandSyntax(&listCmd, lgc.CommandName())
	if output != "" {
		lgc.textReply(output)
	}
	if lgc.exit {
		return
	}

	IList(lgc.NewMessageContext(log), groupCode, listCmd.Site)
}

func (lgc *LspGroupCommand) RollCommand() {
	log := lgc.DefaultLoggerWithCommand(lgc.CommandName())
	log.Infof("run %v command", lgc.CommandName())
	defer func() { log.Infof("%v command end", lgc.CommandName()) }()

	var rollCmd struct {
		RangeArg []string `arg:"" optional:"" help:"roll range, eg. 100 / 50-100 / opt1 opt2 opt3"`
	}
	_, output := lgc.parseCommandSyntax(&rollCmd, lgc.CommandName())
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
					lgc.textReply(rollarg)
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
	log := lgc.DefaultLoggerWithCommand(lgc.CommandName())
	log.Infof("run %v command", lgc.CommandName())
	defer func() { log.Infof("%v command end", lgc.CommandName()) }()

	var checkinCmd struct{}
	_, output := lgc.parseCommandSyntax(&checkinCmd, lgc.CommandName())
	if output != "" {
		lgc.textReply(output)
	}
	if lgc.exit {
		return
	}

	date := time.Now().Format("20060102")

	var score int64
	var success bool
	err := localdb.RWCover(func() error {
		var err error
		scoreKey := localdb.Key("Score", lgc.groupCode(), lgc.uin())
		dateMarker := localdb.Key("ScoreDate", lgc.groupCode(), lgc.uin(), date)

		score, err = localdb.GetInt64(scoreKey, localdb.IgnoreNotFoundOpt())
		if err != nil {
			return err
		}
		if localdb.Exist(dateMarker) {
			log = log.WithField("current_score", score)
			success = false
			return nil
		}

		score, err = localdb.SeqNext(scoreKey)
		if err != nil {
			return err
		}

		err = localdb.Set(dateMarker, "", localdb.SetExpireOpt(time.Hour*24*3))
		if err != nil {
			return err
		}
		log = log.WithField("new_score", score)
		success = true
		return nil
	})
	if err != nil {
		lgc.textSend("失败 - 内部错误")
		log.Errorf("checkin error %v", err)
		return
	}
	lgc.sendChain(lgc.templateMsg("command.group.checkin.tmpl", map[string]interface{}{
		"score":   score,
		"success": success,
	}))
}

func (lgc *LspGroupCommand) ScoreCommand() {
	log := lgc.DefaultLoggerWithCommand(lgc.CommandName())
	log.Infof("run %v command", lgc.CommandName())
	defer func() { log.Infof("%v command end", lgc.CommandName()) }()

	var scoreCmd struct{}
	_, output := lgc.parseCommandSyntax(&scoreCmd, lgc.CommandName())
	if output != "" {
		lgc.textReply(output)
	}
	if lgc.exit {
		return
	}

	var score int64

	err := localdb.RCover(func() error {
		var err error
		key := localdb.Key("Score", lgc.groupCode(), lgc.uin())
		score, err = localdb.GetInt64(key, localdb.IgnoreNotFoundOpt())
		return err
	})
	if err != nil {
		lgc.textSend("失败 - 内部错误")
	} else {
		lgc.textReplyF("当前积分为%v", score)
	}
}

func (lgc *LspGroupCommand) EnableCommand(disable bool) {

	log := lgc.DefaultLoggerWithCommand(lgc.CommandName())
	log.Infof("run %v command", lgc.CommandName())
	defer func() { log.Infof("%v command end", lgc.CommandName()) }()

	var enableCmd struct {
		Command string `arg:"" optional:"" help:"command name"`
	}
	_, output := lgc.parseCommandSyntax(&enableCmd, lgc.CommandName())
	if output != "" {
		lgc.textReply(output)
	}
	if lgc.exit {
		return
	}

	log = log.WithField("targetCommand", enableCmd.Command)

	IEnable(lgc.NewMessageContext(log), lgc.groupCode(), enableCmd.Command, disable)
}

func (lgc *LspGroupCommand) GrantCommand() {
	log := lgc.DefaultLoggerWithCommand(lgc.CommandName())
	log.Infof("run %v command", lgc.CommandName())
	defer func() { log.Infof("%v command end", lgc.CommandName()) }()

	var grantCmd struct {
		Command string `required:"" short:"c" xor:"1" help:"命令名"`
		Role    string `required:"" short:"r" xor:"1" enum:"Admin,GroupAdmin" help:"Admin / GroupAdmin"`
		Delete  bool   `short:"d" help:"删除模式，执行删除权限操作"`
		Target  int64  `arg:"" help:"目标qq号"`
	}
	_, output := lgc.parseCommandSyntax(&grantCmd, lgc.CommandName())
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
		IGrantRole(lgc.NewMessageContext(log), lgc.groupCode(), permission.NewRoleFromString(grantCmd.Role), grantTo, del)
	}
}

func (lgc *LspGroupCommand) SilenceCommand() {
	log := lgc.DefaultLoggerWithCommand(lgc.CommandName())
	log.Infof("run %v command", lgc.CommandName())
	defer func() { log.Infof("%v command end", lgc.CommandName()) }()

	var silenceCmd struct {
		Delete bool `optional:"" short:"d" help:"取消设置"`
	}

	_, output := lgc.parseCommandSyntax(&silenceCmd, lgc.CommandName(), kong.Description("设置沉默模式"), kong.UsageOnError())
	if output != "" {
		lgc.textReply(output)
	}
	if lgc.exit {
		return
	}

	ISilenceCmd(lgc.NewMessageContext(log), lgc.groupCode(), silenceCmd.Delete)
}

func (lgc *LspGroupCommand) ConfigCommand() {
	log := lgc.DefaultLoggerWithCommand(lgc.CommandName())
	log.Infof("run %v command", lgc.CommandName())
	defer func() { log.Infof("%v command end", lgc.CommandName()) }()

	var configCmd struct {
		At struct {
			Site   string  `optional:"" short:"s" default:"bilibili" help:"网站参数"`
			Id     string  `arg:"" help:"配置的主播id"`
			Action string  `arg:"" enum:"add,remove,clear,show" help:"add / remove / clear / show"`
			QQ     []int64 `arg:"" optional:"" help:"需要@的成员QQ号码"`
		} `cmd:"" help:"配置推送时的@人员列表，默认为空" name:"at"`
		AtAll struct {
			Site   string `optional:"" short:"s" default:"bilibili" help:"网站参数"`
			Id     string `arg:"" help:"配置的主播id"`
			Switch string `arg:"" default:"on" enum:"on,off" help:"on / off"`
		} `cmd:"" help:"配置推送时@全体成员，默认关闭，需要管理员权限" name:"at_all"`
		TitleNotify struct {
			Site   string `optional:"" short:"s" default:"bilibili" help:"网站参数"`
			Id     string `arg:"" help:"配置的主播id"`
			Switch string `arg:"" default:"off" enum:"on,off" help:"on / off"`
		} `cmd:"" help:"配置直播间标题发生变化时是否进行推送，默认不推送" name:"title_notify"`
		OfflineNotify struct {
			Site   string `optional:"" short:"s" default:"bilibili" help:"网站参数"`
			Id     string `arg:"" help:"配置的主播id"`
			Switch string `arg:"" default:"off" enum:"on,off," help:"on / off"`
		} `cmd:"" help:"配置下播时是否进行推送，默认不推送" name:"offline_notify"`
		Filter struct {
			Site string `optional:"" short:"s" default:"bilibili" help:"网站参数"`
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
		} `cmd:"" help:"配置动态过滤器" name:"filter"`
	}

	kongCtx, output := lgc.parseCommandSyntax(&configCmd, lgc.CommandName(),
		kong.Description("管理BOT的配置，目前支持配置@成员、@全体成员、开启下播推送、开启标题推送、推送过滤"),
	)
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

func (lgc *LspGroupCommand) ReverseCommand() {
	log := lgc.DefaultLoggerWithCommand(lgc.CommandName())
	log.Infof("run %v command", lgc.CommandName())
	defer func() { log.Infof("%v command end", lgc.CommandName()) }()

	_, output := lgc.parseCommandSyntax(&struct{}{}, lgc.CommandName(), kong.Description("电脑使用/倒放 [图片] 或者 回复图片消息+/倒放触发"))
	if output != "" {
		lgc.textReply(output)
	}
	if lgc.exit {
		return
	}

	for _, e := range lgc.msg.Elements {
		if e.Type() == message.Image {
			switch ie := e.(type) {
			case *message.GroupImageElement:
				lgc.reserveGif(ie.Url)
				return
			default:
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
	log := lgc.DefaultLoggerWithCommand(lgc.CommandName())
	log.Infof("run %v command", lgc.CommandName())
	defer func() { log.Infof("%v command end", lgc.CommandName()) }()

	_, output := lgc.parseCommandSyntax(&struct{}{}, lgc.CommandName(), kong.Description("print help message"))
	if output != "" {
		lgc.textReply(output)
	}
	if lgc.exit {
		return
	}
	lgc.sendChain(lgc.templateMsg("command.group.help.tmpl", nil))
}

func (lgc *LspGroupCommand) CleanConcernCommand() {
	log := lgc.DefaultLoggerWithCommand(lgc.CommandName())
	log.Infof("run %v command", lgc.CommandName())
	defer func() { log.Infof("%v command end", lgc.CommandName()) }()

	var cleanConcernCmd struct {
		Site string `optional:"" short:"s" help:"清除指定的网站订阅,默认为全部"`
		Type string `optional:"" short:"t" help:"清除指定的订阅类型,默认为全部"`
	}
	_, output := lgc.parseCommandSyntax(&cleanConcernCmd, lgc.CommandName(), kong.Description("print help message"))
	if output != "" {
		lgc.textReply(output)
	}
	if lgc.exit {
		return
	}

	ICleanConcern(lgc.NewMessageContext(log),
		false, []int64{lgc.groupCode()}, cleanConcernCmd.Site, cleanConcernCmd.Type)

}

func (lgc *LspGroupCommand) WaifuCommand() {
	log := lgc.DefaultLoggerWithCommand(lgc.CommandName())
	log.Infof("run %v command", lgc.CommandName())
	defer func() { log.Infof("%v command end", lgc.CommandName()) }()

	_, output := lgc.parseCommandSyntax(&struct{}{}, lgc.CommandName(), kong.Description("/今日老婆 抽取今天的老婆！"))
	if output != "" {
		lgc.textReply(output)
	}
	if lgc.exit {
		return
	}

	date := time.Now().Format("20060102")
	waifuKey := localdb.Key("Waifu", lgc.groupCode(), lgc.uin(), date)
	log.Infof("waifuKey: %s", waifuKey)
	var waifuInfo *mirai_client.GroupMemberInfo
	var waifuExist bool
	// pre-check daily waifu existence
	err := localdb.RCover(func() error {
		var lErr error

		// check database if the user had a waifu in the group today already
		lErr = localdb.GetJson(waifuKey, &waifuInfo, localdb.IgnoreNotFoundOpt())
		if lErr != nil {
			return lErr
		}
		if localdb.Exist(waifuKey) {
			waifuExist = true
		}
		return nil
	})
	if err != nil {
		lgc.textSend("贴贴老婆失败 - 内部错误")
		log.Errorf("daily waifu precheck error: %v", err)
		return
	}

	if waifuExist {
		// if waifu exists, log the old waifu and reply
		log = log.WithField("old_waifu_uin", waifuInfo.Uin).
			WithField("old_waifu_nickname", waifuInfo.Nickname).
			WithField("old_waifu_cardname", waifuInfo.CardName)
		log.Infof("old waifu found")
		var waifu_displayname string
		if waifuInfo.CardName != "" {
			waifu_displayname = waifuInfo.CardName
		} else {
			waifu_displayname = waifuInfo.Nickname
		}
		lgc.reply(
			lgc.templateMsg(
				"command.group.waifu.tmpl",
				map[string]interface{}{
					"waifu_exist":       waifuExist,
					"waifu_uin":         waifuInfo.Uin,
					"waifu_displayname": waifu_displayname,
					// "waifu_icon_url": fmt.Sprintf("http://q2.qlogo.cn/headimg_dl?dst_uin=%d&spec=100", waifuInfo.Uin),
				},
			),
		)
		return
	}

	// no waifu yet today
	// get group member list from the group message
	groupInfo, err := (*lgc.bot.Bot).GetGroupInfo(lgc.msg.GroupCode)
	if err != nil {
		lgc.textSend("贴贴老婆失败 - 内部错误")
		log.Errorf("waifu GetGroupInfo error: %v", err)
		return
	}
	groupMembers, err := (*lgc.bot.Bot).GetGroupMembers(groupInfo)
	if err != nil {
		lgc.textSend("贴贴老婆失败 - 内部错误")
		log.Errorf("waifu GetGroupMembers error: %v", err)
		return
	}

	// get a random user from group member list
	waifuInfo = groupMembers[rand.Intn(len(groupMembers))]
	// Fixme: this is a temporary fix to avoid infinite pointer loop during json Marshal
	//		should use a dedicated structure to store waifu info rather than storing the original GroupMemberInfo
	waifuInfo.Group = nil

	log = log.WithField("new_waifu_uin", waifuInfo.Uin).
		WithField("new_waifu_nickname", waifuInfo.Nickname).
		WithField("new_waifu_cardname", waifuInfo.CardName)
	log.Infof("new waifu rolled!")

	// record daily waifu of the user in this group in database
	err = localdb.RWCover(func() error {
		var lErr error

		if localdb.Exist(waifuKey) {
			waifuExist = true
			// another goroutine selected a waifu between the waifu selection
			lErr = localdb.GetJson(waifuKey, &waifuInfo, localdb.IgnoreNotFoundOpt())
			if lErr != nil {
				return lErr
			}
			// if waifu exists, log waifu and return
			log = log.WithField("old_waifu_uin", waifuInfo.Uin).
				WithField("old_waifu_nickname", waifuInfo.Nickname).
				WithField("old_waifu_cardname", waifuInfo.CardName)
			log.Infof("old waifu found After rolled new waifu")
			return nil
		}

		lErr = localdb.SetJson(
			waifuKey,
			waifuInfo,
			localdb.SetExpireOpt(time.Hour*24*3),
		)
		if lErr != nil {
			return lErr
		}
		return nil
	})
	if err != nil {
		lgc.textSend("贴贴老婆失败 - 内部错误")
		log.Errorf("waifu insert new daily waifu error: %v", err)
		return
	}
	// reply the waifu to group
	var waifu_displayname string
	if waifuInfo.CardName != "" {
		waifu_displayname = waifuInfo.CardName
	} else {
		waifu_displayname = waifuInfo.Nickname
	}
	lgc.reply(
		lgc.templateMsg(
			"command.group.waifu.tmpl",
			map[string]interface{}{
				"waifu_exist":       waifuExist,
				"waifu_uin":         waifuInfo.Uin,
				"waifu_displayname": waifu_displayname,
				// "waifu_icon_url": fmt.Sprintf("http://q2.qlogo.cn/headimg_dl?dst_uin=%d&spec=100", waifuInfo.Uin),
			},
		),
	)
	return
}

func (lgc *LspGroupCommand) DivinationCommand() {
	log := lgc.DefaultLoggerWithCommand(lgc.CommandName())
	log.Infof("run %v command", lgc.CommandName())
	defer func() { log.Infof("%v command end", lgc.CommandName()) }()

	_, output := lgc.parseCommandSyntax(&struct{}{}, lgc.CommandName(), kong.Description("/占卜 赛博算命"))
	if output != "" {
		lgc.textReply(output)
	}
	if lgc.exit {
		return
	}

	date := time.Now().Format("20060102")
	var divineKey = localdb.Key("Divination", lgc.uin(), date)
	var divinationSN int64
	var divinationExist = false
	err := localdb.RWCover(func() error {
		var lErr error
		divinationSN, lErr = localdb.GetInt64(divineKey, localdb.IgnoreNotFoundOpt())
		if lErr != nil {
			return lErr
		}
		if localdb.Exist(divineKey) {
			divinationExist = true
			return lErr
		}

		divinationSN = int64(rand.Intn(len(utils.Divinations)))
		lErr = localdb.SetInt64(divineKey, divinationSN, localdb.SetExpireOpt(time.Hour*24*3))
		if lErr != nil {
			return lErr
		}
		return nil
	})
	if err != nil {
		lgc.textSend("占卜失败 - 内部错误")
		log.Errorf("divination error: %v", err)
		return
	}

	divination := utils.Divinations[divinationSN]

	divInscription, err := os.ReadFile(divination.InscriptionPath)
	if err != nil {
		lgc.textSend("占卜失败 - 内部错误")
		log.Errorf("divination open inscription file error: %v", err)
		return
	}
	inscription := string(divInscription)

	// reply the divination to group
	lgc.reply(
		lgc.templateMsg(
			"command.group.divination.tmpl",
			map[string]interface{}{
				"divination_exist": divinationExist,
				"divination_title": divination.Title,
				"divination_image": divination.ImagePath,
				"inscription":      inscription,
			},
		),
	)

	return
}

func (lgc *LspGroupCommand) FightCommand() {
	log := lgc.DefaultLoggerWithCommand(lgc.CommandName())
	log.Infof("run %v command", lgc.CommandName())
	defer func() { log.Infof("%v command end", lgc.CommandName()) }()

	_, output := lgc.parseCommandSyntax(
		&struct{}{}, lgc.CommandName(),
		kong.Description("at 一位群友打ta，或者留空随机打一位无辜群友"),
	)
	if output != "" {
		lgc.textReply(output)
	}
	if lgc.exit {
		return
	}

	var err error

	// find if the message at someone
	var atElement *message.AtElement
	for _, e := range lgc.msg.Elements {
		if atElement != nil {
			break
		}
		if e.Type() == message.At {
			switch ae := e.(type) {
			case *message.AtElement:
				atElement = ae
			default:
				log.Errorf("cast message element to AtElement failed")
				lgc.textReply("打人失败 - 内部错误 可能是网线过不去")
				return
			}
		}
	}

	var victimInfo *mirai_client.GroupMemberInfo

	if atElement != nil {
		log.WithField("target", atElement.Target).
			WithField("display", atElement.Display).
			WithField("subtype", atElement.SubType).
			Infof("atElement exists")

		victimInfo, err = (*lgc.bot.Bot).GetMemberInfo(lgc.msg.GroupCode, atElement.Target)
		if err != nil {
			lgc.textSend("打人失败 - 内部错误")
			log.Errorf("fight GetMemberInfo error: %v", err)
			return
		}
	} else {
		// did not at, randomly pick a victim
		var groupInfo *mirai_client.GroupInfo
		groupInfo, err = (*lgc.bot.Bot).GetGroupInfo(lgc.msg.GroupCode)
		if err != nil {
			lgc.textSend("打人失败 - 内部错误")
			log.Errorf("fight GetGroupInfo error: %v", err)
			return
		}
		var groupMembers []*mirai_client.GroupMemberInfo
		groupMembers, err = (*lgc.bot.Bot).GetGroupMembers(groupInfo)
		if err != nil {
			lgc.textSend("打人失败 - 内部错误")
			log.Errorf("fight GetGroupMembers error: %v", err)
			return
		}

		victimInfo = groupMembers[rand.Intn(len(groupMembers))]
	}

	log.
		WithField("victim_uin", victimInfo.Uin).
		WithField("victim_nickname", victimInfo.Nickname).
		WithField("victim_cardname", victimInfo.CardName).
		Infof("victim selected")

	var victimDisplayName string
	if victimInfo.CardName != "" {
		victimDisplayName = victimInfo.CardName
	} else {
		victimDisplayName = victimInfo.Nickname
	}

	lgc.reply(
		lgc.templateMsg(
			"command.group.fight.tmpl",
			map[string]interface{}{
				"victim_uin":         victimInfo.Uin,
				"victim_displayname": victimDisplayName,
			},
		),
	)

}

func (lgc *LspGroupCommand) DefaultLogger() *logrus.Entry {
	return logger.WithField("Name", lgc.displayName()).
		WithField("Uin", lgc.uin()).
		WithFields(utils.GroupLogFields(lgc.groupCode()))
}

func (lgc *LspGroupCommand) DefaultLoggerWithCommand(command string) *logrus.Entry {
	return lgc.DefaultLogger().WithField("Command", command)
}

func (lgc *LspGroupCommand) reserveGif(url string) {
	log := lgc.DefaultLoggerWithCommand(lgc.CommandName())
	log.WithField("reserve_url", url).Debug("reserve image")
	img, err := utils.ImageGet(url)
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
	lgc.reply(mmsg.NewMSG().Image(img, ""))
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

func (lgc *LspGroupCommand) textReply(text string) *message.GroupMessage {
	return lgc.reply(mmsg.NewText(text))
}

func (lgc *LspGroupCommand) textReplyF(format string, args ...interface{}) *message.GroupMessage {
	return lgc.reply(mmsg.NewTextf(format, args...))
}

func (lgc *LspGroupCommand) textSend(text string) *message.GroupMessage {
	return lgc.send(mmsg.NewText(text))
}

func (lgc *LspGroupCommand) textSendF(format string, args ...interface{}) *message.GroupMessage {
	return lgc.send(mmsg.NewTextf(format, args...))
}

func (lgc *LspGroupCommand) reply(msg *mmsg.MSG) *message.GroupMessage {
	m := mmsg.NewMSG()
	m.Append(message.NewReply(lgc.msg))
	m.Append(msg.Elements()...)
	return lgc.send(m)
}

func (lgc *LspGroupCommand) send(msg *mmsg.MSG) *message.GroupMessage {
	return lgc.l.GM(lgc.l.SendMsg(msg, mmsg.NewGroupTarget(lgc.groupCode())))[0]
}

func (lgc *LspGroupCommand) sendChain(msg *mmsg.MSG) []*message.GroupMessage {
	return lgc.l.GM(lgc.l.SendMsg(msg, mmsg.NewGroupTarget(lgc.groupCode())))
}

func (lgc *LspGroupCommand) sender() *message.Sender {
	return lgc.msg.Sender
}

func (lgc *LspGroupCommand) noPermissionReply() *message.GroupMessage {
	return lgc.textReply("权限不够")
}

func (lgc *LspGroupCommand) globalDisabledReply() *message.GroupMessage {
	return lgc.textReply("无法操作该命令，该命令已被管理员禁用")
}

func (lgc *LspGroupCommand) commonTemplateData() map[string]interface{} {
	return map[string]interface{}{
		"msg":         lgc.msg,
		"group_code":  lgc.groupCode(),
		"group_name":  lgc.groupName(),
		"member_code": lgc.sender().Uin,
		"member_name": lgc.sender().DisplayName(),
		"command":     CommandMaps,
	}
}

func (lgc *LspGroupCommand) templateMsg(name string, data map[string]interface{}) *mmsg.MSG {
	commonData := lgc.commonTemplateData()
	for k, v := range data {
		commonData[k] = v
	}
	commonData["template_name"] = name
	m, err := template.LoadAndExec(name, commonData)
	if err != nil {
		logger.Errorf("LoadAndExec error %v", err)
		lgc.textReply(fmt.Sprintf("错误 - %v", err))
		return nil
	}
	return m
}

func (lgc *LspGroupCommand) NewMessageContext(log *logrus.Entry) *MessageContext {
	ctx := NewMessageContext()
	ctx.Target = mmsg.NewGroupTarget(lgc.groupCode())
	ctx.Lsp = lgc.l
	ctx.Log = log
	ctx.SendFunc = func(m *mmsg.MSG) interface{} {
		return lgc.send(m)
	}
	ctx.ReplyFunc = func(m *mmsg.MSG) interface{} {
		return lgc.reply(m)
	}
	ctx.NoPermissionReplyFunc = func() interface{} {
		ctx.Log.Debugf("no permission")
		if !lgc.l.PermissionStateManager.CheckGroupSilence(lgc.groupCode()) {
			return lgc.noPermissionReply()
		}
		return nil
	}
	ctx.DisabledReply = func() interface{} {
		ctx.Log.Debugf("disabled")
		return nil
	}
	ctx.GlobalDisabledReply = func() interface{} {
		ctx.Log.Debugf("global disabled")
		if !lgc.l.PermissionStateManager.CheckGroupSilence(lgc.groupCode()) {
			return lgc.globalDisabledReply()
		}
		return nil
	}
	ctx.Sender = lgc.sender()
	return ctx
}
