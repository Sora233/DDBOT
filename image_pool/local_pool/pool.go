package local_pool

import (
	"errors"
	"io/ioutil"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/Sora233/DDBOT/v2/image_pool"
	localutils "github.com/Sora233/DDBOT/v2/utils"
	"github.com/Sora233/MiraiGo-Template/utils"
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

func (pool *LocalPool) Get(opts ...image_pool.OptionFunc) ([]image_pool.Image, error) {
	if pool == nil {
		return nil, errors.New("pool status down")
	}
	pool.freshMutex.Lock()
	defer pool.freshMutex.Unlock()

	if len(pool.imageList) == 0 {
		return nil, errors.New("no image")
	}

	var (
		result []image_pool.Image
		option = make(image_pool.Option)
		num    = 1
	)

	for _, opt := range opts {
		opt(option)
	}

	for k, v := range option {
		switch k {
		case "num":
			_v, ok := v.(int)
			if ok {
				num = _v
			}
		}
	}

	for i := 0; i < num; i++ {
		result = append(result, &Image{
			Path: pool.imageList[rand.Intn(len(pool.imageList))],
		})
	}

	return result, nil
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
