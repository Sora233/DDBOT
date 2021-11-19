package miraigo_logging

import (
	"github.com/Mrs4s/MiraiGo/client"
	"github.com/Sora233/MiraiGo-Template/bot"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/sirupsen/logrus"
	"path"
	"sync"
	"time"
)

func init() {
	instance = &logging{}
	bot.RegisterModule(instance)
}

const moduleId = "ddbot.miraigo-logging"

type logging struct{}

var instance *logging

var logger *logrus.Entry

func (l *logging) MiraiGoModule() bot.ModuleInfo {
	return bot.ModuleInfo{
		ID:       moduleId,
		Instance: instance,
	}
}

func (l *logging) Init() {
	writer, err := rotatelogs.New(
		path.Join("miraigo-logs", "%Y-%m-%d.log"),
		rotatelogs.WithMaxAge(7*24*time.Hour),
		rotatelogs.WithRotationTime(24*time.Hour),
	)
	if err != nil {
		logrus.WithError(err).Error("unable to write logs")
		return
	}
	miraigoLog := logrus.New()
	miraigoLog.SetOutput(writer)
	miraigoLog.SetLevel(logrus.DebugLevel)
	miraigoLog.SetFormatter(&logrus.JSONFormatter{DisableHTMLEscape: false})
	logger = miraigoLog.WithField("module", moduleId)
}

func (l *logging) PostInit() {
}

func (l *logging) Serve(bot *bot.Bot) {
	bot.OnLog(func(qqClient *client.QQClient, event *client.LogEvent) {
		logger.WithField("type", event.Type).Debug(event.Message)
	})
}

func (l *logging) Start(bot *bot.Bot) {
}

func (l *logging) Stop(bot *bot.Bot, wg *sync.WaitGroup) {
	wg.Done()
}
