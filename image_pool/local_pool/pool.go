package local_pool

import (
	"errors"
	"github.com/Logiase/MiraiGo-Template/utils"
	"github.com/Sora233/Sora233-MiraiGo/image_pool"
	localutils "github.com/Sora233/Sora233-MiraiGo/utils"
	"io/ioutil"
	"math/rand"
	"os"
	"sync"
	"time"
)

var logger = utils.GetModuleLogger("local_pool")

type LocalPool struct {
	freshMutex *sync.Mutex

	imageDir  string
	imageList []string
}

type Image struct {
	Path string
}

func (i *Image) Content() ([]byte, error) {
	return ioutil.ReadFile(i.Path)
}

func (pool *LocalPool) Get(option ...image_pool.OptionFunc) ([]image_pool.Image, error) {
	if pool == nil {
		return nil, errors.New("pool status down")
	}
	pool.freshMutex.Lock()
	defer pool.freshMutex.Unlock()

	if len(pool.imageList) == 0 {
		return nil, errors.New("no image")
	}

	img := &Image{
		Path: pool.imageList[rand.Intn(len(pool.imageList))],
	}
	return []image_pool.Image{img}, nil
}

func (pool *LocalPool) RefreshImage() error {
	pool.freshMutex.Lock()
	defer pool.freshMutex.Unlock()
	files, err := localutils.FilePathWalkDir(pool.imageDir)
	if err != nil {
		return err
	} else {
		pool.imageList = files
	}
	return nil
}

func NewLocalPool(path string) (*LocalPool, error) {
	if i, err := os.Stat(path); err != nil || i == nil {
		return nil, errors.New("invalid path")
	}
	pool := &LocalPool{
		imageDir:   path,
		freshMutex: new(sync.Mutex),
		imageList:  make([]string, 0),
	}
	err := pool.RefreshImage()
	if err != nil {
		return nil, err
	}
	go func() {
		timer := time.NewTimer(time.Minute)
		for {
			select {
			case <-timer.C:
				err := pool.RefreshImage()
				if err != nil {
					logger.Errorf("local pool refresh failed %v", err)
				}
				timer.Reset(time.Minute)
			}
		}
	}()
	return pool, nil
}
