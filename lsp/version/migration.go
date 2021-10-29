package version

import (
	"errors"
	"github.com/Logiase/MiraiGo-Template/utils"
	localdb "github.com/Sora233/DDBOT/lsp/buntdb"
)

var logger = utils.GetModuleLogger("ddbot.version")

// MigrationFunc 定义了迁移函数，实现的时候只能在当前goroutine内操作，不可创建其他goroutine，执行时会在RW事务中
type MigrationFunc func() error

type Migration interface {
	Func() MigrationFunc
	TargetVersion() int64
}

type MigrationMap interface {
	From(v int64) Migration
}

func DoMigration(name string, m MigrationMap) error {
	if m == nil {
		return errors.New("<nil> MigrationMap")
	}
	log := logger.WithField("Name", name)
	err := localdb.RWCover(func() error {
		for {
			curV := GetCurrentVersion(name)
			if curV == -1 {
				return ErrUnknownVersion
			}
			mig := m.From(curV)
			if mig == nil {
				return nil
			}
			log.Infof(`即将更新<%v>从 %v 迁移到 %v`, name, curV, mig.TargetVersion())
			err := mig.Func()()
			if err != nil {
				return err
			}
			_, err = SetVersion(name, mig.TargetVersion())
			if err != nil {
				return err
			}
			log.Infof(`已将<%v>从 %v 迁移到 %v`, name, curV, mig.TargetVersion())
		}
	})
	return err
}
