package lsp

import (
	"errors"
	"fmt"
	"github.com/Logiase/MiraiGo-Template/bot"
	"github.com/Logiase/MiraiGo-Template/config"
	"github.com/Logiase/MiraiGo-Template/utils"
	"github.com/Mrs4s/MiraiGo/client"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/Sora233/Sora233-MiraiGo/lsp/aliyun"
	"github.com/Sora233/Sora233-MiraiGo/lsp/bilibili"
	"github.com/Sora233/Sora233-MiraiGo/lsp/concern"
	"github.com/asmcos/requests"
	"io/ioutil"
	"math/rand"
	"strings"
	"sync"
	"time"
)

const ModuleName = "me.sora233.Lsp"

var logger = utils.GetModuleLogger(ModuleName)

type Lsp struct {
	HImageList      []string
	BilibiliConcern *bilibili.Concern

	freshMutex    *sync.RWMutex
	concernNotify chan concern.Notify
	stop          chan interface{}
}

func (l *Lsp) MiraiGoModule() bot.ModuleInfo {
	return bot.ModuleInfo{
		ID:       ModuleName,
		Instance: instance,
	}
}

func (l *Lsp) Init() {
	aliyun.InitAliyun()
	HimageDir := config.GlobalConfig.GetString("HimageDir")
	err := l.RefreshImage(HimageDir)
	if err != nil {
		logger.Errorf("refresh image error %v", err)
	} else {
		go func() {
			mt := time.NewTicker(time.Minute)
			defer mt.Stop()
			for {
				select {
				case <-mt.C:
					go func() {
						err := l.RefreshImage(HimageDir)
						if err != nil {
							logger.Errorf("refresh image error %v", err)
						}
					}()
				case <-l.stop:
					return
				}
			}
		}()
	}
}

func (l *Lsp) PostInit() {
}

func (l *Lsp) Serve(bot *bot.Bot) {
	bot.OnGroupMessage(func(qqClient *client.QQClient, msg *message.GroupMessage) {
		if len(msg.Elements) <= 0 {
			return
		}
		cmd := NewLspGroupCommand(qqClient, msg, l)
		cmd.Execute()
	})

	bot.OnPrivateMessage(func(qqClient *client.QQClient, msg *message.PrivateMessage) {
		cmds := strings.Split(msg.ToString(), " ")
		if cmds[0] == "/ping" {
			sendingMsg := message.NewSendingMessage()
			sendingMsg.Append(message.NewText("pong"))
			qqClient.SendPrivateMessage(msg.Sender.Uin, sendingMsg)
		}
	})
}

func (l *Lsp) Start(bot *bot.Bot) {
	l.BilibiliConcern.Start()
	go l.ConcernNotify(bot.QQClient)
}

func (l *Lsp) Stop(bot *bot.Bot, wg *sync.WaitGroup) {
	defer wg.Done()
	if l.stop != nil {
		close(l.stop)
	}
	l.BilibiliConcern.Stop()
}

func (l *Lsp) RefreshImage(path string) error {
	l.freshMutex.Lock()
	defer l.freshMutex.Unlock()
	files, err := filePathWalkDir(path)
	if err != nil {
		return err
	} else {
		l.HImageList = files
	}
	return nil
}

func (l *Lsp) checkImage(img *message.ImageElement) string {
	logger.WithField("image_url", img.Url).Info("image here")
	resp, err := aliyun.Audit(img.Url)
	if err != nil {
		logger.Errorf("aliyun request error %v", err)
		return ""
	} else if resp.Data.Results[0].Code != 0 {
		logger.Errorf("aliyun response code %v, msg %v", resp.Data.Results[0].Code, resp.Data.Results[0].Message)
		return ""
	}

	logger.WithField("label", resp.Data.Results[0].SubResults[0].Label).
		WithField("rate", resp.Data.Results[0].SubResults[0].Rate).
		Debug("detect done")
	return resp.Data.Results[0].SubResults[0].Label
}

func (l *Lsp) getHImage() ([]byte, error) {
	l.freshMutex.RLock()
	defer l.freshMutex.RUnlock()
	size := len(l.HImageList)
	if size == 0 {
		return nil, errors.New("empty image list")
	}
	img := l.HImageList[rand.Intn(size)]
	logger.Debugf("choose image %v", img)
	return ioutil.ReadFile(img)
}

func (l *Lsp) ConcernNotify(qqClient *client.QQClient) {
	for {
		select {
		case inotify := <-l.concernNotify:
			switch inotify.Type() {
			case concern.BibiliLive:
				notify := (inotify).(*bilibili.ConcernLiveNotify)
				logger.WithField("GroupCode", notify.GroupCode).
					WithField("Name", notify.Name).
					WithField("Title", notify.LiveTitle).
					WithField("Status", notify.Status.String()).
					Debugf("notify")
				if notify.Status == bilibili.LiveStatus_Living {
					sendingMsg := message.NewSendingMessage()
					notifyMsg := l.NotifyMessage(qqClient, notify)
					for _, msg := range notifyMsg {
						sendingMsg.Append(msg)
					}
					qqClient.SendGroupMessage(notify.GroupCode, sendingMsg)
				}
			case concern.BilibiliNews:
				// TODO
			}
		}
	}
}

func (l *Lsp) NotifyMessage(qqClient *client.QQClient, inotify concern.Notify) []message.IMessageElement {
	var result []message.IMessageElement
	switch inotify.Type() {
	case concern.BibiliLive:
		notify := (inotify).(*bilibili.ConcernLiveNotify)
		switch notify.Status {
		case bilibili.LiveStatus_Living:
			result = append(result, message.NewText(fmt.Sprintf("%s正在直播【%s】", notify.Name, notify.LiveTitle)))
			result = append(result, message.NewText(notify.RoomUrl))
			coverResp, err := requests.Get(notify.Cover)
			if err == nil {
				if cover, err := qqClient.UploadGroupImage(notify.GroupCode, coverResp.Content()); err == nil {
					result = append(result, cover)
				}
			}
		case bilibili.LiveStatus_NoLiving:
			result = append(result, message.NewText(fmt.Sprintf("%s暂未直播", notify.Name)))
			result = append(result, message.NewText(notify.RoomUrl))
		}
	case concern.BilibiliNews:

	}

	return result
}

var instance *Lsp

func init() {
	notify := make(chan concern.Notify, 500)
	instance = &Lsp{
		freshMutex:      new(sync.RWMutex),
		concernNotify:   notify,
		stop:            make(chan interface{}),
		BilibiliConcern: bilibili.NewConcern(notify),
	}
	bot.RegisterModule(instance)
}
