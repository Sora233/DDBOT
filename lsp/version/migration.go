package version

import (
	"errors"
	localdb "github.com/Sora233/DDBOT/lsp/buntdb"
	"github.com/Sora233/MiraiGo-Template/utils"
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

type simpleMigration struct {
	targetVersion int64
	f             MigrationFunc
}

func (s *simpleMigration) Func() MigrationFunc {
	return s.f
}

func (s *simpleMigration) TargetVersion() int64 {
	return s.targetVersion
}

// CreateSimpleMigration 可以用来快速创建一个 Migration 的helper
// 如果逻辑较为复杂，需要其他更多信息，也可以自行实现 Migration
func CreateSimpleMigration(targetVersion int64, f MigrationFunc) Migration {
	return &simpleMigration{
		targetVersion: targetVersion,
		f:             f,
	}
}
