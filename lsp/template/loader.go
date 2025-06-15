package template

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"

	"github.com/Sora233/DDBOT/v2/lsp/mmsg"
)

const templateDir = "template"

var rootT = New("root")
var mu sync.RWMutex

//go:embed default/*.tmpl
var tfs embed.FS

var once sync.Once
var watcher *fsnotify.Watcher

func initRootT() {
	once.Do(func() {
		mu.Lock()
		rootT = Must(rootT.ParseFS(tfs, "default/*.tmpl"))
		mu.Unlock()
	})
}

func InitTemplateLoader() {
	var err error
	initRootT()
	watcher, err = fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}
	if _, err := os.Stat(templateDir); err != nil {
		if os.IsNotExist(err) {
			logger.Infof("监测到<%v>目录不存在，将自动创建，请将所有自定义模板放在<%v>文件夹内", templateDir, templateDir)
			os.Mkdir(templateDir, 0766)
		} else {
			logger.Errorf("监测<%v>目录失败：%v", templateDir, err)
			return
		}
	}
	if err := watcher.Add(templateDir); err != nil {
		logger.Errorf("监测<%v>文件夹失败，自定义模板可能无法生效: %v", templateDir, err)
		return
	}
	parseExternalTemplate := func() {
		if _, err := rootT.ParseGlob(filepath.Join(templateDir, "*.tmpl")); err != nil {
			if err == ErrGlobNotMatch {
				logger.Infof(`没有发现模板文件，注意模板必须以".tmpl"结尾`)
			} else {
				logger.Errorf("解析模板错误：%v", err)
			}
		}
	}
	parseExternalTemplate()
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if !strings.HasSuffix(event.Name, "tmpl") {
					continue
				}
				logger.Infof("监测到%v文件变动: %v", templateDir, event.String())
				mu.Lock()
				rootT = Must(rootT.ParseFS(tfs, "default/*.tmpl"))
				parseExternalTemplate()
				mu.Unlock()
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				logger.Errorf("监测template发生错误: %v", err)
			}
		}
	}()
}

func Close() {
	if watcher != nil {
		watcher.Close()
	}
}

func LoadTemplate(name string) *Template {
	initRootT()
	mu.RLock()
	defer mu.RUnlock()
	return rootT.Lookup(name)
}

func LoadAndExec(name string, data interface{}) (*mmsg.MSG, error) {
	initRootT()
	m := mmsg.NewMSG()
	t := LoadTemplate(name)
	if t == nil {
		return nil, fmt.Errorf("<!missing template %v>", name)
	}
	if err := t.Execute(m, data); err != nil {
		logger.WithField("data", data).Errorf("template: %v execute error: %v", name, err)
		return nil, err
	}
	return m, nil
}
